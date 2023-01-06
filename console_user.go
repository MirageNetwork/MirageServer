package headscale

import (
	"bytes"
	_ "embed"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	eiam_developerapi20220225 "github.com/alibabacloud-go/eiam-developerapi-20220225/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/rs/zerolog/log"
)

type UserReg struct {
	Fp      string
	Name    string
	SMSCode string
}

func CreateIDaaSClient() (_result *eiam_developerapi20220225.Client, _err error) {
	// 支持匿名访问的 API，不需要 AccessKey ID 等鉴权配置
	config := &openapi.Config{}
	// 访问的域名
	config.Endpoint = tea.String("eiam-developerapi.cn-hangzhou.aliyuncs.com")
	_result = &eiam_developerapi20220225.Client{}
	_result, _err = eiam_developerapi20220225.NewClient(config)
	return _result, _err
}

//go:embed admin/addUser.html
var addUserHTML string

// 用户注册页面
func (h *Headscale) AddUserPage(
	writer http.ResponseWriter,
	req *http.Request,
) {
	addUserT := template.Must(template.New("addUser").Parse(addUserHTML))
	var payload bytes.Buffer
	if err := addUserT.Execute(&payload, nil); err != nil {
		log.Error().
			Str("handler", "addUserHTML").
			Err(err).
			Msg("Could not render addUser HTML")

		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusInternalServerError)
		_, err := writer.Write([]byte("Could not render addUserForm index template"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		return
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

// 调用阿里云IDaaS接口创建用户
func (h *Headscale) AddUserToIDaaS(
	name string, mobile string,
) (_result *eiam_developerapi20220225.CreateUserResponse, _err error) {
	generateTokenRequest := &eiam_developerapi20220225.GenerateTokenRequest{
		ClientId:     &h.cfg.ali_IDaaS.ali_cli_id,
		ClientSecret: &h.cfg.ali_IDaaS.ali_cli_key,
		GrantType:    tea.String("client_credentials"),
	}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	client, _ := CreateIDaaSClient()
	genTokenRes, _ := client.GenerateTokenWithOptions(&h.cfg.ali_IDaaS.ali_instance, &h.cfg.ali_IDaaS.ali_app_id, generateTokenRequest, headers, runtime)
	bearToken := "Bearer " + *genTokenRes.Body.AccessToken
	varTrue := true
	client, _ = CreateIDaaSClient()
	createUserHeaders := &eiam_developerapi20220225.CreateUserHeaders{
		Authorization: &bearToken,
	}
	createUserRequest := &eiam_developerapi20220225.CreateUserRequest{
		Username:                    &mobile,
		PhoneRegion:                 tea.String("86"),
		PhoneNumber:                 &mobile,
		PhoneNumberVerified:         &varTrue,
		DisplayName:                 &name,
		PrimaryOrganizationalUnitId: &h.cfg.ali_IDaaS.ali_org_id,
	}
	runtime = &util.RuntimeOptions{}
	// 复制代码运行请自行打印 API 的返回值
	createUserRes, err1 := client.CreateUserWithOptions(&h.cfg.ali_IDaaS.ali_instance, &h.cfg.ali_IDaaS.ali_app_id, createUserRequest, createUserHeaders, runtime)
	return createUserRes, err1
}

//go:embed admin/addUserVerify.html
var verifyHTML string
var verifyTemplate = template.Must(template.New("verify").Parse(verifyHTML))

type verifyTemplateConfig struct {
	FAIL  bool
	Name  string
	Phone string
}

// 用来为控制台网页生成结果页
func renderVerifyPage(
	writer http.ResponseWriter,
	FAIL bool, Name string, Phone string,
) error {
	var payload bytes.Buffer
	if err := verifyTemplate.Execute(&payload, verifyTemplateConfig{
		FAIL:  FAIL,
		Name:  Name,
		Phone: Phone,
	}); err != nil {
		log.Error().
			Str("func", "renderVerifyPage").
			Str("type", "VerifyPage").
			Err(err).
			Msg("Could not generate verify page")
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusInternalServerError)
		_, werr := writer.Write([]byte("服务器把结果页搞丢了~_~"))
		if werr != nil {
			log.Error().
				Caller().
				Err(werr).
				Msg("Failed to write response")
		}
		return err
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
	return nil
}

// 用户注册处理，提交信息和校验验证码两步均在此处处理
func (h *Headscale) AddUserAction(
	writer http.ResponseWriter,
	req *http.Request,
) {
	req.ParseForm()

	// 从前端获取姓名、手机号和指纹信息
	mobile := req.Form["mobile"][0]
	fp := req.Form["fp"][0]
	name := req.Form["name"][0]

	if req.Form["verifyCode"] == nil {
		// 发送短信验证码流程

		// 生成六位验证码
		rand.Seed(time.Now().UnixNano())
		code := rand.Intn(899999) + 100000
		verifyCode := strconv.Itoa(code)

		if _, waittime, find := h.smsCodeCache.GetWithExpiration(mobile); find {
			// 该手机号上次验证码还在有效期内
			renderResult(writer, true, "该手机号距上次获取验证码不足5分钟，"+waittime.Format("2006年01月02日 15:04:05")+"之后可以再试", "/addUser", "重新注册")
			return
		}
		if _, waittime, find := h.smsCodeCache.GetWithExpiration(fp); find {
			// 该指纹上次验证码还在有效期内
			renderResult(writer, true, "你的设备距上次获取验证码不足5分钟，"+waittime.Format("2006年01月02日 15:04:05")+"之后可以再试", "/addUser", "重新注册")
			return
		}

		newUserReg := UserReg{
			Fp:      fp,
			Name:    name,
			SMSCode: verifyCode,
		}
		//记录验证码缓存，避免大量发送
		h.smsCodeCache.Set(mobile, newUserReg, smsCacheExpiration)
		h.smsCodeCache.Set(fp, newUserReg, smsCacheExpiration)

		config := &openapi.Config{
			AccessKeyId:     &h.cfg.ali_IDaaS.ali_access_id,
			AccessKeySecret: &h.cfg.ali_IDaaS.ali_access_key,
		}
		// 访问的域名
		config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
		client := &dysmsapi20170525.Client{}
		client, _err := dysmsapi20170525.NewClient(config)
		if _err != nil {
			log.Error().Msg("发送短信客户端创建错误：" + _err.Error())
			renderResult(writer, true, "服务器发送短信验证码出错"+_err.Error()+"，请稍后再试！", "/addUser", "重新注册")
			return
		}

		sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
			PhoneNumbers:  &mobile,
			SignName:      &h.cfg.ali_IDaaS.ali_sms_sign,
			TemplateCode:  &h.cfg.ali_IDaaS.ali_sms_template,
			TemplateParam: tea.String("{\"code\":\"" + verifyCode + "\"}"),
		}
		runtime := &util.RuntimeOptions{}
		tryErr := func() (_e error) {
			defer func() {
				if r := tea.Recover(recover()); r != nil {
					_e = r
				}
			}()
			_, _err = client.SendSmsWithOptions(sendSmsRequest, runtime)
			if _err != nil {
				log.Error().Msg("发送短信验证码错误：" + _err.Error())
				renderResult(writer, true, "服务器发送短信验证码出错"+_err.Error()+"，请稍后再试！", "/addUser", "重新注册")
				return
			}
			log.Info().Msg("发送短信验证码成功！" + mobile)
			// 发送验证码成功，跳转渲染验证码输入页面
			renderVerifyPage(writer, false, name, mobile)
			return
		}()

		if tryErr != nil {
			var error = &tea.SDKError{}
			if _t, ok := tryErr.(*tea.SDKError); ok {
				error = _t
			} else {
				error.Message = tea.String(tryErr.Error())
			}
			_, _err = util.AssertAsString(error.Message)
			if _err != nil {
				log.Error().Msg("发送短信验证码错误：" + _err.Error())
				renderResult(writer, true, "服务器发送短信验证码出错"+_err.Error()+"，请稍后再试！", "/addUser", "重新注册")
				return
			}
		}
	} else {
		// 校验验证码流程
		verifyCode := req.Form["verifyCode"][0]
		log.Info().Msg("用户返回校验码为: " + verifyCode)
		if userRegInterface, ok := h.smsCodeCache.Get(mobile); ok {
			regCacheInfo := userRegInterface.(UserReg)
			if regCacheInfo.SMSCode == verifyCode && regCacheInfo.Fp == fp && regCacheInfo.Name == name {
				// 验证码校验通过，进行用户注册
				createUserRes, err := h.AddUserToIDaaS(name, mobile)
				if err != nil {
					resMsg := "创建用户失败： " + err.Error()
					renderResult(writer, true, resMsg, "/", "返回首页")
				} else {
					resMsg := "恭喜你注册成功！ 姓名：" + req.Form["name"][0] + " 手机号：" + req.Form["mobile"][0] + " 用户ID：" + *createUserRes.Body.UserId + "\n 请安装客户端使用手机号登录！"
					renderResult(writer, false, resMsg, "/", "去首页登录")
				}
			} else {
				log.Error().Msg("短信验证码校验不通过！" + mobile)
				// 验证码校验失败，返回验证码输入页面
				renderVerifyPage(writer, true, name, mobile)
				return
			}
		} else {
			log.Error().Msg("短信验证码校验错误： 验证信息缓存读取不到")
			renderResult(writer, true, "服务器校验短信验证码出错：验证信息缓存读取不到，请稍后再试！", "/addUser", "重新注册")
			return
		}

	}
}

package controller

import (
	_ "embed"
	"encoding/json"
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
	reqIP   string
	Name    string
	SMSCode string
}

func CreateIDaaSClient() (_result *eiam_developerapi20220225.Client, _err error) {
	// 支持匿名访问的 API，不需要 AccessKey ID 等鉴权配置
	config := &openapi.Config{}
	// 访问的域名
	config.Endpoint = tea.String("eiam-developerapi.cn-hangzhou.aliyuncs.com")
	//_result = &eiam_developerapi20220225.Client{}
	_result, _err = eiam_developerapi20220225.NewClient(config)
	return _result, _err
}

// 调用阿里云IDaaS接口创建用户
func (h *Mirage) AddUserToIDaaS(
	name string, mobile string,
) (_result *eiam_developerapi20220225.CreateUserResponse, _err error) {
	generateTokenRequest := &eiam_developerapi20220225.GenerateTokenRequest{
		ClientId:     &h.cfg.IDaaS.ClientID,
		ClientSecret: &h.cfg.IDaaS.ClientKey,
		GrantType:    tea.String("client_credentials"),
	}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	client, _ := CreateIDaaSClient()
	genTokenRes, _ := client.GenerateTokenWithOptions(&h.cfg.IDaaS.Instance, &h.cfg.IDaaS.App, generateTokenRequest, headers, runtime)
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
		PrimaryOrganizationalUnitId: &h.cfg.IDaaS.OrgID,
	}
	runtime = &util.RuntimeOptions{}
	// 复制代码运行请自行打印 API 的返回值
	createUserRes, err1 := client.CreateUserWithOptions(&h.cfg.IDaaS.Instance, &h.cfg.IDaaS.App, createUserRequest, createUserHeaders, runtime)
	return createUserRes, err1
}

// 用户注册处理，提交信息和校验验证码两步均在此处处理
func (h *Mirage) RegisterUserAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {

	reqData := make(map[string]string)
	json.NewDecoder(req.Body).Decode(&reqData)
	// 从前端获取姓名、手机号和指纹信息
	mobile, mobileOK := reqData["mobile"]
	name, nameOK := reqData["name"]
	if !(mobileOK && nameOK) {
		h.doAPIResponse(writer, "用户名手机号填写错误", nil)
		return
	}

	reqAddr := req.RemoteAddr

	verifyCode, codeOK := reqData["verifyCode"]

	if !codeOK {
		// 发送短信验证码流程

		// 生成六位验证码
		rand.Seed(time.Now().UnixNano())
		code := rand.Intn(899999) + 100000
		newVerifyCode := strconv.Itoa(code)

		if _, waittime, find := h.smsCodeCache.GetWithExpiration(mobile); find {
			// 该手机号上次验证码还在有效期内
			h.doAPIResponse(writer, "该手机号距上次获取验证码不足5分钟"+waittime.Format("2006年01月02日 15:04:05")+"之后可以再试", nil)
			return
		}
		if _, waittime, find := h.smsCodeCache.GetWithExpiration(reqAddr); find {
			// 该IP上次验证码还在有效期内
			h.doAPIResponse(writer, "该IP距上次获取验证码不足5分钟"+waittime.Format("2006年01月02日 15:04:05")+"之后可以再试", nil)
			return
		}

		newUserReg := UserReg{
			reqIP:   reqAddr,
			Name:    name,
			SMSCode: newVerifyCode,
		}
		//记录验证码缓存，避免大量发送
		h.smsCodeCache.Set(mobile, newUserReg, smsCacheExpiration)
		h.smsCodeCache.Set(reqAddr, newUserReg, smsCacheExpiration)
		//	h.smsCodeCache.Set(fp, newUserReg, smsCacheExpiration)

		config := &openapi.Config{
			AccessKeyId:     &h.cfg.SMS.ID,
			AccessKeySecret: &h.cfg.SMS.Key,
		}
		// 访问的域名
		config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
		client := &dysmsapi20170525.Client{}
		client, _err := dysmsapi20170525.NewClient(config)
		if _err != nil {
			log.Error().Msg("发送短信客户端创建错误：" + _err.Error())
			h.doAPIResponse(writer, "服务器发送短信验证码出错"+_err.Error()+"，请稍后再试！", nil)
			return
		}

		sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
			PhoneNumbers:  &mobile,
			SignName:      &h.cfg.SMS.Sign,
			TemplateCode:  &h.cfg.SMS.Template,
			TemplateParam: tea.String("{\"code\":\"" + newVerifyCode + "\"}"),
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
				h.doAPIResponse(writer, "服务器发送短信验证码出错"+_err.Error()+"，请稍后再试！", nil)
				return
			}
			log.Info().Msg("发送短信验证码成功！" + mobile)
			// 发送验证码成功
			h.doAPIResponse(writer, "", "TODO发送短信验证码成功该发什么")
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
				h.doAPIResponse(writer, "服务器发送短信验证码出错"+_err.Error()+"，请稍后再试！", nil)
				return
			}
		}
	} else {
		// 校验验证码流程
		log.Info().Msg("用户返回校验码为: " + verifyCode)
		if userRegInterface, ok := h.smsCodeCache.Get(mobile); ok {
			regCacheInfo := userRegInterface.(UserReg)
			if regCacheInfo.SMSCode == verifyCode /*&& regCacheInfo.reqIP == reqAddr*/ && regCacheInfo.Name == name {

				if regCacheInfo.reqIP != reqAddr {
					resMsg := "创建用户失败： IP与获取验证码时不同！"
					h.doAPIResponse(writer, resMsg, nil)
				} else {
					// 验证码校验通过，进行用户注册
					_ /*createUserRes*/, err := h.AddUserToIDaaS(name, mobile)
					if err != nil {
						resMsg := "创建用户失败： " + err.Error()
						h.doAPIResponse(writer, resMsg, nil)
					} else {
						resMsg := "恭喜你注册成功！#10 姓名：" + name + "#10 手机号：" + mobile /* + " 用户ID：" + *createUserRes.Body.UserId*/ + "#10 请安装客户端使用手机号登录接入！"
						h.doAPIResponse(writer, "", resMsg)
					}
				}
			} else {
				log.Error().Msg("短信验证码校验不通过！" + mobile)
				// 验证码校验失败，返回验证码输入页面
				h.doAPIResponse(writer, "短信验证码校验不通过！", nil)
				return
			}
		} else {
			log.Error().Msg("短信验证码校验错误： 验证信息缓存读取不到")
			h.doAPIResponse(writer, "服务器验证信息缓存读取出错", nil)
			return
		}

	}
}

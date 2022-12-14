package headscale

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	eiam_developerapi20220225 "github.com/alibabacloud-go/eiam-developerapi-20220225/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/rs/zerolog/log"
)

func CreateClient() (_result *eiam_developerapi20220225.Client, _err error) {
	// 支持匿名访问的 API，不需要 AccessKey ID 等鉴权配置
	config := &openapi.Config{}
	// 访问的域名
	config.Endpoint = tea.String("eiam-developerapi.cn-hangzhou.aliyuncs.com")
	_result = &eiam_developerapi20220225.Client{}
	_result, _err = eiam_developerapi20220225.NewClient(config)
	return _result, _err
}

//go:embed templates/addUserForm.html
var addUserForm string

// AddUserForm offer a page for individual register new accounta.
func (h *Headscale) AddUserForm(
	writer http.ResponseWriter,
	req *http.Request,
) {
	addUserTemplate := template.Must(template.New("addUser").Parse(addUserForm))

	var payload bytes.Buffer
	if err := addUserTemplate.Execute(&payload, nil); err != nil {
		log.Error().
			Str("handler", "addUserForm").
			Err(err).
			Msg("Could not render addUser template")

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

type addUserDoneTemplateConfig struct {
	Name   string
	Phone  string
	UserID string
}

var addUserDoneTemplate = template.Must(
	template.New("addUserDone").Parse(`<html>
	<body>
	<h1>创建用户成功！</h1>
	<p>
			恭喜你注册成功！ 姓名{{.Name}} 手机号：{{.Phone}} 用户ID：{{.UserID}} 。请安装客户端使用手机号登录！
	</p>
	</body>
	</html>`),
)

func (h *Headscale) AddUserAction(
	writer http.ResponseWriter,
	req *http.Request,
) {
	req.ParseForm()

	generateTokenRequest := &eiam_developerapi20220225.GenerateTokenRequest{
		ClientId:     &h.cfg.ali_IDaaS.ali_cli_id,
		ClientSecret: &h.cfg.ali_IDaaS.ali_cli_key,
		GrantType:    tea.String("client_credentials"),
	}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	client, _ := CreateClient()
	genTokenRes, _ := client.GenerateTokenWithOptions(&h.cfg.ali_IDaaS.ali_instance, &h.cfg.ali_IDaaS.ali_app_id, generateTokenRequest, headers, runtime)
	bearToken := "Bearer " + *genTokenRes.Body.AccessToken
	varTrue := true
	client, _ = CreateClient()
	createUserHeaders := &eiam_developerapi20220225.CreateUserHeaders{
		Authorization: &bearToken,
	}
	createUserRequest := &eiam_developerapi20220225.CreateUserRequest{
		Username:                    &req.Form["name"][0],
		PhoneRegion:                 tea.String("86"),
		PhoneNumber:                 &req.Form["mobile"][0],
		PhoneNumberVerified:         &varTrue,
		DisplayName:                 &req.Form["name"][0],
		PrimaryOrganizationalUnitId: &h.cfg.ali_IDaaS.ali_org_id,
	}
	runtime = &util.RuntimeOptions{}
	// 复制代码运行请自行打印 API 的返回值
	createUserRes, _ := client.CreateUserWithOptions(&h.cfg.ali_IDaaS.ali_instance, &h.cfg.ali_IDaaS.ali_app_id, createUserRequest, createUserHeaders, runtime)

	content, err := renderAddUserResult(writer, req.Form["name"][0], req.Form["mobile"][0], *createUserRes.Body.UserId)
	if err != nil {
		return
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(content.Bytes()); err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

func renderAddUserResult(
	writer http.ResponseWriter,
	name string, phone string, userID string,
) (*bytes.Buffer, error) {
	var content bytes.Buffer
	if err := addUserDoneTemplate.Execute(&content, addUserDoneTemplateConfig{
		Name:   name,
		Phone:  phone,
		UserID: userID,
	}); err != nil {
		log.Error().
			Str("func", "renderAddUserResult").
			Str("type", "addUser").
			Err(err).
			Msg("Could not Add User")

		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusInternalServerError)
		_, werr := writer.Write([]byte("Could not render Add User Result"))
		if werr != nil {
			log.Error().
				Caller().
				Err(werr).
				Msg("Failed to write response")
		}

		return nil, err
	}

	return &content, nil
}

package headscale

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

//go:embed admin/console.html
var adminHTML string

type machineItem struct {
	GiveName    string   `json:"givename"`
	UserAccount string   `json:"useraccount"`
	MIPv4       string   `json:"mipv4"`
	MIPv6       string   `json:"mipv6"`
	MSubnetList []string `json:"msubnetlist"`
	OS          string   `json:"os"`
	Version     string   `json:"version"`
	IfOnline    bool     `json:"ifonline"`
	LastSeen    string   `json:"lastseen"`

	IsSharedIn       bool `json:"issharedin"`
	IsSharedOut      bool `json:"issharedout"`
	IsExpiryDisabled bool `json:"isexpirydisabled"`
	IsExitNode       bool `json:"isexitnode"`
	IsSubnet         bool `json:"issubnet"`
}
type adminTemplateConfig struct {
	ErrorMsg     string                 `json:"errormsg"`
	Basedomain   string                 `json:"basedomain"`
	UserName     string                 `json:"username"`
	UserNameHead string                 `json:"usernamehead"`
	UserAccount  string                 `json:"useraccount"`
	MList        map[string]machineItem `json:"mlist"`
}

func (h *Headscale) ConsoleSelfAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	tokenCookie, _ := req.Cookie("OIDC_Token")
	rawToken := tokenCookie.Value
	idToken, err := h.verifyIDTokenForOIDCCallback(req.Context(), writer, rawToken)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "验证Token失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	claims, err := extractIDTokenClaims(writer, idToken)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "解析用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	namespaceName, _ /*namespaceUID*/, namespaceDisName, err := getNamespaceName(writer, claims, h.cfg.OIDC.StripEmaildomain)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "提取用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	userNameHead := string([]rune(namespaceDisName)[0])

	renderData := adminTemplateConfig{
		Basedomain:   h.cfg.BaseDomain,
		UserNameHead: userNameHead,
		UserName:     namespaceDisName,
		UserAccount:  namespaceName,
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(&renderData)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

func (h *Headscale) ConsoleMachinesAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	tokenCookie, _ := req.Cookie("OIDC_Token")
	rawToken := tokenCookie.Value
	idToken, err := h.verifyIDTokenForOIDCCallback(req.Context(), writer, rawToken)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "验证Token失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	claims, err := extractIDTokenClaims(writer, idToken)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "解析用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	namespaceName, _ /*namespaceUID*/, _ /*namespaceDisName*/, err := getNamespaceName(writer, claims, h.cfg.OIDC.StripEmaildomain)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "提取用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}

	UserMachines, err := h.ListMachinesInNamespace(namespaceName)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "查询用户节点列表失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}

	mlist := make(map[string]machineItem)
	for _, machine := range UserMachines {
		IPNver := machine.HostInfo.IPNVersion
		if strings.Contains(IPNver, "-") {
			IPNver = strings.Split(machine.HostInfo.IPNVersion, "-")[0]
		}
		tz, _ := time.LoadLocation("Asia/Shanghai")
		tmpMachine := machineItem{
			GiveName:    machine.GivenName,
			UserAccount: machine.Namespace.Name,
			OS:          machine.HostInfo.OS,
			Version:     IPNver,
			LastSeen:    machine.LastSeen.In(tz).Format("2006年01月02日 15:04:05"),
			IfOnline:    machine.isOnline(),
			MSubnetList: make([]string, 0),
		}
		if machine.IPAddresses[0].Is4() {
			tmpMachine.MIPv4 = machine.IPAddresses[0].String()
			tmpMachine.MIPv6 = machine.IPAddresses[1].String()
		} else if machine.IPAddresses[1].Is4() {
			tmpMachine.MIPv6 = machine.IPAddresses[0].String()
			tmpMachine.MIPv4 = machine.IPAddresses[1].String()
		}
		mlist["machine"+strconv.FormatUint(machine.ID, 10)] = tmpMachine
	}

	renderData := adminTemplateConfig{
		MList: mlist,
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(&renderData)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

func (h *Headscale) ConsolePanel(
	writer http.ResponseWriter,
	req *http.Request,
) {
	adminT := template.Must(template.New("admin").Parse(adminHTML))

	tokenCookie, _ := req.Cookie("OIDC_Token")
	rawToken := tokenCookie.Value
	idToken, err := h.verifyIDTokenForOIDCCallback(req.Context(), writer, rawToken)
	if err != nil {
		renderResult(writer, true, "验证Token失败", "/", "返回首页")
		return
	}
	claims, err := extractIDTokenClaims(writer, idToken)
	if err != nil {
		renderResult(writer, true, "解析用户信息失败", "/", "返回首页")
		return
	}
	namespaceName, _ /*namespaceUID*/, namespaceDisName, err := getNamespaceName(writer, claims, h.cfg.OIDC.StripEmaildomain)
	if err != nil {
		renderResult(writer, true, "提取用户信息失败", "/", "返回首页")
		return
	}
	userNameHead := string([]rune(namespaceDisName)[0])

	UserMachines, err := h.ListMachinesInNamespace(namespaceName)
	if err != nil {
		renderResult(writer, true, "查询用户节点列表失败", "/", "返回首页")
		return
	}

	mlist := make(map[string]machineItem)
	for _, machine := range UserMachines {
		IPNver := machine.HostInfo.IPNVersion
		if strings.Contains(IPNver, "-") {
			IPNver = strings.Split(machine.HostInfo.IPNVersion, "-")[0]
		}
		tz, _ := time.LoadLocation("Asia/Shanghai")
		tmpMachine := machineItem{
			GiveName:    machine.GivenName,
			UserAccount: machine.Namespace.Name,
			OS:          machine.HostInfo.OS,
			Version:     IPNver,
			LastSeen:    machine.LastSeen.In(tz).Format("2006年01月02日 15:04:05"),
			IfOnline:    machine.isOnline(),
			MSubnetList: make([]string, 0),
		}
		if machine.IPAddresses[0].Is4() {
			tmpMachine.MIPv4 = machine.IPAddresses[0].String()
			tmpMachine.MIPv6 = machine.IPAddresses[1].String()
		} else if machine.IPAddresses[1].Is4() {
			tmpMachine.MIPv6 = machine.IPAddresses[0].String()
			tmpMachine.MIPv4 = machine.IPAddresses[1].String()
		}
		mlist["machine"+strconv.FormatUint(machine.ID, 10)] = tmpMachine
	}

	renderData := adminTemplateConfig{
		Basedomain:   h.cfg.BaseDomain,
		UserNameHead: userNameHead,
		UserName:     namespaceDisName,
		UserAccount:  namespaceName,
		MList:        mlist,
	}

	var payload bytes.Buffer
	if err := adminT.Execute(&payload, renderData); err != nil {
		log.Error().
			Str("handler", "adminHTML").
			Err(err).
			Msg("Could not render admin HTML")

		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusInternalServerError)
		_, err := writer.Write([]byte("Could not render admin index template"))
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
	_, err = writer.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

//go:embed admin/result.html
var resultHTML string
var resultTemplate = template.Must(template.New("result").Parse(resultHTML))

type resultTemplateConfig struct {
	ERROR   bool
	Msg     string
	Next    string
	NextMsg string
}

// 用来为控制台网页生成结果页
func renderResult(
	writer http.ResponseWriter,
	isErr bool, Msg string, Next string, NextMsg string,
) error {
	var payload bytes.Buffer
	if err := resultTemplate.Execute(&payload, resultTemplateConfig{
		ERROR:   isErr,
		Msg:     Msg,
		Next:    Next,
		NextMsg: NextMsg,
	}); err != nil {
		log.Error().
			Str("func", "renderResult").
			Str("type", "Result").
			Err(err).
			Msg("Could not generate result page")
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

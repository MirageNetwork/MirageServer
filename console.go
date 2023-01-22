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

type machineItem struct {
	GiveName     string   `json:"givename"`
	UserAccount  string   `json:"useraccount"`
	UserNameHead string   `json:"usernamehead"`
	MIPv4        string   `json:"mipv4"`
	MIPv6        string   `json:"mipv6"`
	MSubnetList  []string `json:"msubnetlist"`
	OS           string   `json:"os"`
	OSHostName   string   `json:"oshostname"`
	Version      string   `json:"version"`
	IfOnline     bool     `json:"ifonline"`
	LastSeen     string   `json:"lastseen"`
	CreateAt     string   `json:"createat"`

	IsSharedIn       bool `json:"issharedin"`
	IsSharedOut      bool `json:"issharedout"`
	IsExpiryDisabled bool `json:"isexpirydisabled"`
	IsExitNode       bool `json:"isexitnode"`
	IsSubnet         bool `json:"issubnet"`

	Varies      bool `json:"varies"`
	HairPinning bool `json:"hairpinning"`
	CanIPv6     bool `json:"ipv6en"`
	CanUDP      bool `json:"udpen"`
	CanUPnP     bool `json:"upnpen"`
	CanPCP      bool `json:"pcpen"`
	CanPMP      bool `json:"pmpen"`

	ExpiryDesc string `json:"expirydesc"`

	Endpoints   []string       `json:"eps"`
	DERPs       map[string]int `json:"derps"`
	PrefferDERP string         `json:"usederp"`
}
type adminTemplateConfig struct {
	ErrorMsg     string                 `json:"errormsg"`
	Basedomain   string                 `json:"basedomain"`
	UserName     string                 `json:"username"`
	UserNameHead string                 `json:"usernamehead"`
	UserAccount  string                 `json:"useraccount"`
	MList        map[string]machineItem `json:"mlist"`
}

// 提供获取用户信息的API
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
	userName, _ /*UID*/, userDisName, err := getUserName(writer, claims, h.cfg.OIDC.StripEmaildomain)
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
	userNameHead := string([]rune(userDisName)[0])

	renderData := adminTemplateConfig{
		Basedomain:   h.cfg.BaseDomain,
		UserNameHead: userNameHead,
		UserName:     userDisName,
		UserAccount:  userName,
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

// 验证Token并获取用户信息
func (h *Headscale) verifyTokenIDandGetUser(
	writer http.ResponseWriter,
	req *http.Request,
) string {
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
		return ""
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
		return ""
	}
	userName, _ /*UID*/, _ /*userDisName*/, err := getUserName(writer, claims, h.cfg.OIDC.StripEmaildomain)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "提取用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return ""
	}
	return userName
}

// 控制台获取机器信息列表的API
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
	userName, _ /*UID*/, _ /*userDisName*/, err := getUserName(writer, claims, h.cfg.OIDC.StripEmaildomain)
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

	UserMachines, err := h.ListMachinesByUser(userName)
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
			GiveName:         machine.GivenName,
			UserAccount:      machine.User.Name,
			UserNameHead:     string([]rune(machine.User.Display_Name)[0]),
			OS:               machine.HostInfo.OS,
			OSHostName:       machine.HostInfo.Hostname,
			Version:          IPNver,
			CreateAt:         machine.CreatedAt.In(tz).Format("2006年01月02日 15:04:05"),
			LastSeen:         machine.LastSeen.In(tz).Format("2006年01月02日 15:04:05"),
			IfOnline:         machine.isOnline(),
			MSubnetList:      make([]string, 0),
			IsExpiryDisabled: *machine.Expiry == time.Time{},

			Varies:      machine.HostInfo.NetInfo.MappingVariesByDestIP.EqualBool(true),
			HairPinning: machine.HostInfo.NetInfo.HairPinning.EqualBool(true),
			CanIPv6:     machine.HostInfo.NetInfo.WorkingIPv6.EqualBool(true),
			CanUDP:      machine.HostInfo.NetInfo.WorkingUDP.EqualBool(true),
			CanUPnP:     machine.HostInfo.NetInfo.UPnP.EqualBool(true),
			CanPCP:      machine.HostInfo.NetInfo.PCP.EqualBool(true),
			CanPMP:      machine.HostInfo.NetInfo.PMP.EqualBool(true),
			Endpoints:   machine.Endpoints,
		}

		if machine.HostInfo.NetInfo.PreferredDERP != 0 {
			tmpMachine.DERPs = make(map[string]int)
			for derpname, latency := range machine.HostInfo.NetInfo.DERPLatency {
				ipver := strings.Split(derpname, "-")[1]
				derpname = strings.Split(derpname, "-")[0]
				if ipver == "v4" {
					if peerlatency, ok := machine.HostInfo.NetInfo.DERPLatency[derpname+"-v6"]; ok {
						if latency < peerlatency {
							tmpMachine.DERPs[derpname] = int(latency * 1000)
						}
					} else {
						tmpMachine.DERPs[derpname] = int(latency * 1000)
					}
				} else if ipver == "v6" {
					if peerlatency, ok := machine.HostInfo.NetInfo.DERPLatency[derpname+"-v4"]; ok {
						if latency < peerlatency {
							tmpMachine.DERPs[derpname] = int(latency * 1000)
						}
					} else {
						tmpMachine.DERPs[derpname] = int(latency * 1000)
					}
				} else {
					tmpMachine.DERPs[derpname] = int(latency * 1000)
				}
			}
			tmpMachine.PrefferDERP = strconv.Itoa(machine.HostInfo.NetInfo.PreferredDERP)
		} else {
			tmpMachine.PrefferDERP = "x"
			tmpMachine.DERPs = nil
		}
		if !tmpMachine.IsExpiryDisabled {
			ExpiryDuration := machine.Expiry.Sub(time.Now())
			if ExpiryDuration.Seconds() <= 0 {
				tmpMachine.ExpiryDesc = "已过期"
			} else if ExpiryDuration.Hours()/24/365 >= 1 {
				tmpMachine.ExpiryDesc = "还剩一年以上有效期"
			} else if ExpiryDuration.Hours()/24/30 >= 1 {
				tmpMachine.ExpiryDesc = "有效期还剩" + strconv.FormatInt(int64(ExpiryDuration.Hours()/24/30), 10) + "个月" + strconv.FormatInt(int64(ExpiryDuration.Hours()/24)-int64(ExpiryDuration.Hours()/24/30)*30, 10) + "天"
			} else if ExpiryDuration.Hours()/24 >= 1 {
				tmpMachine.ExpiryDesc = "有效期还剩" + strconv.FormatInt(int64(ExpiryDuration.Hours()/24), 10) + "天"
			} else if ExpiryDuration.Hours() >= 1 {
				tmpMachine.ExpiryDesc = "有效期还剩" + strconv.FormatInt(int64(ExpiryDuration.Hours()), 10) + "小时"
			} else if ExpiryDuration.Minutes() >= 1 {
				tmpMachine.ExpiryDesc = "有效期还剩" + strconv.FormatInt(int64(ExpiryDuration.Minutes()), 10) + "分钟"
			} else {
				tmpMachine.ExpiryDesc = "马上就要过期"
			}
		}
		if machine.IPAddresses[0].Is4() {
			tmpMachine.MIPv4 = machine.IPAddresses[0].String()
			tmpMachine.MIPv6 = machine.IPAddresses[1].String()
		} else if machine.IPAddresses[1].Is4() {
			tmpMachine.MIPv6 = machine.IPAddresses[0].String()
			tmpMachine.MIPv4 = machine.IPAddresses[1].String()
		}
		mlist[strconv.FormatUint(machine.ID, 10)] = tmpMachine
	}

	renderData := adminTemplateConfig{
		Basedomain: h.cfg.BaseDomain,
		MList:      mlist,
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

type removeMachineRes struct {
	Status string `json:"status"`
	ErrMsg string `json:"errmsg"`
}

func (h *Headscale) ConsoleRemoveMachineAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	userName := h.verifyTokenIDandGetUser(writer, req)
	resData := removeMachineRes{}
	if userName == "" {
		resData.Status = "Error"
		resData.ErrMsg = "用户信息核对失败"
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		err := json.NewEncoder(writer).Encode(&resData)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	UserMachines, err := h.ListMachinesByUser(userName)
	if err != nil {
		resData.Status = "Error"
		resData.ErrMsg = "用户设备检索失败"
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		err := json.NewEncoder(writer).Encode(&resData)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	err = req.ParseForm()
	if err != nil {
		resData.Status = "Error"
		resData.ErrMsg = "用户请求解析失败"
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		err := json.NewEncoder(writer).Encode(&resData)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	reqData := make(map[string]string)
	json.NewDecoder(req.Body).Decode(&reqData)
	wantRemoveID := reqData["mid"]
	for _, machine := range UserMachines {
		if strconv.FormatUint(machine.ID, 10) == wantRemoveID {
			err = h.HardDeleteMachine(&machine)
			if err != nil {
				resData.Status = "Error"
				resData.ErrMsg = "用户设备删除失败"
				writer.Header().Set("Content-Type", "application/json; charset=utf-8")
				writer.WriteHeader(http.StatusOK)
				err := json.NewEncoder(writer).Encode(&resData)
				if err != nil {
					log.Error().
						Caller().
						Err(err).
						Msg("Failed to write response")
				}
				return
			}
			resData.Status = "OK"
			resData.ErrMsg = "用户设备成功删除"
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			writer.WriteHeader(http.StatusOK)
			err := json.NewEncoder(writer).Encode(&resData)
			if err != nil {
				log.Error().
					Caller().
					Err(err).
					Msg("Failed to write response")
			}
			return
		}
	}
	resData.Status = "Error"
	resData.ErrMsg = "未找到目标设备"
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(&resData)
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

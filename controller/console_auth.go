package controller

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"tailscale.com/tailcfg"
	"tailscale.com/types/key"
)

const (
	// TODO(juan): remove this once https://github.com/juanfont/headscale/issues/727 is fixed.
	registrationHoldoff                      = time.Second * 5
	reservedResponseHeaderSize               = 4
	RegisterMethodAuthKey                    = "authkey"
	RegisterMethodOIDC                       = "oidc"
	RegisterMethodCLI                        = "cli"
	ErrRegisterMethodCLIDoesNotSupportExpire = Error(
		"machines registered with CLI does not support expire",
	)
)

func (h *Mirage) doLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	provider := r.Form["provider"][0]
	nextURL := r.Form["next_url"][0]
	nextURL, err := url.QueryUnescape(nextURL)
	if err != nil {
		h.ErrMessage(w, r, 500, "路径解析错误")
		return
	}
	stateCode := h.GenStateCode()
	stateCodeItem := StateCacheItem{
		nextURL:    nextURL,
		provider:   provider,
		uid:        -1,
		machineKey: key.MachinePublic{},
	}
	if strings.HasPrefix(nextURL, "/a/") {
		aCode := strings.TrimPrefix(nextURL, "/a/")
		aCodeC, ok := h.aCodeCache.Get(aCode)
		if ok && aCodeC.(ACacheItem).uid == -1 {
			stateCode = aCodeC.(ACacheItem).stateCode
			stateCodeC, ok := h.stateCodeCache.Get(stateCode)
			if ok && stateCodeC.(StateCacheItem).uid != -1 {
				h.ErrMessage(w, r, 400, "授权流程已进行")
				return
			}
			stateCodeItem = stateCodeC.(StateCacheItem)
			stateCodeItem.provider = provider
		}
	}
	h.stateCodeCache.Set(stateCode, stateCodeItem, time.Now().AddDate(0, 1, 0).Sub(time.Now()))
	stateCodeCookie := &http.Cookie{
		Name:     "mirage-authstate2",
		Value:    stateCode,
		Domain:   h.cfg.ServerURL,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, stateCodeCookie)

	switch provider {
	case "Github", "Microsoft", "Google", "Apple", "Ali":
		h.doDexLogin(w, r, stateCode, provider)
	case "WXScan":
		h.doWXScanLogin(w, r, stateCode)
	}
}

func (h *Mirage) doDexLogin(w http.ResponseWriter, r *http.Request, stateCode, provider string) {
	if h.cfg.OIDC.Issuer != "" {
		err := h.initOIDC()
		if err != nil {
			log.Warn().Err(err).Msg("failed to set up OIDC provider, falling back to CLI based authentication")
		}
	}

	extras := make([]oauth2.AuthCodeOption, 0, len(h.cfg.OIDC.ExtraParams))
	for k, v := range h.cfg.OIDC.ExtraParams {
		extras = append(extras, oauth2.SetAuthURLParam(k, v))
	}
	extras = append(extras, oauth2.SetAuthURLParam("connector_id", provider))

	log.Trace().Msg("之后会跳转到：" + fmt.Sprintf(
		"https://%s/%s",
		h.cfg.ServerURL,
		"a/oauth_response",
	))

	authURL := h.oauth2Config.AuthCodeURL(stateCode, extras...)
	log.Debug().Msgf("Redirecting to %s for authentication", authURL)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Mirage) checkWXMini(w http.ResponseWriter, r *http.Request) {
	reqData := make(map[string]string)
	json.NewDecoder(r.Body).Decode(&reqData)

	h.doWXScanLogin(w, r, reqData["state"])
}

func (h *Mirage) doWXScanLogin(w http.ResponseWriter, r *http.Request, stateCode string) {
	url := h.cfg.wxScanURL + "/fetchQR"
	message := map[string]string{"state": stateCode}

	// 将 message 转换为 JSON 格式
	requestBody, err := json.Marshal(message)
	if err != nil {
		log.Error().Caller().Msgf("创建微信小程序码拉取请求结构体出错")
	}

	// 创建一个新的请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error().Caller().Msgf("创建微信小程序码拉取请求出错")
	}

	// 设置请求的 Content-Type
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Caller().Msgf("发送微信小程序码拉取请求出错")
	}
	defer resp.Body.Close()

	resData := make(map[string]string)
	json.NewDecoder(resp.Body).Decode(&resData)
	resData["state"] = stateCode

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&resData)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
	return
}

func (h *Mirage) loginMidware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextURL := r.URL.Query().Get("next_url")
		refresh := r.URL.Query().Get("refresh")
		controlCodeCookie, err := r.Cookie("miragecontrol")
		if refresh == "true" || err == http.ErrNoCookie {
			next.ServeHTTP(w, r)
			return
		}
		_, controlCodeExpiration, ok := h.controlCodeCache.GetWithExpiration(controlCodeCookie.Value)
		if !ok || controlCodeExpiration.Compare(time.Now()) != 1 {
			next.ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, nextURL, http.StatusFound)
		return
	})
}

// WebUI控制台鉴权中间件
func (h *Mirage) ConsoleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "admin/api") {
			h.APIAuth(next).ServeHTTP(w, r)
			return
		}
		controlCodeCookie, err := r.Cookie("miragecontrol")
		if err == http.ErrNoCookie {
			log.Error().Msg("未能从Cookie读取到OIDC Token！")
			nextURL := r.URL.Path
			newQuery := r.URL.Query()
			newQuery.Add("next_url", nextURL)
			r.URL.RawQuery = newQuery.Encode()
			http.Redirect(w, r, "/login?"+r.URL.RawQuery, http.StatusFound)
			return
		}
		_, controcontrolCodeExpiration, ok := h.controlCodeCache.GetWithExpiration(controlCodeCookie.Value)
		if !ok || controcontrolCodeExpiration.Compare(time.Now()) != 1 {
			log.Debug().
				Caller().
				Msg("could not verifyIDTokenForOIDCCallback")
			nextURL := r.URL.Path
			newQuery := r.URL.Query()
			newQuery.Add("next_url", nextURL)
			r.URL.RawQuery = newQuery.Encode()
			http.Redirect(w, r, "/login?"+r.URL.RawQuery, http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// API鉴权结果响应
type APICheckRes struct {
	NeedReauth bool   `json:"needreauth"`
	Reason     string `json:"needreauthreason"`
}

// API鉴权中间件
func (h *Mirage) APIAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controlCodeCookie, err := r.Cookie("miragecontrol")
		if err == http.ErrNoCookie {
			log.Error().Msg("未能从Cookie读取到OIDC Token！")
			renderData := APICheckRes{
				NeedReauth: true,
				Reason:     "未读取到Token",
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&renderData)
			return
		}
		_, controcontrolCodeExpiration, ok := h.controlCodeCache.GetWithExpiration(controlCodeCookie.Value)
		if !ok || controcontrolCodeExpiration.Compare(time.Now()) != 1 {
			log.Debug().
				Caller().
				Msg("could not verifyIDTokenForOIDCCallback")
			renderData := APICheckRes{
				NeedReauth: true,
				Reason:     "IDaaS无法校验",
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&renderData)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Mirage) deviceRegPortal(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	aCode, ok := vars["aCode"]

	log.Debug().
		Caller().
		Str("ACode", aCode).
		Bool("ok", ok).
		Msg("Received oidc register call")

	//是普通aCode，先检查是否存在，不存在的回400
	aC, ok := h.aCodeCache.Get(aCode)
	if !ok {
		h.ErrMessage(w, r, 400, "未知的鉴别码")
		return
	}
	aCodeItem := aC.(ACacheItem)

	// 无论哪种情形，当前没有control都应该跳转到login页面
	controlCodeCookie, controlCodeErr := r.Cookie("miragecontrol")
	if controlCodeErr == http.ErrNoCookie {
		newQuery := r.URL.Query()
		newQuery.Add("next_url", "/a/"+aCode)
		r.URL.RawQuery = newQuery.Encode()
		http.Redirect(w, r, "/login?refresh=true&"+r.URL.RawQuery, http.StatusFound)
		return
	}
	//cookie中对应的control查不到，显示403授权过期
	controlCodeC, controlCodeExpiration, ok := h.controlCodeCache.GetWithExpiration(controlCodeCookie.Value)
	if !ok {
		if aCodeItem.uid != -1 {
			h.ErrMessage(w, r, 403, "网页授权已过期，请重新登陆")
			return
		} else {
			newQuery := r.URL.Query()
			newQuery.Add("next_url", "/a/"+aCode)
			r.URL.RawQuery = newQuery.Encode()
			http.Redirect(w, r, "/login?"+r.URL.RawQuery, http.StatusFound)
			return
		}
	}
	// 按TS官方做法似乎超过5分钟的control不能用于机器授权，跳转重新获取
	controlCodeItem := controlCodeC.(ControlCacheItem)
	if time.Now().AddDate(0, 1, 0).Sub(controlCodeExpiration) > time.Minute*5 {
		newQuery := r.URL.Query()
		newQuery.Add("next_url", "/a/"+aCode)
		r.URL.RawQuery = newQuery.Encode()
		http.Redirect(w, r, "/login?refresh=true&"+r.URL.RawQuery, http.StatusFound)
		return
	}

	// 对存在的aCode，判断是否有对应的用户
	// 1、还没确定对应用户，确认绑定显示connectDevice页面
	// 2、已确定对应用户，但还未确认绑定，显示400页面
	// 3、已确认接入设备，显示跳转页面（？设备授权页面）

	if aCodeItem.uid == -1 {
		//未绑定用户，显示connectDevice
		user, _ := h.GetUserByID(controlCodeItem.uid)
		Hostname := aCodeItem.regReq.Hostinfo.Hostname
		Netname := user.Name
		Nodekey := aCodeItem.regReq.NodeKey.String()
		OS := aCodeItem.regReq.Hostinfo.OS + "(" + aCodeItem.regReq.Hostinfo.OSVersion + ")"
		ClientVer := aCodeItem.regReq.Hostinfo.IPNVersion
		NextURL := "/a/" + aCode
		h.sendConnectDevicePage(w, r, Hostname, Netname, Nodekey, OS, ClientVer, NextURL)
		return
	}

	if aCodeItem.uid != controlCodeItem.uid {
		h.ErrMessage(w, r, 403, "用户未被授权查看此页面")
		return
	}

	machine, err := h.GetMachineByNodeKey(aCodeItem.regReq.NodeKey)
	if err != nil {
		h.ErrMessage(w, r, 500, "获取设备信息出错")
		return
	}
	Hostname := machine.GivenName
	Netname := machine.User.Name
	MIP := machine.IPAddresses[0].String()
	if len(machine.IPAddresses) > 1 && machine.IPAddresses[1].Is4() {
		MIP = machine.IPAddresses[1].String()
	}
	h.sendDeviceRedirectPage(w, r, Hostname, Netname, MIP)
	// 做过用户登录，接下来判断是否已连接机器??

}

// 处理connectDevice页面的POST请求，用于真正注册设备
func (h *Mirage) deviceReg(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	aCode, ok := vars["aCode"]

	log.Debug().
		Caller().
		Str("ACode", aCode).
		Bool("ok", ok).
		Msg("Received connect device call")

	//是普通aCode，先检查是否存在，不存在的回400
	aC, aCodeExpiration, ok := h.aCodeCache.GetWithExpiration(aCode)
	if !ok {
		h.ErrMessage(w, r, 400, "未知的鉴别码")
		return
	}
	aCodeItem := aC.(ACacheItem)
	// 无论哪种情形，当前没有control都应该跳转到login页面
	controlCodeCookie, controlCodeErr := r.Cookie("miragecontrol")
	if controlCodeErr == http.ErrNoCookie {
		newQuery := r.URL.Query()
		newQuery.Add("next_url", "/a/"+aCode)
		r.URL.RawQuery = newQuery.Encode()
		http.Redirect(w, r, "/login?"+r.URL.RawQuery, http.StatusFound)
		return
	}
	//cookie中对应的control查不到，显示403授权过期
	controlCodeC, controlCodeExpiration, ok := h.controlCodeCache.GetWithExpiration(controlCodeCookie.Value)
	if !ok {
		newQuery := r.URL.Query()
		newQuery.Add("next_url", "/a/"+aCode)
		r.URL.RawQuery = newQuery.Encode()
		http.Redirect(w, r, "/login?"+r.URL.RawQuery, http.StatusFound)
		return
	}
	// 按TS官方做法似乎超过5分钟的control不能用于机器授权，跳转重新获取
	controlCodeItem := controlCodeC.(ControlCacheItem)
	if time.Now().AddDate(0, 1, 0).Sub(controlCodeExpiration) > time.Minute*5 {
		newQuery := r.URL.Query()
		newQuery.Add("next_url", "/a/"+aCode)
		r.URL.RawQuery = newQuery.Encode()
		http.Redirect(w, r, "/login?"+r.URL.RawQuery, http.StatusFound)
		return
	}
	// 已经完成过设备接入
	if aCodeItem.uid != -1 {
		if aCodeItem.uid != controlCodeItem.uid {
			h.ErrMessage(w, r, 403, "你无权进行此操作")
			return
		}
		machine, err := h.GetMachineByNodeKey(aCodeItem.regReq.NodeKey)
		if err != nil {
			h.ErrMessage(w, r, 500, "获取设备信息出错")
			return
		}
		Hostname := machine.GivenName
		Netname := machine.User.Name
		MIP := machine.IPAddresses[0].String()
		if len(machine.IPAddresses) > 1 && machine.IPAddresses[1].Is4() {
			MIP = machine.IPAddresses[1].String()
		}
		h.sendDeviceRedirectPage(w, r, Hostname, Netname, MIP)
	}

	aCodeItem.uid = controlCodeItem.uid
	h.aCodeCache.Set(aCode, aCodeItem, aCodeExpiration.Sub(time.Now()))
	// 过期时间先按照用户标准过期时间，后续可以考虑加入单独设备设置，与用户标准联合限制
	machine, err := h.registerMachineFromConsole(aCodeItem)
	if err != nil {
		h.ErrMessage(w, r, 500, "注册设备信息出错")
		return
	}

	h.longPollChanPool[aCode] <- "ok" // longpoll的救赎

	Hostname := machine.GivenName
	Netname := machine.User.Name
	MIP := machine.IPAddresses[0].String()
	if len(machine.IPAddresses) > 1 && machine.IPAddresses[1].Is4() {
		MIP = machine.IPAddresses[1].String()
	}
	h.sendDeviceRedirectPage(w, r, Hostname, Netname, MIP)
}

//go:embed templates/OrgSelector.html
var OrgSelectTemplate string

// 接受OIDC认证返回的GET，进行跳转和Token写入
func (h *Mirage) oauthResponse(
	w http.ResponseWriter,
	r *http.Request,
) {
	qState := r.URL.Query().Get("state")
	// 之后对返回判断进行校验
	qStateC, qStateExpiration, ok := h.stateCodeCache.GetWithExpiration(qState)
	if !ok {
		h.ErrMessage(w, r, 409, "未知的state参数")
		return
	}
	qStateItem := qStateC.(StateCacheItem)
	// 对于任何已经之前经过认证的stateCode都往目标URL跳转，由目标URL校验是否放行
	if qStateItem.uid != -1 {
		http.Redirect(w, r, qStateItem.nextURL, http.StatusFound)
		return
	}
	// 对于还未经过认证（即无对应用户身份信息的），需要进行oauth验证流程
	code := r.URL.Query().Get("code")
	cState, err := r.Cookie("mirage-authstate2")
	if err == http.ErrNoCookie {
		h.ErrMessage(w, r, 401, "authstate2曲奇缺失")
		return
	}
	if cState.Value != qState {
		h.ErrMessage(w, r, 401, "authstate2曲奇不匹配")
		return
	}
	// TODO: 后续多Provider时从state码中读取对应的校验器
	userName := ""
	userDisName := ""
	orgName := ""
	switch qStateItem.provider {
	case "Microsoft", "Google", "Github", "Apple", "Ali":
		oauth2Token, err := h.oauth2Config.Exchange(r.Context(), code)
		if err != nil {
			h.ErrMessage(w, r, 403, "三方登录认证错误")
			return
		}
		rawIDToken, rawIDTokenOK := oauth2Token.Extra("id_token").(string)
		if !rawIDTokenOK {
			h.ErrMessage(w, r, 403, "三方登录认证解析错误1")
			return
		}
		idToken, err := h.verifyIDTokenForOIDCCallback(r.Context(), w, rawIDToken)
		if err != nil {
			h.ErrMessage(w, r, 403, "三方登录认证解析错误2")
			return
		}
		claims, err := extractIDTokenClaims(w, idToken)
		if err != nil {
			h.ErrMessage(w, r, 403, "三方登录认证解析用户错误")
			return
		}
		//userName, userDisName, err = getUserName(w, claims, h.cfg.OIDC.StripEmaildomain)
		userName = claims.Email
		userDisName = claims.Name
		if err != nil {
			h.ErrMessage(w, r, 500, "三方登录用户信息解析出错")
			return
		}
		qStateItem.userName = userName
		qStateItem.userDisName = userDisName
		h.stateCodeCache.Set(qState, qStateItem, qStateExpiration.Sub(time.Now()))

		if claims.Groups == nil || len(claims.Groups) == 0 {
			orgName = userName
		} else if len(claims.Groups) == 1 { // 对Github而言，至少有一个个人组织，是Groups中的最末一项
			orgName = claims.Groups[0]
		} else { // 渲染组织选择页面
			if qStateItem.provider == "Github" { // 除Github之外其他情况有待讨论
				orgSelectT := template.Must(template.New("orgSelector").Parse(OrgSelectTemplate))

				config := map[string]interface{}{
					"State":         qState,
					"UserName":      strings.TrimSuffix(userName, "@github"),
					"PersonalGroup": claims.Groups[len(claims.Groups)-1],
					"Groups":        claims.Groups[:len(claims.Groups)-1],
				}

				var payload bytes.Buffer
				if err := orgSelectT.Execute(&payload, config); err != nil {
					log.Error().
						Str("handler", "orgSelector").
						Err(err).
						Msg("Could not render orgSelector template")

					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					w.WriteHeader(http.StatusInternalServerError)
					_, err := w.Write([]byte("Could not render orgSelector template"))
					if err != nil {
						log.Error().
							Caller().
							Err(err).
							Msg("Failed to write response")
					}
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(payload.Bytes())
				if err != nil {
					log.Error().
						Caller().
						Err(err).
						Msg("Failed to write response")
				}
				return
			}
		}
	case "WXScan":
		url := h.cfg.wxScanURL + "/verify"
		message := map[string]string{"code": code}
		// 将 message 转换为 JSON 格式
		requestBody, err := json.Marshal(message)
		if err != nil {
			log.Error().Caller().Msgf("创建微信小程序码验证请求结构体出错")
		}
		// 创建一个新的请求
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			log.Error().Caller().Msgf("创建微信小程序码验证请求出错")
		}
		// 设置请求的 Content-Type
		req.Header.Set("Content-Type", "application/json")
		// 发送请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Error().Caller().Msgf("发送微信小程序码验证请求出错")
		}
		defer resp.Body.Close()

		resData := make(map[string]string)
		json.NewDecoder(resp.Body).Decode(&resData)
		if resData["status"] == "OK" {
			userName = resData["user_name"]
			userDisName = resData["display_name"]
			orgName = userName + ".WeChat"
		}

		qStateItem.userName = userName
		qStateItem.userDisName = userDisName
		h.stateCodeCache.Set(qState, qStateItem, qStateExpiration.Sub(time.Now()))
	}

	h.finishOauthResponse(w, r, qState, orgName)
}

// 接受选择组织的请求
func (h *Mirage) selectOrgForLogin(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.ParseForm()
	state := r.Form["state"][0]
	orgName := r.Form["org"][0]
	h.finishOauthResponse(w, r, state, orgName)
}

// 真正完成登录或注册
func (h *Mirage) finishOauthResponse(
	w http.ResponseWriter,
	r *http.Request,
	state string,
	OrgName string,
) {
	stateC, qStateExpiration, ok := h.stateCodeCache.GetWithExpiration(state)
	if !ok {
		h.ErrMessage(w, r, 409, "未知的state参数")
		return
	}
	stateItem := stateC.(StateCacheItem)
	// 对于任何已经之前经过认证的stateCode都往目标URL跳转，由目标URL校验是否放行
	if stateItem.uid != -1 {
		http.Redirect(w, r, stateItem.nextURL, http.StatusFound)
		return
	}
	// TODO:添加判断用户是否存在及自动创建逻辑
	user, err := h.findOrCreateNewUserForOIDCCallback(stateItem.userName, stateItem.userDisName, OrgName, stateItem.provider)
	if err != nil { // TODO: 后续这里理论上不会出错，因为会自动创建用户
		h.ErrMessage(w, r, 500, "服务器用户获取出错")
		return
	}
	stateItem.uid = user.toTailscaleUser().ID
	h.stateCodeCache.Set(state, stateItem, qStateExpiration.Sub(time.Now()))
	controlCode := h.GenStateCode()
	h.controlCodeCache.Set(
		controlCode,
		ControlCacheItem{
			uid: user.toTailscaleUser().ID,
		},
		time.Now().AddDate(0, 1, 0).Sub(time.Now()),
	)
	machineKey := stateItem.machineKey
	// 确认state来自机器注册用，需要记录与机器码对应关系，后续机器有新注册时要将原有对应control全部删除
	if !machineKey.IsZero() {
		machineControlCodes, machineControlCodeExpiration, ok := h.machineControlCodeCache.GetWithExpiration(machineKey.String())
		if !ok {
			machineControlCodes = MachineControlCodeCacheItem{
				controlCodes: make([]string, 0),
			}
			machineControlCodeExpiration = time.Now().AddDate(0, 1, 0)
		}
		machineControlItem := machineControlCodes.(MachineControlCodeCacheItem)
		machineControlItem.controlCodes = append(machineControlItem.controlCodes, controlCode)
		h.machineControlCodeCache.Set(machineKey.String(), machineControlItem, machineControlCodeExpiration.Sub(time.Now()))
	}

	controlCodeCookie := &http.Cookie{
		Name:     "miragecontrol",
		Value:    controlCode,
		Domain:   h.cfg.ServerURL,
		Path:     "/",
		Expires:  time.Now().AddDate(0, 1, 0),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, controlCodeCookie)
	http.Redirect(w, r, stateItem.nextURL, http.StatusFound)
	return
}

type StateCacheItem struct {
	nextURL     string
	provider    string
	uid         tailcfg.UserID
	userName    string
	userDisName string
	machineKey  key.MachinePublic
}

type ControlCacheItem struct {
	uid tailcfg.UserID
}

type MachineControlCodeCacheItem struct {
	controlCodes []string
}

//go:embed templates/connectDevice.html
var connectDeviceTemplate string

// 接入设备页面
func (h *Mirage) sendConnectDevicePage(
	w http.ResponseWriter,
	r *http.Request,
	Hostname, Netname, Nodekey, OS, ClientVer, NextURL string,
) {
	connDevT := template.Must(template.New("connectDevice").Parse(connectDeviceTemplate))

	config := map[string]interface{}{
		"Hostname":  Hostname,
		"Netname":   Netname,
		"Nodekey":   Nodekey,
		"OS":        OS,
		"ClientVer": ClientVer,
		"NextURL":   NextURL,
	}

	var payload bytes.Buffer
	if err := connDevT.Execute(&payload, config); err != nil {
		log.Error().
			Str("handler", "connectDevice").
			Err(err).
			Msg("Could not render connectDevice template")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Could not render connectDevice template"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

//go:embed templates/deviceRedirect.html
var deviceRedirectTemplate string

// 接入设备页面
func (h *Mirage) sendDeviceRedirectPage(
	w http.ResponseWriter,
	r *http.Request,
	Hostname, Netname, MIP string,
) {
	devRedirectT := template.Must(template.New("devRedirect").Parse(deviceRedirectTemplate))

	config := map[string]interface{}{
		"Hostname": Hostname,
		"Netname":  Netname,
		"MIP":      MIP,
	}

	var payload bytes.Buffer
	if err := devRedirectT.Execute(&payload, config); err != nil {
		log.Error().
			Str("handler", "deviceRedirect").
			Err(err).
			Msg("Could not render deviceRedirect template")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Could not render deviceRedirect template"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

func (h *Mirage) registerMachineFromConsole(
	aCodeItem ACacheItem,
) (*Machine, error) {
	nodeKey := aCodeItem.regReq.NodeKey
	user, err := h.GetUserByID(aCodeItem.uid)
	if err != nil {
		return nil, err
	}
	log.Debug().
		Str("machineKey", aCodeItem.mKey.ShortString()).
		Str("nodeKey", nodeKey.ShortString()).
		Str("userName", user.Name).
		Str("expiresAt", fmt.Sprintf("%v", time.Now().AddDate(0, 0, int(user.Org.ExpiryDuration)))).
		Msg("Registering machine from console confirm")

	now := time.Now()
	expiration := time.Now().AddDate(0, 0, int(user.Org.ExpiryDuration))
	givenName := h.GenMachineName(aCodeItem.regReq.Hostinfo.Hostname, aCodeItem.uid, MachinePublicKeyStripPrefix(aCodeItem.mKey))
	newmachine := Machine{
		MachineKey:           MachinePublicKeyStripPrefix(aCodeItem.mKey),
		Hostname:             aCodeItem.regReq.Hostinfo.Hostname,
		GivenName:            givenName,
		AutoGenName:          true,
		NodeKey:              NodePublicKeyStripPrefix(aCodeItem.regReq.NodeKey),
		UserID:               user.ID,
		ForcedTags:           aCodeItem.regReq.Hostinfo.RequestTags,
		LastSeen:             &now,
		LastSuccessfulUpdate: &now,
		Expiry:               &expiration,
		HostInfo:             HostInfo(*aCodeItem.regReq.Hostinfo.Clone()),
	}
	oldmachine, _ := h.GetUserMachineByMachineKey(aCodeItem.mKey, aCodeItem.uid)

	if oldmachine != nil {
		log.Trace().
			Caller().
			Str("machine", oldmachine.Hostname).
			Msg("machine already registered, reauthenticating")
		newmachine.ID = oldmachine.ID
		newmachine.GivenName = oldmachine.GivenName
		newmachine.AutoGenName = oldmachine.AutoGenName
		newmachine.IPAddresses = oldmachine.IPAddresses
		newmachine.RegisterMethod = RegisterMethodOIDC
		err := h.RestructMachine(&newmachine, expiration)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to restruct machine")
			return nil, ErrCouldNotConvertMachineInterface
		}
		machine, err := h.GetMachineByID(newmachine.ID)
		return machine, nil
	} else {
		machine, err := h.RegisterMachine(newmachine)
		return machine, err
	}
}

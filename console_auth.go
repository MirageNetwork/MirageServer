package headscale

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

func (h *Headscale) doLogin(w http.ResponseWriter, r *http.Request) {
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
		}
	}
	h.stateCodeCache.Set(stateCode, stateCodeItem, time.Now().AddDate(0, 1, 0).Sub(time.Now()))
	stateCodeCookie := &http.Cookie{
		Name:     "mirage-authstate2",
		Value:    stateCode,
		Domain:   strings.Split(h.cfg.ServerURL, "://")[1],
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, stateCodeCookie)

	switch provider {
	case "Ali":
		h.doAliLogin(w, r, stateCode)
	}
}
func (h *Headscale) doAliLogin(w http.ResponseWriter, r *http.Request, stateCode string) {
	extras := make([]oauth2.AuthCodeOption, 0, len(h.cfg.OIDC.ExtraParams))
	for k, v := range h.cfg.OIDC.ExtraParams {
		extras = append(extras, oauth2.SetAuthURLParam(k, v))
	}

	log.Error().Msg("之后会跳转到：" + fmt.Sprintf(
		"%s/%s",
		strings.TrimSuffix(h.cfg.ServerURL, "/"),
		"a/oauth_response",
	))

	adminOauth2Config := &oauth2.Config{
		ClientID:     h.cfg.OIDC.ClientID,
		ClientSecret: h.cfg.OIDC.ClientSecret,
		Endpoint:     h.oidcProvider.Endpoint(),
		RedirectURL: fmt.Sprintf(
			"%s/%s",
			strings.TrimSuffix(h.cfg.ServerURL, "/"),
			"a/oauth_response",
		),
		Scopes: h.cfg.OIDC.Scope,
	}

	authURL := adminOauth2Config.AuthCodeURL(stateCode, extras...)
	log.Debug().Msgf("Redirecting to %s for authentication", authURL)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Headscale) loginMidware(next http.Handler) http.Handler {
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
func (h *Headscale) ConsoleAuth(next http.Handler) http.Handler {
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
			log.Error().
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
func (h *Headscale) APIAuth(next http.Handler) http.Handler {
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
			log.Error().
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

func (h *Headscale) deviceRegPortal(
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
		Hostname := aCodeItem.regReq.Hostinfo.Hostname
		user, _ := h.GetUserByID(controlCodeItem.uid)
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
func (h *Headscale) deviceReg(
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
	Hostname := machine.GivenName
	Netname := machine.User.Name
	MIP := machine.IPAddresses[0].String()
	if len(machine.IPAddresses) > 1 && machine.IPAddresses[1].Is4() {
		MIP = machine.IPAddresses[1].String()
	}
	h.sendDeviceRedirectPage(w, r, Hostname, Netname, MIP)
}

// 接受OIDC认证返回的GET，进行跳转和Token写入
func (h *Headscale) oauthResponse(
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
	userName, _ /*UID*/, _ /*userDisName*/, err := getUserName(w, claims, h.cfg.OIDC.StripEmaildomain)
	if err != nil { // TODO: 后续这里理论上不会出错，因为会自动创建用户
		h.ErrMessage(w, r, 500, "服务器用户查询出错")
		return
	}
	// TODO:添加判断用户是否存在及自动创建逻辑
	user, err := h.GetUser(userName)
	if err != nil { // TODO: 后续这里理论上不会出错，因为会自动创建用户
		h.ErrMessage(w, r, 500, "服务器用户查询出错")
		return
	}
	qStateItem.uid = user.toTailscaleUser().ID
	h.stateCodeCache.Set(qState, qStateItem, qStateExpiration.Sub(time.Now()))
	controlCode := h.GenStateCode()
	h.controlCodeCache.Set(
		controlCode,
		ControlCacheItem{
			uid: user.toTailscaleUser().ID,
		},
		time.Now().AddDate(0, 1, 0).Sub(time.Now()),
	)
	machineKey := qStateItem.machineKey
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
		Domain:   strings.Split(h.cfg.ServerURL, "://")[1],
		Path:     "/",
		Expires:  time.Now().AddDate(0, 1, 0),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, controlCodeCookie)
	http.Redirect(w, r, qStateItem.nextURL, http.StatusFound)
	return
}

type StateCacheItem struct {
	nextURL    string
	provider   string
	uid        tailcfg.UserID
	machineKey key.MachinePublic
}

type ControlCacheItem struct {
	uid tailcfg.UserID
}

type MachineControlCodeCacheItem struct {
	controlCodes []string
}

//go:embed templates/BadCode.html
var errTemplate string

// cgao6: 用这个向前端返回错误页面
func (h *Headscale) ErrMessage(
	w http.ResponseWriter,
	r *http.Request,
	code int,
	msg string,
) {
	errT := template.Must(template.New("err").Parse(errTemplate))

	config := map[string]interface{}{
		"ErrCode": code,
		"ErrMsg":  msg,
	}

	var payload bytes.Buffer
	if err := errT.Execute(&payload, config); err != nil {
		log.Error().
			Str("handler", "ErrMessage").
			Err(err).
			Msg("Could not render Error template")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Could not render Error template"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, err := w.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

//go:embed templates/connectDevice.html
var connectDeviceTemplate string

// 接入设备页面
func (h *Headscale) sendConnectDevicePage(
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
func (h *Headscale) sendDeviceRedirectPage(
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

func (h *Headscale) registerMachineFromConsole(
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
		Str("expiresAt", fmt.Sprintf("%v", time.Now().AddDate(0, 0, int(user.ExpiryDuration)))).
		Msg("Registering machine from console confirm")

	now := time.Now()
	expiration := time.Now().AddDate(0, 0, int(user.ExpiryDuration))
	givenName := h.GenMachineName(aCodeItem.regReq.Hostinfo.Hostname, aCodeItem.uid, aCodeItem.mKey)
	newmachine := Machine{
		MachineKey:           MachinePublicKeyStripPrefix(aCodeItem.mKey),
		Hostname:             aCodeItem.regReq.Hostinfo.Hostname,
		GivenName:            givenName,
		AutoGenName:          true,
		NodeKey:              NodePublicKeyStripPrefix(aCodeItem.regReq.NodeKey),
		UserID:               user.ID,
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

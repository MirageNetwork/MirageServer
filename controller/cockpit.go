package controller

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"math/rand"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	webauthn "github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"go4.org/netipx"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type Cockpit struct {
	App  *Mirage
	db   *gorm.DB
	Addr string

	serviceState bool
	CtrlChn      chan CtrlMsg
	MsgChn       chan CtrlMsg
	hasAdmin     bool

	author     *webauthn.WebAuthn
	superAdmin *MirageSuperAdmin
	authCache  *cache.Cache
}

type CtrlMsg struct {
	Msg    string
	Err    error
	SysCfg *Config
}

type MirageSuperAdmin struct {
	cred AdminCredential
}

func (user *MirageSuperAdmin) WebAuthnID() []byte {
	return []byte("MirageSuperAdmin")
}
func (user *MirageSuperAdmin) WebAuthnName() string {
	return "MirageSuperAdmin"
}
func (user *MirageSuperAdmin) WebAuthnDisplayName() string {
	return "蜃境超级管理员"
}
func (user *MirageSuperAdmin) WebAuthnIcon() string {
	return ""
}
func (user *MirageSuperAdmin) WebAuthnCredentials() []webauthn.Credential {
	return []webauthn.Credential{webauthn.Credential(user.cred)}
}

// NewCockpit 创建一个新的Cockpit实例
// @sysAddr 监听地址
// @ctrlChn 对外控制通道
// @db 数据库实例
// @return Cockpit实例
// @return 错误信息，如果没有错误则为nil
func NewCockpit(sysAddr string, ctrlChn, msgChn chan CtrlMsg, db *gorm.DB) (*Cockpit, error) {
	cockpit := &Cockpit{
		db:   db,
		Addr: sysAddr,

		serviceState: false,
		CtrlChn:      ctrlChn,
		MsgChn:       msgChn,
	}
	cockpit.authCache = cache.New(0, 0)
	cockpit.superAdmin, cockpit.hasAdmin = cockpit.GetAdmin()

	return cockpit, nil
}

// isAdminSet 检查是否已经设置了管理员凭证
// @return 如果已经设置了管理员凭证则返回true，否则返回false
func (c *Cockpit) GetAdmin() (*MirageSuperAdmin, bool) {
	if sysCfg := c.GetSysCfg(); sysCfg != nil {
		return &MirageSuperAdmin{
			cred: sysCfg.AdminCredential,
		}, true
	}
	return nil, false
}

func (c *Cockpit) GetSysCfg() *SysConfig {
	cfg := []SysConfig{}
	err := c.db.Find(&cfg).Error
	if err != nil || cfg == nil || len(cfg) == 0 {
		return nil
	}
	if cfg[0].NaviDeployKey == "" {
		pri, pub, err := genSSHKeypair()
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		cfg[0].NaviDeployPub = pub
		cfg[0].NaviDeployKey = pri
		err = c.db.Save(&cfg[0]).Error
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}
	return &cfg[0]
}

//go:embed html/cockpit
var cockpitFS embed.FS

// createRouter 创建路由
// @return 路由实例
func (c *Cockpit) createRouter() *mux.Router {
	router := mux.NewRouter()

	cockpitDir, err := fs.Sub(cockpitFS, "html/cockpit")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	cockpit_router := router.PathPrefix("/cockpit").Subrouter()
	cockpit_router.Use(c.Auth)

	cockpit_router.HandleFunc("/api/regAdmin", c.RegisterAdmin).Methods(http.MethodPost)
	cockpit_router.HandleFunc("/api/login", c.Login).Methods(http.MethodPost)
	cockpit_router.HandleFunc("/api/setting/general", c.SetSettingGeneral).Methods(http.MethodPost)
	cockpit_router.HandleFunc("/api/service/start", c.DoServiceStart).Methods(http.MethodPost)
	cockpit_router.HandleFunc("/api/service/stop", c.DoServiceStop).Methods(http.MethodPost)
	cockpit_router.HandleFunc("/api/tenants", c.CAPIPostTenants).Methods(http.MethodPost)
	cockpit_router.HandleFunc("/api/publish/{os}", c.CAPIPublishClient).Methods(http.MethodPost)
	cockpit_router.HandleFunc("/api/derp/add", c.CAPIAddDERP).Methods(http.MethodPost)

	cockpit_router.HandleFunc("/api/logout", c.Logout).Methods(http.MethodGet)
	cockpit_router.HandleFunc("/api/service/state", c.GetServiceState).Methods(http.MethodGet)
	cockpit_router.HandleFunc("/api/setting/general", c.GetSettingGeneral).Methods(http.MethodGet)
	cockpit_router.HandleFunc("/api/tenants", c.CAPIGetTenant).Methods(http.MethodGet)
	cockpit_router.HandleFunc("/api/derp/query", c.CAPIQueryDERP).Methods(http.MethodGet)

	cockpit_router.PathPrefix("/api/derp/{id}").HandlerFunc(c.CAPIDelNaviNode).Methods(http.MethodDelete)

	cockpit_router.PathPrefix("").Handler(http.StripPrefix("/cockpit", http.FileServer(http.FS(cockpitDir))))

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.ErrMessage(w, r, 404, "你迷失在蜃境中了吗？这里什么都没有")
	})
	return router
}

// Auth 验证用户是否已经登录
func (c *Cockpit) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/cockpit/assets/") || strings.HasPrefix(r.URL.Path, "/cockpit/imgs/") {
			next.ServeHTTP(w, r)
			return
		}

		// If admin credential is not set, allow user to register admin credential
		if !c.hasAdmin {
			if r.URL.Path == "/cockpit/api/regAdmin" {
				next.ServeHTTP(w, r)
				return
			}
			// If user is not trying to register admin credential, return noadmin error
			if strings.Contains(r.URL.Path, "/cockpit/api/") {
				c.doAPIResponse(w, "noadmin", nil)
				return
			}
			next.ServeHTTP(w, r) // If user is not trying to access API, allow user to
			return
		}

		if r.URL.Path == "/cockpit/api/login" {
			next.ServeHTTP(w, r)
			return
		}

		// If user is not trying to login, check if user is logged in
		authCookie, err := r.Cookie("mirage_cockpit_auth")
		AuthCode, ok := c.authCache.Get("AuthCode")
		if err != nil || !ok || authCookie.Value != AuthCode.(string) {
			// If user is not logged in, return unauthorized error
			if strings.Contains(r.URL.Path, "/cockpit/api/") {
				c.doAPIResponse(w, "unauthorized", nil)
				return
			}
			next.ServeHTTP(w, r) // If user is not trying to access API, allow user to
			return
		}
		next.ServeHTTP(w, r) // If user is logged in, allow user to access API
	})
}

// RegisterAdmin 注册超级管理员
func (c *Cockpit) RegisterAdmin(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Query().Get("phase") == "response" { // 注册响应
		response, err := protocol.ParseCredentialCreationResponseBody(r.Body)
		if err != nil {
			c.doAPIResponse(w, "解析注册响应失败", nil)
			return
		}

		sessionInt, ok := c.authCache.Get("MirageSuperAdmin")
		if !ok || sessionInt == nil {
			c.doAPIResponse(w, "注册会话不存在或已过期", nil)
			return
		}
		session := sessionInt.(*webauthn.SessionData)

		credential, err := c.author.CreateCredential(c.superAdmin, *session, response)
		if err != nil {
			c.doAPIResponse(w, "创建超管凭证失败", nil)
			return
		}
		newSysCfg := c.GetSysCfg()
		if newSysCfg != nil {
			newSysCfg.AdminCredential = AdminCredential(*credential)
		} else {
			dexSecret := c.GenAuthCode()
			newSysCfg = &SysConfig{
				AdminCredential: AdminCredential(*credential),
				DexSecret:       dexSecret,
			}
		}
		c.db.Save(newSysCfg)
		c.superAdmin, c.hasAdmin = c.GetAdmin()
		if !c.hasAdmin {
			c.doAPIResponse(w, "保存超管凭证失败", nil)
			return
		}
		c.doAPIResponse(w, "", "ok")
		return
	} else { // 注册请求
		if c.author == nil {
			wconfig := &webauthn.Config{
				RPDisplayName: "蜃境网络",                        // Display Name for your site
				RPID:          r.Host,                        // Generally the FQDN for your site
				RPOrigins:     []string{"https://" + r.Host}, //[]string{"https://" + serverURL}, // The origin URLs allowed for WebAuthn requests
			}
			webAuthor, err := webauthn.New(wconfig)
			if err != nil {
				c.doAPIResponse(w, "创建WebAuthn验证器失败", nil)
				return
			}
			c.author = webAuthor
		}
		options, webAuthSession, err := c.author.BeginRegistration(c.superAdmin)
		c.authCache.Set("MirageSuperAdmin", webAuthSession, 5*time.Minute)
		if err != nil {
			c.doAPIResponse(w, "启动超管注册失败", nil)
			return
		}
		c.doAPIResponse(w, "", options)
	}
}

func (c *Cockpit) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.URL.Query().Get("phase") == "response" { // 登录响应
		response, err := protocol.ParseCredentialRequestResponseBody(r.Body)
		if err != nil {
			c.doAPIResponse(w, "解析登录响应失败", nil)
			return
		}
		sessionInt, ok := c.authCache.Get("MirageSuperAdmin")
		if !ok || sessionInt == nil {
			c.doAPIResponse(w, "登录会话不存在或已过期", nil)
			return
		}
		_, err = c.author.ValidateLogin(c.superAdmin, *sessionInt.(*webauthn.SessionData), response)
		if err != nil {
			c.doAPIResponse(w, "登录验证失败", nil)
			return
		}

		// 登录成功，生成 authCode 并返回给客户端
		authCode := c.GenAuthCode()
		c.authCache.Set("AuthCode", authCode, 5*time.Hour)
		authCookie := &http.Cookie{
			Name:     "mirage_cockpit_auth",
			Value:    authCode,
			Domain:   r.URL.Host,
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(1 * time.Hour),
		}
		http.SetCookie(w, authCookie)
		c.doAPIResponse(w, "", "ok")
	} else { // 登录请求
		if c.author == nil {
			wconfig := &webauthn.Config{
				RPDisplayName: "蜃境网络",                        // Display Name for your site
				RPID:          r.Host,                        // Generally the FQDN for your site
				RPOrigins:     []string{"https://" + r.Host}, //[]string{"https://" + serverURL}, // The origin URLs allowed for WebAuthn requests
			}
			webAuthor, err := webauthn.New(wconfig)
			if err != nil {
				c.doAPIResponse(w, "创建WebAuthn验证器失败", nil)
				return
			}
			c.author = webAuthor
		}
		options, session, err := c.author.BeginLogin(c.superAdmin)
		if err != nil {
			c.doAPIResponse(w, "启动超管登录失败", nil)
			return
		}
		c.authCache.Set("MirageSuperAdmin", session, 5*time.Minute)
		c.doAPIResponse(w, "", options)
	}
}

// Logout 登出
func (c *Cockpit) Logout(
	w http.ResponseWriter,
	r *http.Request,
) {
	c.authCache.Delete("AuthCode")
	authCookie := &http.Cookie{
		Name:     "mirage_cockpit_auth",
		Value:    "",
		Domain:   r.URL.Host,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, authCookie)
	c.doAPIResponse(w, "", "ok")
}

// GetServiceState 获取服务状态
func (c *Cockpit) GetServiceState(
	w http.ResponseWriter,
	r *http.Request,
) {
	c.doAPIResponse(w, "", c.serviceState)
}

// DoServiceStart 启动服务
func (c *Cockpit) DoServiceStart(
	w http.ResponseWriter,
	r *http.Request,
) {
	if c.serviceState {
		c.doAPIResponse(w, "", true)
		return
	}
	if cfg, ok := c.CheckCfgValid(); ok {
		c.serviceState = true
		c.CtrlChn <- CtrlMsg{
			Msg:    "start",
			SysCfg: cfg,
		}
		c.doAPIResponse(w, "", true)
		return
	}
	c.doAPIResponse(w, "配置不完整", false)
	return
}

// DoServiceStop 停止服务
func (c *Cockpit) DoServiceStop(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !c.serviceState {
		c.doAPIResponse(w, "", true)
		return
	}
	c.serviceState = false
	c.CtrlChn <- CtrlMsg{
		Msg: "stop",
	}
	c.doAPIResponse(w, "", false)
	return
}

// CheckCfgValid 检查配置是否有效
func (c *Cockpit) CheckCfgValid() (cfg *Config, ok bool) {
	var err error
	if sysCfg := c.GetSysCfg(); sysCfg != nil {
		cfg, err = sysCfg.toSrvConfig()
		if err != nil {
			return
		}
		if cfg.ServerURL == "" || cfg.Addr == "" || cfg.IPPrefixes == nil || cfg.BaseDomain == "" || cfg.DERPURL == "" {
			return
		}
		/*
			if cfg.OIDC.Issuer == "" || cfg.OIDC.ClientID == "" || cfg.OIDC.ClientSecret == "" {
				return
			}
				if cfg.IDaaS.App == "" || cfg.IDaaS.ClientID == "" || cfg.IDaaS.ClientKey == "" || cfg.IDaaS.Instance == "" || cfg.IDaaS.OrgID == "" {
					return
				}
				if cfg.SMS.ID == "" || cfg.SMS.Key == "" || cfg.SMS.Sign == "" || cfg.SMS.Template == "" {
					return
				}
		*/
		ok = true
	}
	return
}

// GetConfiguration 获取系统配置
func (c *Cockpit) GetSettingGeneral(
	w http.ResponseWriter,
	r *http.Request,
) {
	sysCfg := c.GetSysCfg()
	if sysCfg == nil {
		c.doAPIResponse(w, "获取系统配置失败", nil)
		return
	}
	gCfg := sysCfg.toGeneralCfg()
	c.doAPIResponse(w, "", gCfg)
}

// SetConfiguration 设置系统配置
func (c *Cockpit) SetSettingGeneral(
	w http.ResponseWriter,
	r *http.Request,
) {
	reqData := make(map[string]interface{})
	json.NewDecoder(r.Body).Decode(&reqData)
	reqState, ok := reqData["state"].(string)
	if !ok {
		c.doAPIResponse(w, "用户请求state解析失败", nil)
		return
	}
	switch reqState {
	case "set-mipv4":
		mipv4, ok := reqData["mipv4"].(string)
		if !ok {
			c.doAPIResponse(w, "用户请求mipv4解析失败", nil)
			return
		}
		mipv4Prefix, err := netip.ParsePrefix(mipv4)
		if err != nil {
			c.doAPIResponse(w, "MIPv4格式有误！", nil)
			return
		}
		mipv4Prefix, ok = netipx.RangeOfPrefix(mipv4Prefix).Prefix()
		if !ok {
			c.doAPIResponse(w, "MIPv4格式有误！", nil)
			return
		}
		if !mipv4Prefix.Addr().Is4() {
			c.doAPIResponse(w, "MIPv4格式有误！", nil)
			return
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.Mip4 = IPPrefix(mipv4Prefix)
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-mipv6":
		mipv6, ok := reqData["mipv6"].(string)
		if !ok {
			c.doAPIResponse(w, "用户请求mipv6解析失败", nil)
			return
		}
		mipv6Prefix, err := netip.ParsePrefix(mipv6)
		if err != nil {
			c.doAPIResponse(w, "MIPv6格式有误！", nil)
			return
		}
		mipv6Prefix, ok = netipx.RangeOfPrefix(mipv6Prefix).Prefix()
		if !ok {
			c.doAPIResponse(w, "MIPv6格式有误！", nil)
			return
		}
		if !mipv6Prefix.Addr().Is6() {
			c.doAPIResponse(w, "MIPv6格式有误！", nil)
			return
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.Mip6 = IPPrefix(mipv6Prefix)
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-srvaddr":
		srvaddr, ok := reqData["srvaddr"].(string)
		if !ok {
			c.doAPIResponse(w, "用户请求srvaddr解析失败", nil)
			return
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.Addr = srvaddr
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-serverurl":
		serverURL, ok := reqData["ServerURL"].(string)
		if !ok {
			c.doAPIResponse(w, "用户请求ServerURL解析失败", nil)
			return
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.ServerURL = serverURL
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-basedomain":
		baseDomain, ok := reqData["BaseDomain"].(string)
		if !ok {
			c.doAPIResponse(w, "用户请求BaseDomain解析失败", nil)
			return
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.Basedomain = baseDomain
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-derpurl":
		derpURL, ok := reqData["DERPURL"].(string)
		if !ok {
			c.doAPIResponse(w, "用户请求DERPURL解析失败", nil)
			return
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.DerpUrl = derpURL
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-routeaccessduemachine":
		SubnetAccessDueMachine, ok := reqData["SubnetAccessDueMachine"].(bool)
		if !ok {
			c.doAPIResponse(w, "用户请求SubnetAccessDueMachine解析失败", nil)
			return
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.RouteAccessDueMachine = SubnetAccessDueMachine
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-es":
		esURL, ok := reqData["ESURL"].(string)
		if !ok {
			c.doAPIResponse(w, "用户请求ESURL解析失败", nil)
			return
		}
		esKey, ok := reqData["ESKey"].(string)
		if !ok {
			c.doAPIResponse(w, "用户请求ESKey解析失败", nil)
			return
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.EsUrl = esURL
		sysCfg.EsKey = esKey
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-wxscanurl":
		wxScanURL, ok := reqData["WXScanURL"].(string)
		if !ok {
			c.doAPIResponse(w, "用户请求WXScanURL解析失败", nil)
			return
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.WXScanURL = wxScanURL
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-sms":
		smsCfgInt, ok := reqData["SMS"].(map[string]interface{})
		if !ok {
			c.doAPIResponse(w, "用户请求SMS解析失败", nil)
			return
		}
		smsCfg := SMSConfig{
			ID:       smsCfgInt["id"].(string),
			Key:      smsCfgInt["key"].(string),
			Sign:     smsCfgInt["sign"].(string),
			Template: smsCfgInt["template"].(string),
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.SMSConfig = smsCfg
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-idaas":
		idaasCfgInt, ok := reqData["IDaaS"].(map[string]interface{})
		if !ok {
			c.doAPIResponse(w, "用户请求IDaaS解析失败", nil)
			return
		}
		idaasCfg := ALIConfig{
			App:       idaasCfgInt["app"].(string),
			ClientID:  idaasCfgInt["id"].(string),
			ClientKey: idaasCfgInt["key"].(string),
			Instance:  idaasCfgInt["instance"].(string),
			OrgID:     idaasCfgInt["org"].(string),
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.IdaasConfig = idaasCfg
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-microsoft":
		MSInt, ok := reqData["Microsoft"].(map[string]interface{})
		if !ok {
			c.doAPIResponse(w, "用户请求Microsoft解析失败", nil)
			return
		}
		MSCfg := MicrosoftCfg{
			ApplicationID: MSInt["app_id"].(string),
			ClientID:      MSInt["client_id"].(string),
			ClientSecret:  MSInt["client_secret"].(string),
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.MicrosoftCfg = MSCfg
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-github":
		GHInt, ok := reqData["Github"].(map[string]interface{})
		if !ok {
			c.doAPIResponse(w, "用户请求Github解析失败", nil)
			return
		}
		GHCfg := GithubCfg{
			ClientID:     GHInt["client_id"].(string),
			ClientSecret: GHInt["client_secret"].(string),
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.GithubCfg = GHCfg
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-google":
		GgInt, ok := reqData["Google"].(map[string]interface{})
		if !ok {
			c.doAPIResponse(w, "用户请求Google解析失败", nil)
			return
		}
		GgCfg := GoogleCfg{
			ClientID:     GgInt["client_id"].(string),
			ClientSecret: GgInt["client_secret"].(string),
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.GoogleCfg = GgCfg
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	case "set-apple":
		AppInt, ok := reqData["Apple"].(map[string]interface{})
		if !ok {
			c.doAPIResponse(w, "用户请求Github解析失败", nil)
			return
		}
		AppCfg := AppleCfg{
			ClientID:   AppInt["client_id"].(string),
			TeamID:     AppInt["team_id"].(string),
			KeyID:      AppInt["key_id"].(string),
			PrivateKey: AppInt["private_key"].(string),
		}
		sysCfg := c.GetSysCfg()
		if sysCfg == nil {
			c.doAPIResponse(w, "获取系统配置失败", nil)
			return
		}
		sysCfg.AppleCfg = AppCfg
		if err := c.db.Save(sysCfg).Error; err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
	default:
		c.doAPIResponse(w, "用户请求state不存在", nil)
		return
	}

	if c.serviceState {
		newCfg, err := c.GetSysCfg().toSrvConfig()
		if err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
		c.CtrlChn <- CtrlMsg{
			Msg:    "update-config",
			SysCfg: newCfg,
		}
	}
	c.GetSettingGeneral(w, r)
}

// GenAuthCode 生成 authCode
func (c *Cockpit) GenAuthCode() string {
	const letterBytes = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 25)
	b[0] = 'm'
	b[1] = 'n'
	b[2] = '-'

	index := make([]byte, 22)
	rand.Read(index)
	for i := 0; i < 22; i++ {
		b[i+3] = letterBytes[index[i]&63]
	}
	return string(b)
}

// API调用的统一响应发报
// @msg 响应状态：成功时data不为nil则忽略，自动设置为success，否则拼接error-{msg}
// @data 响应数据：key值为data的json对象
func (c *Cockpit) doAPIResponse(writer http.ResponseWriter, msg string, data interface{}) {
	res := APIResponse{}
	if msg == "" {
		res.Status = "success"
		res.Data = data
	} else {
		res.Status = "error-" + msg
	}
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	err := json.NewEncoder(writer).Encode(&res)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

// cgao6: 用这个向前端返回错误页面
func (c *Cockpit) ErrMessage(
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

func (c *Cockpit) Run() error {
	errorGroup := new(errgroup.Group)

	router := c.createRouter()
	httpServer := &http.Server{
		Addr:         c.Addr,
		Handler:      router,
		ReadTimeout:  HTTPReadTimeout,
		WriteTimeout: 0,
	}
	httpListener, err := net.Listen("tcp", c.Addr)
	if err != nil {
		return fmt.Errorf("failed to bind to TCP address: %w", err)
	}

	errorGroup.Go(func() error { return httpServer.Serve(httpListener) })

	msgFunc := func(c *Cockpit) {
		for {
			msg := <-c.MsgChn
			switch msg.Msg {
			case "error":
				log.Info().Msg("received service fatal error message, should be stopped")
				c.serviceState = false
			}
		}
	}
	errorGroup.Go(func() error {
		msgFunc(c)
		return nil
	})

	if sysCfg, ok := c.CheckCfgValid(); ok {
		c.serviceState = true
		// 启动定时任务
		c.CtrlChn <- CtrlMsg{
			Msg:    "start",
			Err:    nil,
			SysCfg: sysCfg,
		}
	}

	return errorGroup.Wait()
}

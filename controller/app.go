package controller

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/dexidp/dex/server"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/puzpuzpuz/xsync/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"tailscale.com/control/controlclient"
	"tailscale.com/tailcfg"
	"tailscale.com/types/key"
)

const (
	//	NoiseKeyPath = "noise.key"
	DatabasePath = "db.sqlite"
	//	DexDBPath    = "db.sqlite"
	//	DexDBType    = "sqlite3"
	AuthPrefix = "Bearer "

	EphemeralNodeInactivityTimeout = 5 * time.Minute  //不得低于65s
	NodeUpdateCheckInterval        = 10 * time.Second //不得大于60s
	updateInterval                 = 5000
	HTTPReadTimeout                = 30 * time.Second
	HTTPShutdownTimeout            = 3 * time.Second
	privateKeyFileMode             = 0o600

	smsCacheExpiration = time.Minute * 5
	smsCacheCleanup    = time.Minute * 5
)

// Mirage represents the base app of the service.
type Mirage struct {
	cfg    *Config
	db     *gorm.DB
	ctx    context.Context
	cancel context.CancelFunc

	noisePrivateKey *key.MachinePrivate
	//	DERPMap         *tailcfg.DERPMap
	DERPNCs    map[string]*controlclient.NoiseClient
	DERPseqnum map[string]int

	aclPolicy *ACLPolicy
	aclRules  []tailcfg.FilterRule
	sshPolicy *tailcfg.SSHPolicy

	lastStateChange *xsync.MapOf[string, time.Time]

	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config

	smsCodeCache *cache.Cache

	aCodeCache              *cache.Cache
	stateCodeCache          *cache.Cache
	controlCodeCache        *cache.Cache
	machineControlCodeCache *cache.Cache
	//organizationCache       *cache.Cache

	tcdCache *cache.Cache

	longPollChanPool map[string]chan string

	ipAllocationMutex sync.Mutex

	shutdownChan       chan struct{}
	pollNetMapStreamWG sync.WaitGroup
}

func NewMirage(cfg *Config, db *gorm.DB) (*Mirage, error) {
	// noisePrivateKey, err := readOrCreatePrivateKey(AbsolutePathFromConfigPath(NoiseKeyPath))
	noisePrivateKey, err := getServerPrivateKey(db)
	if err != nil {
		return nil, fmt.Errorf("failed to read or create Noise protocol private key: %w", err)
	}

	//cgao6: 注册机制探索
	aCodeCache := cache.New(0, 0)
	stateCodeCache := cache.New(0, 0)
	controlCodeCache := cache.New(0, 0)
	machineControlCodeCache := cache.New(0, 0)

	smsCodeCache := cache.New(
		smsCacheExpiration,
		smsCacheCleanup,
	)
	longPollChanPool := make(map[string]chan string, 0)

	InitESLogger(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	app := Mirage{
		cfg:    cfg,
		db:     db,
		ctx:    ctx,
		cancel: cancel,

		noisePrivateKey: noisePrivateKey,
		DERPNCs:         make(map[string]*controlclient.NoiseClient),
		DERPseqnum:      make(map[string]int),
		aclRules:        tailcfg.FilterAllowAll, // default allowall

		aCodeCache:              aCodeCache,
		stateCodeCache:          stateCodeCache,
		controlCodeCache:        controlCodeCache,
		machineControlCodeCache: machineControlCodeCache,
		tcdCache:                cache.New(0, 0),
		longPollChanPool:        longPollChanPool,
		smsCodeCache:            smsCodeCache,
		shutdownChan:            make(chan struct{}),
		pollNetMapStreamWG:      sync.WaitGroup{},
		lastStateChange:         xsync.NewMapOf[time.Time](),
	}

	nrs := app.ListNaviRegions()
	for _, nr := range nrs {
		nns := app.ListNaviNodes(nr.ID)
		for _, nn := range nns {
			nckey := key.MachinePublic{}
			nckey.UnmarshalText([]byte(MachinePublicKeyEnsurePrefix(nn.NaviKey)))
			nc, err := app.GetNaviNoiseClient(nckey, nn.HostName, nn.DERPPort)
			if err != nil {
				log.Error().Err(err).Msg("GetNaviNoiseClient Error: " + err.Error())
			}
			app.DERPNCs[nn.ID] = nc
			app.DERPseqnum[nn.ID] = 0
		}
	}

	/* 由于可能我们会使用内建的dex，所以这里可能并不能正确初始化OIDC
	if cfg.OIDC.Issuer != "" {
		err = app.initOIDC()
		if err != nil {
			log.Warn().Err(err).Msg("failed to set up OIDC provider, falling back to CLI based authentication")
		}
	}
	*/

	return &app, nil
}

// expireEphemeralNodes deletes ephemeral machine records that have not been
// seen for longer than h.cfg.EphemeralNodeInactivityTimeout.
func (h *Mirage) expireEphemeralNodes(ticker *time.Ticker) { //milliSeconds int64) {
	//ticker := time.NewTicker(time.Duration(milliSeconds) * time.Millisecond)
	for range ticker.C {
		h.expireEphemeralNodesWorker()
	}
}

// expireExpiredMachines expires machines that have an explicit expiry set
// after that expiry time has passed.
func (h *Mirage) expireExpiredMachines(ticker *time.Ticker) { //milliSeconds int64) {
	//ticker := time.NewTicker(time.Duration(milliSeconds) * time.Millisecond)
	for range ticker.C {
		h.expireExpiredMachinesWorker()
	}
}

func (h *Mirage) failoverSubnetRoutes(ticker *time.Ticker) { //milliSeconds int64) {
	//ticker := time.NewTicker(time.Duration(milliSeconds) * time.Millisecond)
	for range ticker.C {
		err := h.handlePrimarySubnetFailover()
		if err != nil {
			log.Error().Err(err).Msg("failed to handle primary subnet failover")
		}
	}
}

func (h *Mirage) expireEphemeralNodesWorker() {
	users, err := h.ListUsers()
	if err != nil {
		log.Error().Err(err).Msg("Error listing users")

		return
	}

	for _, user := range users {
		machines, err := h.ListMachinesByUser(user.ID)
		if err != nil {
			log.Error().
				Err(err).
				Str("user", user.Name).
				Msg("Error listing machines in user")

			return
		}

		expiredFound := false
		for _, machine := range machines {
			if machine.isEphemeral() && machine.LastSeen != nil &&
				time.Now().
					After(machine.LastSeen.Add(EphemeralNodeInactivityTimeout)) {
				expiredFound = true
				log.Info().
					Str("machine", machine.Hostname).
					Msg("Ephemeral client removed from database")

				err = h.db.Unscoped().Delete(machine).Error
				if err != nil {
					log.Error().
						Err(err).
						Str("machine", machine.Hostname).
						Msg("🤮 Cannot delete ephemeral machine from the database")
				}

				// TODO: 并不好的处理
				h.NotifyNaviOrgNodesChange(machine.User.OrganizationID, "", machine.NodeKey)
			}
		}

		if expiredFound {
			h.setOrgLastStateChangeToNow(user.OrganizationID)
		}
	}
}

func (h *Mirage) expireExpiredMachinesWorker() {
	orgs, err := h.ListOrgnaizations()
	if err != nil {
		log.Error().Err(err).Msg("Error listing organizations")

		return
	}
	orgChangeSet := NewUtilsSet[int64]()
	for _, org := range orgs {
		users, err := h.ListOrgUsers(org.ID)
		if err != nil {
			log.Error().Err(err).Msg("Error listing users")

			return
		}
		for _, user := range users {
			machines, err := h.ListMachinesByUser(user.ID)
			if err != nil {
				log.Error().
					Err(err).
					Str("user", user.Name).
					Msg("Error listing machines in user")

				return
			}

			for index, machine := range machines {
				if machine.isExpired() &&
					machine.Expiry.After(h.getOrgLastStateChange(user.OrganizationID)) {
					orgChangeSet.SetKey(org.ID)

					err := h.ExpireMachine(&machines[index])
					if err != nil {
						log.Error().
							Err(err).
							Str("machine", machine.Hostname).
							Str("name", machine.GivenName).
							Msg("🤮 Cannot expire machine")
					} else {
						log.Info().
							Str("machine", machine.Hostname).
							Str("name", machine.GivenName).
							Msg("Machine successfully expired")
					}

					h.NotifyNaviOrgNodesChange(machine.User.OrganizationID, "", machine.NodeKey)
				}
			}
		}
	}
	changedOrgList := orgChangeSet.GetKeys()
	if len(changedOrgList) > 0 {
		h.setOrgLastStateChangeToNow(changedOrgList...)
	}
}

//go:embed console_html/admin
var adminFS embed.FS

//go:embed console_html
var mainpageFS embed.FS

//go:embed console_html/login
var loginFS embed.FS

//go:embed console_html/downloads
var downloadsFS embed.FS

func (h *Mirage) initRouter(router *mux.Router) {

	adminDir, err := fs.Sub(adminFS, "console_html/admin")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	loginDir, err := fs.Sub(loginFS, "console_html/login")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	mainpageDir, err := fs.Sub(mainpageFS, "console_html")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	downloadsDir, err := fs.Sub(downloadsFS, "console_html/downloads")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	router.PathPrefix("/downloads").Handler(http.StripPrefix("/downloads", http.FileServer(http.FS(downloadsDir))))

	router.PathPrefix("/download").Handler(http.StripPrefix("/download", http.FileServer(http.Dir("download"))))

	router.HandleFunc("/logout", h.ConsoleLogout).Methods(http.MethodGet)
	//注册
	router.PathPrefix("/api/register").HandlerFunc(h.RegisterUserAPI).Methods(http.MethodPost)
	router.PathPrefix("/api/idps").HandlerFunc(h.ListIdps).Methods(http.MethodGet)

	//登录
	router.PathPrefix("/login").HandlerFunc(h.doLogin).Methods(http.MethodPost)
	router.PathPrefix("/wxmini").HandlerFunc(h.checkWXMini).Methods(http.MethodPost)
	login_router := router.PathPrefix("/login").Subrouter()
	login_router.Use(h.loginMidware)
	login_router.PathPrefix("").Handler(http.StripPrefix("/login", http.FileServer(http.FS(loginDir))))

	//cgao6: APage也算是控制台中的一环，逻辑类似
	//对于特殊路径"/a/oauth_response"是login到第三方后验证通过的回写token逻辑
	router.HandleFunc("/a/oauth_response", h.oauthResponse).Methods(http.MethodGet)
	router.HandleFunc("/a/{aCode}", h.deviceRegPortal).Methods(http.MethodGet)
	router.HandleFunc("/a/oauth_response", h.selectOrgForLogin).Methods(http.MethodPost)
	router.HandleFunc("/a/{aCode}", h.deviceReg).Methods(http.MethodPost)

	// 控制台所需的全部API接口（由APIAuth身份验证放行）
	api_router := router.PathPrefix("/admin/api").Subrouter()
	api_router.Use(h.APIAuth)

	// 控制台页面全部路由（由ConsoleAuth身份验证放行），与前一个和最后面的juanfont的API接口之后要进行统一
	console_router := router.PathPrefix("/admin").Subrouter()
	console_router.Use(h.ConsoleAuth)

	// GET(查询类)API
	console_router.HandleFunc("/api/self", h.ConsoleSelfAPI).Methods(http.MethodGet)
	console_router.HandleFunc("/api/users", h.CAPIGetUsers).Methods(http.MethodGet)
	console_router.HandleFunc("/api/machines", h.ConsoleMachinesAPI).Methods(http.MethodGet)
	console_router.HandleFunc("/api/machine-debug", h.ConsoleMachineDebugAPI).Methods(http.MethodGet)
	console_router.HandleFunc("/api/dns", h.CAPIGetDNS).Methods(http.MethodGet)
	console_router.HandleFunc("/api/tcd/offers", h.CAPIGetTCDOffers).Methods(http.MethodGet)
	console_router.HandleFunc("/api/netsettings", h.getNetSettingAPI).Methods(http.MethodGet)
	console_router.HandleFunc("/api/keys", h.CAPIGetKeys).Methods(http.MethodGet)
	console_router.HandleFunc("/api/acls/tags", h.CAPIGetTags).Methods(http.MethodGet)
	console_router.HandleFunc("/api/subscription", h.CAPIGetSubscription).Methods(http.MethodGet)
	console_router.HandleFunc("/api/derp/query", h.CAPIQueryDERP).Methods(http.MethodGet)

	// POST(更新类)API
	console_router.HandleFunc("/api/users", h.CAPIPostUsers).Methods(http.MethodPost)
	console_router.HandleFunc("/api/machines", h.ConsoleMachinesUpdateAPI).Methods(http.MethodPost)
	console_router.HandleFunc("/api/machine/remove", h.ConsoleRemoveMachineAPI).Methods(http.MethodPost)
	console_router.HandleFunc("/api/netsetting/updatekeyexpiry", h.ConsoleUpdateKeyExpiryAPI).Methods(http.MethodPost)
	console_router.HandleFunc("/api/keys", h.CAPIPostKeys).Methods(http.MethodPost)
	console_router.HandleFunc("/api/acls/tags", h.CAPIPostTags).Methods(http.MethodPost)
	console_router.HandleFunc("/api/dns", h.CAPIPostDNS).Methods(http.MethodPost)
	console_router.HandleFunc("/api/tcd", h.CAPIPostTCD).Methods(http.MethodPost)
	console_router.HandleFunc("/api/derp/add", h.CAPIAddDERP).Methods(http.MethodPost)
	console_router.HandleFunc("/api/derp/ban/{id}", h.CAPISwitchRegionBan).Methods(http.MethodPost)

	// DELETE(删除类)API
	console_router.PathPrefix("/api/keys/").HandlerFunc(h.CAPIDelKeys).Methods(http.MethodDelete)
	console_router.PathPrefix("/api/acls/tags/").HandlerFunc(h.CAPIDelTags).Methods(http.MethodDelete)
	console_router.PathPrefix("/api/derp/{id}").HandlerFunc(h.CAPIDelNaviNode).Methods(http.MethodDelete)

	// TODO: 登出及页面转至VUE，要考虑logout是否有必要发消息给服务端
	//cgao6: 改成不需检查登录信息	console_router.HandleFunc("/logout", h.ConsoleLogout).Methods(http.MethodGet)
	console_router.PathPrefix("").Handler(http.StripPrefix("/admin", http.FileServer(http.FS(adminDir))))

	// 核心与客户端通信协议，不动
	router.HandleFunc("/ts2021", h.NoiseUpgradeHandler).Methods(http.MethodPost)
	router.HandleFunc("/key", h.KeyHandler).Methods(http.MethodGet)

	// 资源目录们
	router.PathPrefix("/img/").Handler(http.StripPrefix("/", http.FileServer(http.FS(mainpageDir))))
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/", http.FileServer(http.FS(mainpageDir))))
	// 其余全部默认返回主页
	router.Path("/").Handler(http.FileServer(http.FS(mainpageDir)))

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ErrMessage(w, r, 404, "你迷失在蜃境中了吗？这里什么都没有")
	})
}

// Serve launches a GIN server with the Mirage API.
func (h *Mirage) Serve(ctrlChn chan CtrlMsg) error {
	var err error

	// Fetch an initial DERP Map before we start serving
	//	h.DERPMap, err = h.LoadDERPMapFromURL(h.cfg.DERPURL)
	//	if err != nil {
	//		return err
	//	}

	ticker := time.NewTicker(time.Millisecond * updateInterval)
	defer ticker.Stop()

	go h.expireEphemeralNodes(ticker)  //updateInterval)
	go h.expireExpiredMachines(ticker) //updateInterval)
	go h.failoverSubnetRoutes(ticker)  //updateInterval)
	go h.refreshNaviStatusPoller(ticker)

	// Prepare group for running listeners
	errorGroup := new(errgroup.Group)

	//
	//
	// HTTP setup
	//
	// This is the regular router that we expose
	// over our main Addr. It also serves the legacy Tailcale API
	router := mux.NewRouter()

	_, err = server.InitDexServer(h.ctx, *h.cfg.DexConfig, router) //cgao6: 这里是dex的初始化
	if err != nil {
		return err
	}
	defer h.cfg.DexConfig.Storage.Close()

	h.initRouter(router)

	httpServer := &http.Server{
		Addr:        h.cfg.Addr,
		Handler:     router,
		ReadTimeout: HTTPReadTimeout,
		// Go does not handle timeouts in HTTP very well, and there is
		// no good way to handle streaming timeouts, therefore we need to
		// keep this at unlimited and be careful to clean up connections
		// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/#aboutstreaming
		WriteTimeout: 0,
	}

	var httpListener net.Listener
	httpListener, err = net.Listen("tcp", h.cfg.Addr)
	if err != nil {
		return fmt.Errorf("failed to bind to TCP address: %w", err)
	}

	errorGroup.Go(func() error { return httpServer.Serve(httpListener) })

	log.Info().
		Msgf("listening and serving HTTP on: %s", h.cfg.Addr)

	if err != nil {
		return fmt.Errorf("failed to bind to TCP address: %w", err)
	}

	ctrlFunc := func(c chan CtrlMsg) {
		for {
			msg := <-c
			switch msg.Msg {
			case "stop":
				log.Info().Msg("Received stop message, shutting down")
				close(h.shutdownChan)
				h.pollNetMapStreamWG.Wait()
				// Gracefully shut down servers
				ctx, cancel := context.WithTimeout(
					context.Background(),
					HTTPShutdownTimeout,
				)
				// Shutdown http server
				if err := httpServer.Shutdown(ctx); err != nil {
					log.Error().Err(err).Msg("Failed to shutdown http")
				}
				// Close network listeners
				err = httpListener.Close()
				if err != nil {
					log.Error().Err(err).Msg("Failed to close http listener")
				}

				h.cancel() // ??
				/*
					// Close db connections
					db, err := h.db.DB()
					if err != nil {
						log.Error().Err(err).Msg("Failed to get db handle")
					}
					err = db.Close()
					if err != nil {
						log.Error().Err(err).Msg("Failed to close db")
					}
				*/
				log.Info().Msg("Mirage stopped")
				cancel()
				return
			case "update-config":
				log.Info().Msg("Received update-config message, updating config")
				h.cfg = msg.SysCfg
			}
		}
	}
	errorGroup.Go(func() error {
		ctrlFunc(ctrlChn)
		return nil
	})

	/*
		// Handle common process-killing signals so we can gracefully shut down:
		h.shutdownChan = make(chan struct{})
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGHUP)
		sigFunc := func(c chan os.Signal) {
			// Wait for a SIGINT or SIGKILL:
			for {
				sig := <-c
				switch sig {
				case syscall.SIGHUP:
					log.Info().
						Str("signal", sig.String()).
						Msg("Received SIGHUP, reloading ACL and Config")

						// TODO(kradalby): Reload config on SIGHUP

					aclPath := AbsolutePathFromConfigPath(ACLPath)
					err := h.LoadACLPolicy(aclPath)
					if err != nil {
						log.Error().Err(err).Msg("Failed to reload ACL policy")
					}
					log.Info().
						Str("path", aclPath).
						Msg("ACL policy successfully reloaded, notifying nodes of change")

					h.setLastStateChangeToNow()
				default:
					log.Info().
						Str("signal", sig.String()).
						Msg("Received signal to stop, shutting down gracefully")

					close(h.shutdownChan)
					h.pollNetMapStreamWG.Wait()

					// Gracefully shut down servers
					ctx, cancel := context.WithTimeout(
						context.Background(),
						HTTPShutdownTimeout,
					)
					if err := httpServer.Shutdown(ctx); err != nil {
						log.Error().Err(err).Msg("Failed to shutdown http")
					}

					// Close network listeners
					httpListener.Close()

					// Close db connections
					db, err := h.db.DB()
					if err != nil {
						log.Error().Err(err).Msg("Failed to get db handle")
					}
					err = db.Close()
					if err != nil {
						log.Error().Err(err).Msg("Failed to close db")
					}

					log.Info().
						Msg("Mirage stopped")

					// And we're done:
					cancel()
					os.Exit(0)
				}
			}
		}
		errorGroup.Go(func() error {
			sigFunc(sigc)

			return nil
		})
	*/

	return errorGroup.Wait()
}

func (h *Mirage) setLastStateChangeToNow() {
	var err error

	now := time.Now().UTC()

	users, err := h.ListUsers()
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("failed to fetch all users, failing to update last changed state.")
	}

	for _, user := range users {
		if h.lastStateChange == nil {
			h.lastStateChange = xsync.NewMapOf[time.Time]()
		}
		h.lastStateChange.Store(user.Name, now)
	}
}

func (h *Mirage) setOrgLastStateChangeToNow(orgId ...int64) {
	var err error
	var users []User
	now := time.Now().UTC()

	if len(orgId) == 1 {
		users, err = h.ListOrgUsers(orgId[0])
	} else {
		users, err = h.ListUsersInOrgs(orgId)
	}
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("failed to fetch organization users, failing to update last changed state.")
	}

	for _, user := range users {
		if h.lastStateChange == nil {
			h.lastStateChange = xsync.NewMapOf[time.Time]()
		}
		h.lastStateChange.Store(user.StableID, now)
	}
}

func (h *Mirage) getLastStateChange(users ...User) time.Time {
	times := []time.Time{}

	// getLastStateChange takes a list of users as a "filter", if no users
	// are past, then use the entier list of users and look for the last update
	if len(users) > 0 {
		for _, user := range users {
			if lastChange, ok := h.lastStateChange.Load(user.Name); ok {
				times = append(times, lastChange)
			}
		}
	} else {
		h.lastStateChange.Range(func(key string, value time.Time) bool {
			times = append(times, value)

			return true
		})
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i].After(times[j])
	})

	if len(times) == 0 {
		return time.Now().UTC()
	} else {
		return times[0]
	}
}

func (h *Mirage) getOrgLastStateChange(orgId int64) time.Time {
	times := []time.Time{}

	// getLastStateChange takes a list of users as a "filter", if no users
	// are past, then use the entier list of users and look for the last update
	users, err := h.ListOrgUsers(orgId)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("failed to fetch organization users, failing to get last changed state.")
	}
	for _, user := range users {
		if lastChange, ok := h.lastStateChange.Load(user.StableID); ok {
			times = append(times, lastChange)
		}
	}

	if len(times) == 0 {
		return time.Now().UTC()
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i].After(times[j])
	})

	return times[0]
}

func stdoutHandler(
	writer http.ResponseWriter,
	req *http.Request,
) {
	body, _ := io.ReadAll(req.Body)

	log.Trace().
		Interface("header", req.Header).
		Interface("proto", req.Proto).
		Interface("url", req.URL).
		Bytes("body", body).
		Msg("Request did not match")
}

func getServerPrivateKey(db *gorm.DB) (*key.MachinePrivate, error) {
	var sysCfg SysConfig
	err := db.First(&sysCfg).Error
	if err != nil || sysCfg.ServerKey == "" {
		return nil, fmt.Errorf("failed to get server private key: %w", err)
	}

	var machineKey key.MachinePrivate
	if err = machineKey.UnmarshalText([]byte(sysCfg.ServerKey)); err != nil {
		log.Error().
			Caller().
			Msg("Convert db server key string to machinekey failed: " + err.Error())
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return &machineKey, nil
}

/*
func readOrCreatePrivateKey(path string) (*key.MachinePrivate, error) {
	privateKey, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		log.Info().Str("path", path).Msg("No private key file at path, creating...")

		machineKey := key.NewMachine()

		machineKeyStr, err := machineKey.MarshalText()
		if err != nil {
			return nil, fmt.Errorf(
				"failed to convert private key to string for saving: %w",
				err,
			)
		}
		err = os.WriteFile(path, machineKeyStr, privateKeyFileMode)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to save private key to disk: %w",
				err,
			)
		}

		return &machineKey, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	trimmedPrivateKey := strings.TrimSpace(string(privateKey))
	privateKeyEnsurePrefix := PrivateKeyEnsurePrefix(trimmedPrivateKey)

	var machineKey key.MachinePrivate
	if err = machineKey.UnmarshalText([]byte(privateKeyEnsurePrefix)); err != nil {
		log.Info().
			Str("path", path).
			Msg("This might be due to a legacy (mirage pre-0.12) private key. " +
				"If the key is in WireGuard format, delete the key and restart mirage. " +
				"A new key will automatically be generated. All Tailscale clients will have to be restarted")

		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &machineKey, nil
}
*/

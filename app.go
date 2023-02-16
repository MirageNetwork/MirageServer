package Mirage

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/puzpuzpuz/xsync/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"tailscale.com/control/controlbase"
	"tailscale.com/tailcfg"
	"tailscale.com/types/dnstype"
	"tailscale.com/types/key"
)

const (
	errSTUNAddressNotSet                   = Error("STUN address not set")
	errUnsupportedDatabase                 = Error("unsupported DB")
	errUnsupportedLetsEncryptChallengeType = Error(
		"unknown value for Lets Encrypt challenge type",
	)
)

const (
	AuthPrefix          = "Bearer "
	Postgres            = "postgres"
	Sqlite              = "sqlite3"
	updateInterval      = 5000
	HTTPReadTimeout     = 30 * time.Second
	HTTPShutdownTimeout = 3 * time.Second
	privateKeyFileMode  = 0o600

	registerCacheExpiration = time.Minute * 15
	registerCacheCleanup    = time.Minute * 20

	smsCacheExpiration = time.Minute * 5
	smsCacheCleanup    = time.Minute * 5

	DisabledClientAuth = "disabled"
	RelaxedClientAuth  = "relaxed"
	EnforcedClientAuth = "enforced"
)

// Mirage represents the base app of the service.
type Mirage struct {
	cfg             *Config
	db              *gorm.DB
	dbString        string
	dbType          string
	dbDebug         bool
	noisePrivateKey *key.MachinePrivate

	noiseConn *controlbase.Conn

	DERPMap *tailcfg.DERPMap

	aclPolicy *ACLPolicy
	aclRules  []tailcfg.FilterRule
	sshPolicy *tailcfg.SSHPolicy

	lastStateChange *xsync.MapOf[string, time.Time]

	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config

	smsCodeCache *cache.Cache

	registrationCache       *cache.Cache
	aCodeCache              *cache.Cache
	stateCodeCache          *cache.Cache
	controlCodeCache        *cache.Cache
	machineControlCodeCache *cache.Cache

	longPollChanPool map[string]chan string

	ipAllocationMutex sync.Mutex

	shutdownChan       chan struct{}
	pollNetMapStreamWG sync.WaitGroup
}

func NewMirage(cfg *Config) (*Mirage, error) {
	noisePrivateKey, err := readOrCreatePrivateKey(cfg.NoisePrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read or create Noise protocol private key: %w", err)
	}

	var dbString string
	switch cfg.DBtype {
	case Postgres:
		dbString = fmt.Sprintf(
			"host=%s dbname=%s user=%s",
			cfg.DBhost,
			cfg.DBname,
			cfg.DBuser,
		)

		if sslEnabled, err := strconv.ParseBool(cfg.DBssl); err == nil {
			if !sslEnabled {
				dbString += " sslmode=disable"
			}
		} else {
			dbString += fmt.Sprintf(" sslmode=%s", cfg.DBssl)
		}

		if cfg.DBport != 0 {
			dbString += fmt.Sprintf(" port=%d", cfg.DBport)
		}

		if cfg.DBpass != "" {
			dbString += fmt.Sprintf(" password=%s", cfg.DBpass)
		}
	case Sqlite:
		dbString = cfg.DBpath
	default:
		return nil, errUnsupportedDatabase
	}

	registrationCache := cache.New(
		registerCacheExpiration,
		registerCacheCleanup,
	)

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

	app := Mirage{
		cfg:             cfg,
		dbType:          cfg.DBtype,
		dbString:        dbString,
		noisePrivateKey: noisePrivateKey,
		aclRules:        tailcfg.FilterAllowAll, // default allowall

		aCodeCache:              aCodeCache,
		stateCodeCache:          stateCodeCache,
		controlCodeCache:        controlCodeCache,
		machineControlCodeCache: machineControlCodeCache,
		longPollChanPool:        longPollChanPool,
		smsCodeCache:            smsCodeCache,
		registrationCache:       registrationCache,
		pollNetMapStreamWG:      sync.WaitGroup{},
		lastStateChange:         xsync.NewMapOf[time.Time](),
	}

	err = app.initDB()
	if err != nil {
		return nil, err
	}

	if cfg.OIDC.Issuer != "" {
		err = app.initOIDC()
		if err != nil {
			if cfg.OIDC.OnlyStartIfOIDCIsAvailable {
				return nil, err
			} else {
				log.Warn().Err(err).Msg("failed to set up OIDC provider, falling back to CLI based authentication")
			}
		}
	}

	if app.cfg.DNSConfig != nil && app.cfg.DNSConfig.Proxied { // if MagicDNS
		magicDNSDomains := generateMagicDNSRootDomains(app.cfg.IPPrefixes)
		// we might have routes already from Split DNS
		if app.cfg.DNSConfig.Routes == nil {
			app.cfg.DNSConfig.Routes = make(map[string][]*dnstype.Resolver)
		}
		for _, d := range magicDNSDomains {
			app.cfg.DNSConfig.Routes[d.WithoutTrailingDot()] = nil
		}
	}

	return &app, nil
}

// expireEphemeralNodes deletes ephemeral machine records that have not been
// seen for longer than h.cfg.EphemeralNodeInactivityTimeout.
func (h *Mirage) expireEphemeralNodes(milliSeconds int64) {
	ticker := time.NewTicker(time.Duration(milliSeconds) * time.Millisecond)
	for range ticker.C {
		h.expireEphemeralNodesWorker()
	}
}

// expireExpiredMachines expires machines that have an explicit expiry set
// after that expiry time has passed.
func (h *Mirage) expireExpiredMachines(milliSeconds int64) {
	ticker := time.NewTicker(time.Duration(milliSeconds) * time.Millisecond)
	for range ticker.C {
		h.expireExpiredMachinesWorker()
	}
}

func (h *Mirage) failoverSubnetRoutes(milliSeconds int64) {
	ticker := time.NewTicker(time.Duration(milliSeconds) * time.Millisecond)
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
		machines, err := h.ListMachinesByUser(user.Name)
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
					After(machine.LastSeen.Add(h.cfg.EphemeralNodeInactivityTimeout)) {
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
			}
		}

		if expiredFound {
			h.setLastStateChangeToNow()
		}
	}
}

func (h *Mirage) expireExpiredMachinesWorker() {
	users, err := h.ListUsers()
	if err != nil {
		log.Error().Err(err).Msg("Error listing users")

		return
	}

	for _, user := range users {
		machines, err := h.ListMachinesByUser(user.Name)
		if err != nil {
			log.Error().
				Err(err).
				Str("user", user.Name).
				Msg("Error listing machines in user")

			return
		}

		expiredFound := false
		for index, machine := range machines {
			if machine.isExpired() &&
				machine.Expiry.After(h.getLastStateChange(user)) {
				expiredFound = true

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
			}
		}

		if expiredFound {
			h.setLastStateChangeToNow()
		}
	}
}

//go:embed console/admin
var adminFS embed.FS

//go:embed console
var mainpageFS embed.FS

//go:embed console/login
var loginFS embed.FS

func (h *Mirage) createRouter() *mux.Router {
	router := mux.NewRouter()

	adminDir, err := fs.Sub(adminFS, "console/admin")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	loginDir, err := fs.Sub(loginFS, "console/login")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	mainpageDir, err := fs.Sub(mainpageFS, "console")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	//注册
	router.PathPrefix("/api/register").HandlerFunc(h.RegisterUserAPI).Methods(http.MethodPost)

	//登录
	router.PathPrefix("/login").HandlerFunc(h.doLogin).Methods(http.MethodPost)
	login_router := router.PathPrefix("/login").Subrouter()
	login_router.Use(h.loginMidware)
	login_router.PathPrefix("").Handler(http.StripPrefix("/login", http.FileServer(http.FS(loginDir))))

	//cgao6: APage也算是控制台中的一环，逻辑类似
	//对于特殊路径"/a/oauth_response"是login到第三方后验证通过的回写token逻辑
	router.HandleFunc("/a/oauth_response", h.oauthResponse).Methods(http.MethodGet)
	router.HandleFunc("/a/{aCode}", h.deviceRegPortal).Methods(http.MethodGet)
	router.HandleFunc("/a/{aCode}", h.deviceReg).Methods(http.MethodPost)

	// 控制台所需的全部API接口（由APIAuth身份验证放行）
	api_router := router.PathPrefix("/admin/api").Subrouter()
	api_router.Use(h.APIAuth)

	// 控制台页面全部路由（由ConsoleAuth身份验证放行），与前一个和最后面的juanfont的API接口之后要进行统一
	console_router := router.PathPrefix("/admin").Subrouter()
	console_router.Use(h.ConsoleAuth)

	// GET(查询类)API
	console_router.HandleFunc("/api/self", h.ConsoleSelfAPI).Methods(http.MethodGet)
	console_router.HandleFunc("/api/machines", h.ConsoleMachinesAPI).Methods(http.MethodGet)
	console_router.HandleFunc("/api/dns", h.CAPIGetDNS).Methods(http.MethodGet)
	console_router.HandleFunc("/api/netsettings", h.getNetSettingAPI).Methods(http.MethodGet)
	console_router.HandleFunc("/api/keys", h.CAPIGetKeys).Methods(http.MethodGet)

	// POST(更新类)API
	console_router.HandleFunc("/api/machines", h.ConsoleMachinesUpdateAPI).Methods(http.MethodPost)
	console_router.HandleFunc("/api/machine/remove", h.ConsoleRemoveMachineAPI).Methods(http.MethodPost)
	console_router.HandleFunc("/api/netsetting/updatekeyexpiry", h.ConsoleUpdateKeyExpiryAPI).Methods(http.MethodPost)
	console_router.HandleFunc("/api/keys", h.CAPIPostKeys).Methods(http.MethodPost)
	console_router.HandleFunc("/api/dns", h.CAPIPostDNS).Methods(http.MethodPost)

	// DELETE(删除类)API
	console_router.PathPrefix("/api/keys/").HandlerFunc(h.CAPIDelKeys).Methods(http.MethodDelete)

	// TODO: 登出及页面转至VUE，要考虑logout是否有必要发消息给服务端
	console_router.HandleFunc("/logout", h.ConsoleLogout).Methods(http.MethodGet)
	console_router.PathPrefix("").Handler(http.StripPrefix("/admin", http.FileServer(http.FS(adminDir))))

	// 核心与客户端通信协议，不动
	router.HandleFunc("/ts2021", h.NoiseUpgradeHandler).Methods(http.MethodPost)
	router.HandleFunc("/key", h.KeyHandler).Methods(http.MethodGet)

	// TODO: OIDC 注册部分，已经部分取代，亟需清理
	router.HandleFunc("/oidc/register/{nkey}", h.RegisterOIDC).Methods(http.MethodGet)
	router.HandleFunc("/oidc/callback", h.OIDCCallback).Methods(http.MethodGet)

	// 资源目录们
	router.PathPrefix("/img/").Handler(http.StripPrefix("/", http.FileServer(http.FS(mainpageDir))))
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/", http.FileServer(http.FS(mainpageDir))))
	// 其余全部默认返回主页
	router.Path("/").Handler(http.FileServer(http.FS(mainpageDir)))

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ErrMessage(w, r, 404, "你迷失在蜃境中了吗？这里什么都没有")
	})
	return router
}

// Serve launches a GIN server with the Mirage API.
func (h *Mirage) Serve() error {
	var err error

	// Fetch an initial DERP Map before we start serving
	h.DERPMap = GetDERPMap(h.cfg.DERP)

	if h.cfg.DERP.AutoUpdate {
		derpMapCancelChannel := make(chan struct{})
		defer func() { derpMapCancelChannel <- struct{}{} }()
		go h.scheduledDERPMapUpdateWorker(derpMapCancelChannel)
	}

	go h.expireEphemeralNodes(updateInterval)
	go h.expireExpiredMachines(updateInterval)

	go h.failoverSubnetRoutes(updateInterval)

	// Prepare group for running listeners
	errorGroup := new(errgroup.Group)

	//
	//
	// HTTP setup
	//
	// This is the regular router that we expose
	// over our main Addr. It also serves the legacy Tailcale API
	router := h.createRouter()

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

				if h.cfg.ACL.PolicyPath != "" {
					aclPath := AbsolutePathFromConfigPath(h.cfg.ACL.PolicyPath)
					err := h.LoadACLPolicy(aclPath)
					if err != nil {
						log.Error().Err(err).Msg("Failed to reload ACL policy")
					}
					log.Info().
						Str("path", aclPath).
						Msg("ACL policy successfully reloaded, notifying nodes of change")

					h.setLastStateChangeToNow()
				}

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

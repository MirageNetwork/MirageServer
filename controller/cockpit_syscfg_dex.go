package controller

import (
	"fmt"
	"time"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/dexidp/dex/connector/apple"
	"github.com/dexidp/dex/connector/github"
	"github.com/dexidp/dex/connector/google"
	"github.com/dexidp/dex/connector/microsoft"
	"github.com/dexidp/dex/server"
	dexStorage "github.com/dexidp/dex/storage"
	dexSQL "github.com/dexidp/dex/storage/sql"
)

func (s *SysConfig) toDexConfig() (*server.Config, error) {
	storageCfg := DexStorage{
		Type: DexDBType,
		Config: &dexSQL.SQLite3{
			File: AbsolutePathFromConfigPath(DexDBPath),
		},
	}
	msConnCfg := &microsoft.Config{
		ClientID:     s.MicrosoftCfg.ClientID,
		ClientSecret: s.MicrosoftCfg.ClientSecret,
		RedirectURI:  "https://" + s.ServerURL + "/issuer/callback",
		//	Scopes:       []string{"openid"}, //, "profile", "email"},
	}
	githubCfg := &github.Config{
		ClientID:      s.GithubCfg.ClientID,
		ClientSecret:  s.GithubCfg.ClientSecret,
		RedirectURI:   "https://" + s.ServerURL + "/issuer/callback",
		LoadAllGroups: true,
		UseLoginAsID:  true,
	}
	googleCfg := &google.Config{
		ClientID:     s.GoogleCfg.ClientID,
		ClientSecret: s.GoogleCfg.ClientSecret,
		RedirectURI:  "https://" + s.ServerURL + "/issuer/callback",
		//		Scopes:       []string{"openid", "profile", "email"},
	}
	appleCfg := &apple.Config{
		//	Issuer:      "https://appleid.apple.com",
		ClientID:    s.AppleCfg.ClientID,
		KeyID:       s.AppleCfg.KeyID,
		TeamID:      s.AppleCfg.TeamID,
		RedirectURI: "https://" + s.ServerURL + "/issuer/callback",
		PrivateKey:  s.AppleCfg.PrivateKey,
	}

	logrussor, err := newLogger("error", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %v", err)
	}
	storage, err := storageCfg.Config.Open(logrussor)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}
	//	defer storage.Close()
	storage = dexStorage.WithStaticClients(storage, []dexStorage.Client{{
		Name:   "MirageServer",
		ID:     "MirageServer",
		Secret: s.DexSecret, // TODO: change this to a random string
		RedirectURIs: []string{
			"https://" + s.ServerURL + "/a/oauth_response",
		},
	}})

	storageConnectors := make([]dexStorage.Connector, 4)
	for i, c := range []Connector{{
		ID:     "Microsoft",
		Name:   "Microsoft",
		Type:   "microsoft",
		Config: msConnCfg,
	}, {
		ID:     "Github",
		Name:   "Github",
		Type:   "github",
		Config: githubCfg,
	}, {
		ID:     "Google",
		Name:   "Google",
		Type:   "google",
		Config: googleCfg,
	}, {
		ID:     "Apple",
		Name:   "Apple",
		Type:   "apple",
		Config: appleCfg,
	}} {
		if c.ID == "" || c.Name == "" || c.Type == "" {
			return nil, fmt.Errorf("invalid config: ID, Type and Name fields are required for a connector")
		}
		if c.Config == nil {
			return nil, fmt.Errorf("invalid config: no config field for connector %q", c.ID)
		}
		logrussor.Infof("config connector: %s", c.ID)
		// convert to a storage connector object
		conn, err := ToStorageConnector(c)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage connectors: %v", err)
		}
		storageConnectors[i] = conn
	}
	storage = dexStorage.WithStaticConnectors(storage, storageConnectors)
	now := func() time.Time { return time.Now().UTC() }
	healthChecker := gosundheit.New()
	refreshTokenPolicy, err := server.NewRefreshTokenPolicy(
		logrussor,
		false, //c.Expiry.RefreshTokens.DisableRotation,
		"",    //c.Expiry.RefreshTokens.ValidIfNotUsedFor,
		"",    //c.Expiry.RefreshTokens.AbsoluteLifetime,
		"",    //c.Expiry.RefreshTokens.ReuseInterval,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token expiration policy config: %v", err)
	}
	return &server.Config{
		Issuer:                 "https://" + s.ServerURL + "/issuer",
		Storage:                storage,
		Logger:                 logrussor,
		SupportedResponseTypes: nil,                //c.OAuth2.ResponseTypes,
		SkipApprovalScreen:     false,              //c.OAuth2.SkipApprovalScreen,
		AlwaysShowLoginScreen:  true,               //c.OAuth2.AlwaysShowLoginScreen,
		PasswordConnector:      "",                 //c.OAuth2.PasswordConnector,
		AllowedOrigins:         nil,                //c.Web.AllowedOrigins,
		Web:                    server.WebConfig{}, //c.Frontend,
		Now:                    now,
		PrometheusRegistry:     nil,
		HealthChecker:          healthChecker,
		RefreshTokenPolicy:     refreshTokenPolicy,
	}, nil
}

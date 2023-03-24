package controller

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

type SysConfig struct {
	gorm.Model
	AdminCredential AdminCredential `gorm:"not null"`

	ServerURL             string
	Addr                  string   `gorm:"default:':8080'"`               // default port
	Mip4                  IPPrefix `gorm:"default:'100.64.0.0/10'"`       // default prefix
	Mip6                  IPPrefix `gorm:"default:'fd7a:115c:a1e0::/48'"` // default prefix
	Basedomain            string   `gorm:"default:'mira.net'"`            // default domain
	DerpUrl               string   `gorm:"default:'https://controlplane.tailscale.com/derpmap/default'"`
	RouteAccessDueMachine bool     `gorm:"default:false"`

	EsUrl string
	EsKey string

	WXScanURL string

	SMSConfig SMSConfig

	IdaasConfig ALIConfig

	//	OidcConfig OIDCConfig
	DexSecret string

	MicrosoftCfg MicrosoftCfg
	GithubCfg    GithubCfg
	GoogleCfg    GoogleCfg
	AppleCfg     AppleCfg

	CreatedAt time.Time
	UpdatedAt time.Time
}

type GeneralCfg struct {
	SrvAddr               string `json:"srvaddr"`
	ServerURL             string `json:"server_url"`
	MIPV4                 string `json:"mipv4"`
	MIPV6                 string `json:"mipv6"`
	BaseDomain            string `json:"basedomain"`
	DERPURL               string `json:"derp_url"`
	RouteAccessDueMachine bool   `json:"route_access_due_machine"`

	ESURL string `json:"es_url"`
	ESKey string `json:"es_key"`

	WXScanURL string `json:"wxscan_url"`

	SMSConfig   SMSConfig `json:"sms"`
	IDaaSConfig ALIConfig `json:"idaas"`

	MicrosoftCfg MicrosoftCfg `json:"microsoft"`
	GithubCfg    GithubCfg    `json:"github"`
	GoogleCfg    GoogleCfg    `json:"google"`
	AppleCfg     AppleCfg     `json:"apple"`
}

func (s *SysConfig) toGeneralCfg() GeneralCfg {
	return GeneralCfg{
		SrvAddr:               s.Addr,
		ServerURL:             s.ServerURL,
		MIPV4:                 s.Mip4.String(),
		MIPV6:                 s.Mip6.String(),
		BaseDomain:            s.Basedomain,
		DERPURL:               s.DerpUrl,
		RouteAccessDueMachine: s.RouteAccessDueMachine,

		ESURL: s.EsUrl,
		ESKey: s.EsKey,

		WXScanURL: s.WXScanURL,

		SMSConfig:    s.SMSConfig,
		IDaaSConfig:  s.IdaasConfig,
		MicrosoftCfg: s.MicrosoftCfg,
		GithubCfg:    s.GithubCfg,
		GoogleCfg:    s.GoogleCfg,
		AppleCfg:     s.AppleCfg,
	}
}
func (s *SysConfig) toSrvConfig() (*Config, error) {
	dexCfg, err := s.toDexConfig()
	if err != nil {
		return nil, err
	}
	idps := []string{}
	if s.MicrosoftCfg.ClientID != "" && s.MicrosoftCfg.ClientSecret != "" {
		idps = append(idps, "Microsoft")
	}
	if s.GithubCfg.ClientID != "" && s.GithubCfg.ClientSecret != "" {
		idps = append(idps, "Github")
	}
	if s.GoogleCfg.ClientID != "" && s.GoogleCfg.ClientSecret != "" {
		idps = append(idps, "Google")
	}
	if s.AppleCfg.ClientID != "" && s.AppleCfg.KeyID != "" && s.AppleCfg.TeamID != "" && s.AppleCfg.PrivateKey != "" {
		idps = append(idps, "Apple")
	}
	if s.WXScanURL != "" {
		idps = append(idps, "WeChat")
	}

	OidcConfig := OIDCConfig{
		Issuer:       "https://" + s.ServerURL + "/issuer",
		ClientID:     "MirageServer",
		ClientSecret: s.DexSecret,
		Scope:        []string{"openid", "profile", "email", "groups", "name"},
		ExtraParams:  map[string]string{"prompt": "login"},
	}

	return &Config{
		ServerURL:              s.ServerURL,
		Addr:                   s.Addr,
		IPPrefixes:             []netip.Prefix{netip.Prefix(s.Mip4), netip.Prefix(s.Mip6)},
		BaseDomain:             s.Basedomain,
		DERPURL:                s.DerpUrl,
		AllowRouteDueToMachine: s.RouteAccessDueMachine,

		ESURL: s.EsUrl,
		ESKey: s.EsKey,

		wxScanURL: s.WXScanURL,

		SMS:       s.SMSConfig,
		IDaaS:     s.IdaasConfig,
		OIDC:      OidcConfig,
		DexConfig: dexCfg,
		IdpList:   idps,
	}, nil
}

type AdminCredential webauthn.Credential

func (ac *AdminCredential) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ac)
	case string:
		return json.Unmarshal([]byte(v), ac)
	default:
		return fmt.Errorf("cannot parse admin credential: unexpected data type %T", value)
	}
}

func (ac AdminCredential) Value() (driver.Value, error) {
	bytes, err := json.Marshal(ac)
	return string(bytes), err
}

type MicrosoftCfg struct {
	ApplicationID string `json:"app_id"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
}

func (mscfg *MicrosoftCfg) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, mscfg)
	case string:
		return json.Unmarshal([]byte(v), mscfg)
	default:
		return fmt.Errorf("cannot parse microsoft config: unexpected data type %T", value)
	}
}

func (mscfg MicrosoftCfg) Value() (driver.Value, error) {
	bytes, err := json.Marshal(mscfg)
	return string(bytes), err
}

type GithubCfg struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (ghCfg *GithubCfg) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ghCfg)
	case string:
		return json.Unmarshal([]byte(v), ghCfg)
	default:
		return fmt.Errorf("cannot parse github config: unexpected data type %T", value)
	}
}

func (ghCfg GithubCfg) Value() (driver.Value, error) {
	bytes, err := json.Marshal(ghCfg)
	return string(bytes), err
}

type GoogleCfg struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (ghCfg *GoogleCfg) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ghCfg)
	case string:
		return json.Unmarshal([]byte(v), ghCfg)
	default:
		return fmt.Errorf("cannot parse github config: unexpected data type %T", value)
	}
}

func (ghCfg GoogleCfg) Value() (driver.Value, error) {
	bytes, err := json.Marshal(ghCfg)
	return string(bytes), err
}

type AppleCfg struct {
	ClientID   string `json:"client_id"`
	TeamID     string `json:"team_id"`
	KeyID      string `json:"key_id"`
	PrivateKey string `json:"private_key"`
}

func (ghCfg *AppleCfg) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ghCfg)
	case string:
		return json.Unmarshal([]byte(v), ghCfg)
	default:
		return fmt.Errorf("cannot parse github config: unexpected data type %T", value)
	}
}

func (ghCfg AppleCfg) Value() (driver.Value, error) {
	bytes, err := json.Marshal(ghCfg)
	return string(bytes), err
}

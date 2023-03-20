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

	OidcConfig OIDCConfig

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *SysConfig) toSrvConfig() (*Config, error) {
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

		SMS:   s.SMSConfig,
		IDaaS: s.IdaasConfig,
		OIDC:  s.OidcConfig,
	}, nil
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

	SMSConfig   SMSConfig  `json:"sms"`
	IDaaSConfig ALIConfig  `json:"idaas"`
	OIDCConfig  OIDCConfig `json:"oidc"`
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

		SMSConfig:   s.SMSConfig,
		IDaaSConfig: s.IdaasConfig,
		OIDCConfig:  s.OidcConfig,
	}
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

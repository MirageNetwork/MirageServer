package controller

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	"github.com/dexidp/dex/server"
)

const (
	JSONLogFormat = "json"
	TextLogFormat = "text"

	defaultOIDCExpiryTime               = 180 * 24 * time.Hour // 180 Days
	maxDuration           time.Duration = 1<<63 - 1
)

// Config contains the initial Mirage configuration.
type Config struct {
	ServerURL  string         //DONE
	Addr       string         //DONE
	IPPrefixes []netip.Prefix //DONE
	BaseDomain string         //DONE

	AllowRouteDueToMachine bool //DONE

	//	DERPURL string //DONE

	ESURL string
	ESKey string

	OIDC OIDCConfig

	wxScanURL string

	IDaaS ALIConfig
	SMS   SMSConfig

	DexConfig *server.Config
	IdpList   []string

	ClientVersion ClientVersionInfo
}

type SMSConfig struct {
	ID       string `json:"id"`
	Key      string `json:"key"`
	Sign     string `json:"sign"`
	Template string `json:"template"`
}

func (ac *SMSConfig) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ac)
	case string:
		return json.Unmarshal([]byte(v), ac)
	default:
		return fmt.Errorf("cannot parse SMS Config: unexpected data type %T", value)
	}
}

func (ac SMSConfig) Value() (driver.Value, error) {
	bytes, err := json.Marshal(ac)
	return string(bytes), err
}

type ALIConfig struct {
	App       string `json:"app"`
	ClientID  string `json:"id"`
	ClientKey string `json:"key"`
	Instance  string `json:"instance"`
	OrgID     string `json:"org"`
}

func (ac *ALIConfig) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ac)
	case string:
		return json.Unmarshal([]byte(v), ac)
	default:
		return fmt.Errorf("cannot parse Ali IDaaS Config: unexpected data type %T", value)
	}
}

func (ac ALIConfig) Value() (driver.Value, error) {
	bytes, err := json.Marshal(ac)
	return string(bytes), err
}

type OIDCConfig struct {
	Issuer           string            `json:"issuer"`
	ClientID         string            `json:"id"`
	ClientSecret     string            `json:"key"`
	Scope            []string          `json:"scope"`
	ExtraParams      map[string]string `json:"extra"`
	StripEmaildomain bool              `json:"strip_flag"`
}

func (ac *OIDCConfig) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ac)
	case string:
		return json.Unmarshal([]byte(v), ac)
	default:
		return fmt.Errorf("cannot parse OIDC Config: unexpected data type %T", value)
	}
}

func (ac OIDCConfig) Value() (driver.Value, error) {
	bytes, err := json.Marshal(ac)
	return string(bytes), err
}

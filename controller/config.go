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

	DERPURL string //DONE

	ESURL string
	ESKey string

	OIDC OIDCConfig

	wxScanURL string

	IDaaS ALIConfig
	SMS   SMSConfig

	DexConfig *server.Config
	IdpList   []string
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

/*
type ACLConfig struct {
	PolicyPath string
}


func LoadConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath(".mirage")

	viper.SetEnvPrefix("mirage")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("oidc.scope", []string{oidc.ScopeOpenID, "profile", "email"})
	viper.SetDefault("oidc.strip_email_domain", true)

	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("Failed to read configuration from disk")

		return fmt.Errorf("fatal error reading config file: %w", err)
	}

	return nil
}

func GetACLConfig() ACLConfig {
	policyPath := viper.GetString("acl_policy_path")

	return ACLConfig{
		PolicyPath: policyPath,
	}
}

func GetBasedomain() string {
	var baseDomain string
	if viper.IsSet("base_domain") {
		baseDomain = viper.GetString("base_domain")
	} else {
		baseDomain = "sdp.net" // does not really matter when MagicDNS is not enabled
	}

	return baseDomain
}

func GetMirageConfig(srvAddr, serverURL string) (*Config, error) {

	baseDomain := GetBasedomain()

	configuredPrefixes := viper.GetStringSlice("ip_prefixes")
	parsedPrefixes := make([]netip.Prefix, 0, len(configuredPrefixes)+1)

	for i, prefixInConfig := range configuredPrefixes {
		prefix, err := netip.ParsePrefix(prefixInConfig)
		if err != nil {
			panic(fmt.Errorf("failed to parse ip_prefixes[%d]: %w", i, err))
		}
		parsedPrefixes = append(parsedPrefixes, prefix)
	}

	prefixes := make([]netip.Prefix, 0, len(parsedPrefixes))
	{
		// dedup
		normalizedPrefixes := make(map[string]int, len(parsedPrefixes))
		for i, p := range parsedPrefixes {
			normalized, _ := netipx.RangeOfPrefix(p).Prefix()
			normalizedPrefixes[normalized.String()] = i
		}

		// convert back to list
		for _, i := range normalizedPrefixes {
			prefixes = append(prefixes, parsedPrefixes[i])
		}
	}

	if len(prefixes) < 1 {
		prefixes = append(prefixes, netip.MustParsePrefix("100.64.0.0/10"))
		log.Warn().
			Msgf("'ip_prefixes' not configured, falling back to default: %v", prefixes)
	}

	wxScanAdapterURL := viper.GetString("wxscan.url")

	return &Config{
		ServerURL:  serverURL,
		Addr:       srvAddr,
		IPPrefixes: prefixes,

		BaseDomain:             baseDomain,
		AllowRouteDueToMachine: viper.GetBool("allow_route_due_to_machine"),

		DERPURL: viper.GetString("derp_url"),

		ESURL: viper.GetString("es_url"),
		ESKey: viper.GetString("es_apikey"),

		OIDC: OIDCConfig{
			Issuer:           viper.GetString("ali.issuer"),
			ClientID:         viper.GetString("ali.client_id"),
			ClientSecret:     viper.GetString("ali.client_secret"),
			Scope:            viper.GetStringSlice("ali.scope"),
			ExtraParams:      viper.GetStringMapString("ali.extra_params"),
			StripEmaildomain: viper.GetBool("ali.strip_email_domain"),
		},
		wxScanURL: wxScanAdapterURL,

		IDaaS: ALIConfig{
			App:       viper.GetString("idaas.app_id"),
			ClientID:  viper.GetString("idaas.cli_id"),
			ClientKey: viper.GetString("idaas.cli_key"),
			Instance:  viper.GetString("idaas.instance"),
			OrgID:     viper.GetString("idaas.org_id"),
		},
		SMS: SMSConfig{
			ID:       viper.GetString("sms.access_id"),
			Key:      viper.GetString("sms.access_key"),
			Sign:     viper.GetString("sms.sms_sign"),
			Template: viper.GetString("sms.sms_template"),
		},
	}, nil
}
*/

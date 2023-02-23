package controller

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go4.org/netipx"
)

const (
	JSONLogFormat = "json"
	TextLogFormat = "text"

	defaultOIDCExpiryTime               = 180 * 24 * time.Hour // 180 Days
	maxDuration           time.Duration = 1<<63 - 1
)

var errOidcMutuallyExclusive = errors.New(
	"oidc_client_secret and oidc_client_secret_path are mutually exclusive",
)

// Config contains the initial Mirage configuration.
type Config struct {
	ServerURL  string
	Addr       string
	IPPrefixes []netip.Prefix
	BaseDomain string

	DERPURL string

	ESURL string
	ESKey string

	OIDC OIDCConfig

	wxScanURL string

	ali_IDaaS ALIConfig

	org_name string
}
type ALIConfig struct {
	ali_app_id       string
	ali_cli_id       string
	ali_cli_key      string
	ali_instance     string
	ali_org_id       string
	ali_access_id    string
	ali_access_key   string
	ali_sms_sign     string
	ali_sms_template string
}

type OIDCConfig struct {
	Issuer           string
	ClientID         string
	ClientSecret     string
	Scope            []string
	ExtraParams      map[string]string
	StripEmaildomain bool
}

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

func GetMirageConfig() (*Config, error) {

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

	oidcClientSecret := viper.GetString("oidc.client_secret")
	oidcClientSecretPath := viper.GetString("oidc.client_secret_path")
	if oidcClientSecretPath != "" && oidcClientSecret != "" {
		return nil, errOidcMutuallyExclusive
	}
	if oidcClientSecretPath != "" {
		secretBytes, err := os.ReadFile(os.ExpandEnv(oidcClientSecretPath))
		if err != nil {
			return nil, err
		}
		oidcClientSecret = string(secretBytes)
	}

	return &Config{
		ServerURL: viper.GetString("server_url"),
		Addr:      viper.GetString("listen_addr"),

		IPPrefixes: prefixes,

		BaseDomain: baseDomain,

		DERPURL: viper.GetString("derp_url"),

		ESURL: viper.GetString("es_url"),
		ESKey: viper.GetString("es_apikey"),

		OIDC: OIDCConfig{
			Issuer:           viper.GetString("oidc.issuer"),
			ClientID:         viper.GetString("oidc.client_id"),
			ClientSecret:     oidcClientSecret,
			Scope:            viper.GetStringSlice("oidc.scope"),
			ExtraParams:      viper.GetStringMapString("oidc.extra_params"),
			StripEmaildomain: viper.GetBool("oidc.strip_email_domain"),
		},
		wxScanURL: wxScanAdapterURL,

		ali_IDaaS: ALIConfig{
			ali_app_id:       viper.GetString("ali_app_id"),
			ali_cli_id:       viper.GetString("ali_cli_id"),
			ali_cli_key:      viper.GetString("ali_cli_key"),
			ali_instance:     viper.GetString("ali_instance"),
			ali_org_id:       viper.GetString("ali_org_id"),
			ali_access_id:    viper.GetString("ali_access_id"),
			ali_access_key:   viper.GetString("ali_access_key"),
			ali_sms_sign:     viper.GetString("ali_sms_sign"),
			ali_sms_template: viper.GetString("ali_sms_template"),
		},
		org_name: viper.GetString("org_name"),
	}, nil
}

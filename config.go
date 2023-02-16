package Mirage

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/prometheus/common/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go4.org/netipx"
	"tailscale.com/tailcfg"
	"tailscale.com/types/dnstype"
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
	ServerURL                      string
	Addr                           string
	EphemeralNodeInactivityTimeout time.Duration
	NodeUpdateCheckInterval        time.Duration
	IPPrefixes                     []netip.Prefix
	PrivateKeyPath                 string
	NoisePrivateKeyPath            string
	BaseDomain                     string
	Log                            LogConfig

	DERP DERPConfig

	DBtype string
	DBpath string
	DBhost string
	DBport int
	DBname string
	DBuser string
	DBpass string
	DBssl  string

	DNSConfig *tailcfg.DNSConfig

	OIDC OIDCConfig

	ACL ACLConfig

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
	OnlyStartIfOIDCIsAvailable bool
	Issuer                     string
	ClientID                   string
	ClientSecret               string
	RevokeURL                  string
	LogoutURL                  string
	Scope                      []string
	ExtraParams                map[string]string
	AllowedDomains             []string
	AllowedUsers               []string
	AllowedGroups              []string
	StripEmaildomain           bool
	Expiry                     time.Duration
	UseExpiryFromToken         bool
}

type DERPConfig struct {
	URLs            []url.URL
	Paths           []string
	AutoUpdate      bool
	UpdateFrequency time.Duration
}

type LogTailConfig struct {
	Enabled bool
}

type CLIConfig struct {
	Address  string
	APIKey   string
	Timeout  time.Duration
	Insecure bool
}

type ACLConfig struct {
	PolicyPath string
}

type LogConfig struct {
	Format string
	Level  zerolog.Level
}

func LoadConfig(path string, isFile bool) error {
	if isFile {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("config")
		if path == "" {
			viper.AddConfigPath("/etc/mirage/")
			viper.AddConfigPath("$HOME/.mirage")
			viper.AddConfigPath(".")
		} else {
			// For testing
			viper.AddConfigPath(path)
		}
	}

	viper.SetEnvPrefix("mirage")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", TextLogFormat)

	viper.SetDefault("dns_config", nil)
	viper.SetDefault("dns_config.override_local_dns", true)

	viper.SetDefault("db_ssl", false)

	viper.SetDefault("oidc.scope", []string{oidc.ScopeOpenID, "profile", "email"})
	viper.SetDefault("oidc.strip_email_domain", true)
	viper.SetDefault("oidc.only_start_if_oidc_is_available", true)
	viper.SetDefault("oidc.expiry", "180d")
	viper.SetDefault("oidc.use_expiry_from_token", false)

	viper.SetDefault("logtail.enabled", false)
	viper.SetDefault("randomize_client_port", false)

	viper.SetDefault("ephemeral_node_inactivity_timeout", "120s")

	viper.SetDefault("node_update_check_interval", "10s")

	if IsCLIConfigured() {
		return nil
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("Failed to read configuration from disk")

		return fmt.Errorf("fatal error reading config file: %w", err)
	}

	// Collect any validation errors and return them all at once
	var errorText string

	if !viper.IsSet("noise") || viper.GetString("noise.private_key_path") == "" {
		errorText += "Fatal config error: headscale now requires a new `noise.private_key_path` field in the config file for the Tailscale v2 protocol\n"
	}

	if !strings.HasPrefix(viper.GetString("server_url"), "http://") &&
		!strings.HasPrefix(viper.GetString("server_url"), "https://") {
		errorText += "Fatal config error: server_url must start with https:// or http://\n"
	}

	// Minimum inactivity time out is keepalive timeout (60s) plus a few seconds
	// to avoid races
	minInactivityTimeout, _ := time.ParseDuration("65s")
	if viper.GetDuration("ephemeral_node_inactivity_timeout") <= minInactivityTimeout {
		errorText += fmt.Sprintf(
			"Fatal config error: ephemeral_node_inactivity_timeout (%s) is set too low, must be more than %s",
			viper.GetString("ephemeral_node_inactivity_timeout"),
			minInactivityTimeout,
		)
	}

	maxNodeUpdateCheckInterval, _ := time.ParseDuration("60s")
	if viper.GetDuration("node_update_check_interval") > maxNodeUpdateCheckInterval {
		errorText += fmt.Sprintf(
			"Fatal config error: node_update_check_interval (%s) is set too high, must be less than %s",
			viper.GetString("node_update_check_interval"),
			maxNodeUpdateCheckInterval,
		)
	}

	if errorText != "" {
		//nolint
		return errors.New(strings.TrimSuffix(errorText, "\n"))
	} else {
		return nil
	}
}

func GetDERPConfig() DERPConfig {
	urlStrs := viper.GetStringSlice("derp.urls")

	urls := make([]url.URL, len(urlStrs))
	for index, urlStr := range urlStrs {
		urlAddr, err := url.Parse(urlStr)
		if err != nil {
			log.Error().
				Str("url", urlStr).
				Err(err).
				Msg("Failed to parse url, ignoring...")
		}

		urls[index] = *urlAddr
	}

	paths := viper.GetStringSlice("derp.paths")

	autoUpdate := viper.GetBool("derp.auto_update_enabled")
	updateFrequency := viper.GetDuration("derp.update_frequency")

	return DERPConfig{
		URLs:            urls,
		Paths:           paths,
		AutoUpdate:      autoUpdate,
		UpdateFrequency: updateFrequency,
	}
}

func GetLogTailConfig() LogTailConfig {
	enabled := viper.GetBool("logtail.enabled")

	return LogTailConfig{
		Enabled: enabled,
	}
}

func GetACLConfig() ACLConfig {
	policyPath := viper.GetString("acl_policy_path")

	return ACLConfig{
		PolicyPath: policyPath,
	}
}

func GetDNSConfig() (*tailcfg.DNSConfig, string) {
	if viper.IsSet("dns_config") {
		dnsConfig := &tailcfg.DNSConfig{}

		overrideLocalDNS := viper.GetBool("dns_config.override_local_dns")

		if viper.IsSet("dns_config.nameservers") {
			nameserversStr := viper.GetStringSlice("dns_config.nameservers")

			nameservers := []netip.Addr{}
			resolvers := []*dnstype.Resolver{}

			for _, nameserverStr := range nameserversStr {
				// Search for explicit DNS-over-HTTPS resolvers
				if strings.HasPrefix(nameserverStr, "https://") {
					resolvers = append(resolvers, &dnstype.Resolver{
						Addr: nameserverStr,
					})

					// This nameserver can not be parsed as an IP address
					continue
				}

				// Parse nameserver as a regular IP
				nameserver, err := netip.ParseAddr(nameserverStr)
				if err != nil {
					log.Error().
						Str("func", "getDNSConfig").
						Err(err).
						Msgf("Could not parse nameserver IP: %s", nameserverStr)
				}

				nameservers = append(nameservers, nameserver)
				resolvers = append(resolvers, &dnstype.Resolver{
					Addr: nameserver.String(),
				})
			}

			dnsConfig.Nameservers = nameservers

			if overrideLocalDNS {
				dnsConfig.Resolvers = resolvers
			} else {
				dnsConfig.FallbackResolvers = resolvers
			}
		}

		//cgao6: split DNS related here
		if viper.IsSet("dns_config.restricted_nameservers") {
			dnsConfig.Routes = make(map[string][]*dnstype.Resolver)
			domains := []string{}
			restrictedDNS := viper.GetStringMapStringSlice(
				"dns_config.restricted_nameservers",
			)
			for domain, restrictedNameservers := range restrictedDNS {
				restrictedResolvers := make(
					[]*dnstype.Resolver,
					len(restrictedNameservers),
				)
				for index, nameserverStr := range restrictedNameservers {
					nameserver, err := netip.ParseAddr(nameserverStr)
					if err != nil {
						log.Error().
							Str("func", "getDNSConfig").
							Err(err).
							Msgf("Could not parse restricted nameserver IP: %s", nameserverStr)
					}
					restrictedResolvers[index] = &dnstype.Resolver{
						Addr: nameserver.String(),
					}
				}
				dnsConfig.Routes[domain] = restrictedResolvers
				domains = append(domains, domain)
			}
			dnsConfig.Domains = domains
		}

		if viper.IsSet("dns_config.domains") {
			domains := viper.GetStringSlice("dns_config.domains")
			if len(dnsConfig.Resolvers) > 0 {
				dnsConfig.Domains = domains
			} else if domains != nil {
				log.Warn().
					Msg("Warning: dns_config.domains is set, but no nameservers are configured. Ignoring domains.")
			}
		}

		if viper.IsSet("dns_config.extra_records") {
			var extraRecords []tailcfg.DNSRecord

			err := viper.UnmarshalKey("dns_config.extra_records", &extraRecords)
			if err != nil {
				log.Error().
					Str("func", "getDNSConfig").
					Err(err).
					Msgf("Could not parse dns_config.extra_records")
			}

			dnsConfig.ExtraRecords = extraRecords
		}

		if viper.IsSet("dns_config.magic_dns") {
			dnsConfig.Proxied = viper.GetBool("dns_config.magic_dns")
		}

		var baseDomain string
		if viper.IsSet("dns_config.base_domain") {
			baseDomain = viper.GetString("dns_config.base_domain")
		} else {
			baseDomain = "headscale.net" // does not really matter when MagicDNS is not enabled
		}

		return dnsConfig, baseDomain
	}

	return nil, ""
}

func GetMirageConfig() (*Config, error) {

	dnsConfig, baseDomain := GetDNSConfig()
	derpConfig := GetDERPConfig()

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

		NoisePrivateKeyPath: AbsolutePathFromConfigPath(
			viper.GetString("noise.private_key_path"),
		),
		BaseDomain: baseDomain,

		DERP: derpConfig,

		EphemeralNodeInactivityTimeout: viper.GetDuration(
			"ephemeral_node_inactivity_timeout",
		),

		NodeUpdateCheckInterval: viper.GetDuration(
			"node_update_check_interval",
		),

		DBtype: viper.GetString("db_type"),
		DBpath: AbsolutePathFromConfigPath(viper.GetString("db_path")),
		DBhost: viper.GetString("db_host"),
		DBport: viper.GetInt("db_port"),
		DBname: viper.GetString("db_name"),
		DBuser: viper.GetString("db_user"),
		DBpass: viper.GetString("db_pass"),
		DBssl:  viper.GetString("db_ssl"),

		DNSConfig: dnsConfig,

		OIDC: OIDCConfig{
			OnlyStartIfOIDCIsAvailable: viper.GetBool(
				"oidc.only_start_if_oidc_is_available",
			),
			Issuer:           viper.GetString("oidc.issuer"),
			ClientID:         viper.GetString("oidc.client_id"),
			ClientSecret:     oidcClientSecret,
			LogoutURL:        viper.GetString("oidc.logout_url"),
			RevokeURL:        viper.GetString("oidc.revoke_url"),
			Scope:            viper.GetStringSlice("oidc.scope"),
			ExtraParams:      viper.GetStringMapString("oidc.extra_params"),
			AllowedDomains:   viper.GetStringSlice("oidc.allowed_domains"),
			AllowedUsers:     viper.GetStringSlice("oidc.allowed_users"),
			AllowedGroups:    viper.GetStringSlice("oidc.allowed_groups"),
			StripEmaildomain: viper.GetBool("oidc.strip_email_domain"),
			Expiry: func() time.Duration {
				// if set to 0, we assume no expiry
				if value := viper.GetString("oidc.expiry"); value == "0" {
					return maxDuration
				} else {
					expiry, err := model.ParseDuration(value)
					if err != nil {
						log.Warn().Msg("failed to parse oidc.expiry, defaulting back to 180 days")

						return defaultOIDCExpiryTime
					}

					return time.Duration(expiry)
				}
			}(),
			UseExpiryFromToken: viper.GetBool("oidc.use_expiry_from_token"),
		},

		ACL: GetACLConfig(),

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

func IsCLIConfigured() bool {
	return viper.GetString("cli.address") != "" && viper.GetString("cli.api_key") != ""
}

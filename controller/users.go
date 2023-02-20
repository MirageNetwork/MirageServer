package controller

import (
	"errors"
	"fmt"
	"net/netip"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"tailscale.com/tailcfg"
	"tailscale.com/types/dnstype"
)

const (
	ErrUserExists        = Error("User already exists")
	ErrUserNotFound      = Error("User not found")
	ErrUserStillHasNodes = Error("User not empty: node(s) found")
	ErrInvalidUserName   = Error("Invalid user name")
)

const (
	// value related to RFC 1123 and 952.
	labelHostnameLength = 63
)

var invalidCharsInUserRegex = regexp.MustCompile("[^a-z0-9-.]+")

// User is the way Mirage implements the concept of users in Tailscale
//
// At the end of the day, users in Tailscale are some kind of 'bubbles' or users
// that contain our machines.
type User struct {
	gorm.Model
	Name           string `gorm:"unique"`
	Display_Name   string `gorm:"unique"`
	ExpiryDuration uint   `gorm:"default:180"`
	EnableMagic    bool   `gorm:"default:false"`
	OverrideLocal  bool   `gorm:"default:false"`
	Nameservers    StringList
	SplitDns       SplitDNS
}

// CreateUser creates a new User. Returns error if could not be created
// or another user already exists.
func (h *Mirage) CreateUser(name string, UID string, DisName string) (*User, error) {
	err := CheckForFQDNRules(name)
	if err != nil {
		return nil, err
	}
	user := User{}
	if err := h.db.Where("name = ?", name).First(&user).Error; err == nil {
		return nil, ErrUserExists
	}
	user.Name = name
	user.ExpiryDuration = 180
	user.Display_Name = DisName
	if err := h.db.Create(&user).Error; err != nil {
		log.Error().
			Str("func", "CreateUser").
			Err(err).
			Msg("Could not create row")

		return nil, err
	}

	return &user, nil
}

// DestroyUser destroys a User. Returns error if the User does
// not exist or if there are machines associated with it.
func (h *Mirage) DestroyUser(name string) error {
	user, err := h.GetUser(name)
	if err != nil {
		return ErrUserNotFound
	}

	machines, err := h.ListMachinesByUser(name)
	if err != nil {
		return err
	}
	if len(machines) > 0 {
		return ErrUserStillHasNodes
	}

	keys, err := h.ListPreAuthKeys(name)
	if err != nil {
		return err
	}
	for _, key := range keys {
		err = h.DestroyPreAuthKey(key)
		if err != nil {
			return err
		}
	}

	if result := h.db.Unscoped().Delete(&user); result.Error != nil {
		return result.Error
	}

	return nil
}

// Update User's node key expiry duration.
func (h *Mirage) UpdateUserKeyExpiry(name string, newDuration uint) error {
	var err error
	user, err := h.GetUser(name)
	if err != nil {
		return err
	}
	user.ExpiryDuration = newDuration

	if result := h.db.Save(&user); result.Error != nil {
		return result.Error
	}
	return nil
}

// RenameUser renames a User. Returns error if the User does
// not exist or if another User exists with the new name.
func (h *Mirage) RenameUser(oldName, newName string) error {
	var err error
	oldUser, err := h.GetUser(oldName)
	if err != nil {
		return err
	}
	err = CheckForFQDNRules(newName)
	if err != nil {
		return err
	}
	_, err = h.GetUser(newName)
	if err == nil {
		return ErrUserExists
	}
	if !errors.Is(err, ErrUserNotFound) {
		return err
	}

	oldUser.Name = newName

	if result := h.db.Save(&oldUser); result.Error != nil {
		return result.Error
	}

	return nil
}

// GetUser fetches a user by name.
func (h *Mirage) GetUserByID(id tailcfg.UserID) (*User, error) {
	user := User{}
	if result := h.db.First(&user, "id = ?", id); errors.Is(
		result.Error,
		gorm.ErrRecordNotFound,
	) {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

// GetUser fetches a user by name.
func (h *Mirage) GetUser(name string) (*User, error) {
	user := User{}
	if result := h.db.First(&user, "name = ?", name); errors.Is(
		result.Error,
		gorm.ErrRecordNotFound,
	) {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

// ListUsers gets all the existing users.
func (h *Mirage) ListUsers() ([]User, error) {
	users := []User{}
	if err := h.db.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// ListMachinesByUser gets all the nodes in a given user.
func (h *Mirage) ListMachinesByUser(name string) ([]Machine, error) {
	err := CheckForFQDNRules(name)
	if err != nil {
		return nil, err
	}
	user, err := h.GetUser(name)
	if err != nil {
		return nil, err
	}

	machines := []Machine{}
	if err := h.db.Preload("AuthKey").Preload("AuthKey.User").Preload("User").Where(&Machine{UserID: user.ID}).Find(&machines).Error; err != nil {
		return nil, err
	}

	return machines, nil
}

// SetMachineUser assigns a Machine to a user.
func (h *Mirage) SetMachineUser(machine *Machine, username string) error {
	err := CheckForFQDNRules(username)
	if err != nil {
		return err
	}
	user, err := h.GetUser(username)
	if err != nil {
		return err
	}
	machine.User = *user
	if result := h.db.Save(&machine); result.Error != nil {
		return result.Error
	}

	return nil
}

func (n *User) toTailscaleUser() *tailcfg.User {
	user := tailcfg.User{
		ID:            tailcfg.UserID(n.ID),
		LoginName:     n.Name,
		DisplayName:   n.Display_Name,
		ProfilePicURL: "",
		Domain:        "headscale.net",
		Logins:        []tailcfg.LoginID{},
		Created:       time.Time{},
	}

	return &user
}

func (n *User) toTailscaleLogin() *tailcfg.Login {
	login := tailcfg.Login{
		ID:            tailcfg.LoginID(n.ID),
		LoginName:     n.Name,
		DisplayName:   n.Name,
		ProfilePicURL: "",
		Domain:        "headscale.net",
	}

	return &login
}

func (h *Mirage) getMapResponseUserProfiles(
	machine Machine,
	peers Machines,
) []tailcfg.UserProfile {
	userMap := make(map[string]User)
	userMap[machine.User.Name] = machine.User
	for _, peer := range peers {
		userMap[peer.User.Name] = peer.User // not worth checking if already is there
	}

	profiles := []tailcfg.UserProfile{}
	for _, user := range userMap {
		/* cgao6 we do not use this logic for current
		displayName := user.Display_Name

		if h.cfg.BaseDomain != "" {
			displayName = fmt.Sprintf("%s@%s", user.Name, h.cfg.BaseDomain)
		}
		*/

		profiles = append(profiles,
			tailcfg.UserProfile{
				ID:          tailcfg.UserID(user.ID),
				LoginName:   user.Name,
				DisplayName: user.Display_Name,
			})
	}

	return profiles
}

// NormalizeToFQDNRules will replace forbidden chars in user
// it can also return an error if the user doesn't respect RFC 952 and 1123.
func NormalizeToFQDNRules(name string, stripEmailDomain bool) (string, error) {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "'", "")
	atIdx := strings.Index(name, "@")
	if stripEmailDomain && atIdx > 0 {
		name = name[:atIdx]
	} else {
		name = strings.ReplaceAll(name, "@", ".")
	}
	name = invalidCharsInUserRegex.ReplaceAllString(name, "-")

	for _, elt := range strings.Split(name, ".") {
		if len(elt) > labelHostnameLength {
			return "", fmt.Errorf(
				"label %v is more than 63 chars: %w",
				elt,
				ErrInvalidUserName,
			)
		}
	}

	return name, nil
}

func CheckForFQDNRules(name string) error {
	if len(name) > labelHostnameLength {
		return fmt.Errorf(
			"DNS segment must not be over 63 chars. %v doesn't comply with this rule: %w",
			name,
			ErrInvalidUserName,
		)
	}
	if strings.ToLower(name) != name {
		return fmt.Errorf(
			"DNS segment should be lowercase. %v doesn't comply with this rule: %w",
			name,
			ErrInvalidUserName,
		)
	}
	if invalidCharsInUserRegex.MatchString(name) {
		return fmt.Errorf(
			"DNS segment should only be composed of lowercase ASCII letters numbers, hyphen and dots. %v doesn't comply with theses rules: %w",
			name,
			ErrInvalidUserName,
		)
	}

	return nil
}

// 增加User独立配置的DNS设置读取
func (me *User) GetDNSConfig(ipPrefixesCfg []netip.Prefix) (*tailcfg.DNSConfig, string) {
	dnsConfig := &tailcfg.DNSConfig{}

	nameserversStr := me.Nameservers

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

	if me.OverrideLocal {
		dnsConfig.Resolvers = resolvers
	} else {
		dnsConfig.FallbackResolvers = resolvers
	}

	//cgao6: split DNS related here
	dnsConfig.Routes = make(map[string][]*dnstype.Resolver)
	domains := []string{}
	restrictedDNS := me.SplitDns
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

	//cgao6: TODO
	/*
		if viper.IsSet("dns_config.domains") {
			domains := viper.GetStringSlice("dns_config.domains")
			if len(dnsConfig.Resolvers) > 0 {
				dnsConfig.Domains = domains
			} else if domains != nil {
				log.Warn().
					Msg("Warning: dns_config.domains is set, but no nameservers are configured. Ignoring domains.")
			}
		}*/

	//cgao6: TODO
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

	dnsConfig.Proxied = me.EnableMagic

	if dnsConfig.Proxied { // if MagicDNS
		magicDNSDomains := generateMagicDNSRootDomains(ipPrefixesCfg)
		// we might have routes already from Split DNS
		if dnsConfig.Routes == nil {
			dnsConfig.Routes = make(map[string][]*dnstype.Resolver)
		}
		for _, d := range magicDNSDomains {
			dnsConfig.Routes[d.WithoutTrailingDot()] = nil
		}
	}

	var baseDomain string
	if viper.IsSet("dns_config.base_domain") {
		baseDomain = viper.GetString("dns_config.base_domain")
	} else {
		baseDomain = "headscale.net" // does not really matter when MagicDNS is not enabled
	}

	return dnsConfig, baseDomain
}

func (h *Mirage) UpdateDNSConfig(userName string, newDNSCfg DNSData) error {
	var err error
	user, err := h.GetUser(userName)
	if err != nil {
		return err
	}
	user.EnableMagic = newDNSCfg.MagicDNS
	user.Nameservers = make([]string, 0)
	if len(newDNSCfg.Resolvers) > 0 {
		user.OverrideLocal = true
		user.Nameservers = newDNSCfg.Resolvers
	} else if len(newDNSCfg.FallbackResolvers) > 0 {
		user.OverrideLocal = false
		user.Nameservers = newDNSCfg.FallbackResolvers
	}
	user.SplitDns = newDNSCfg.Routes

	if result := h.db.Save(user); result.Error != nil {
		return result.Error
	}

	return nil
}

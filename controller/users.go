package controller

import (
	"errors"
	"fmt"
	"net/netip"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
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
	ErrChangeUserRole    = Error("Change user role failed")
)

const ( // TODO: 重新规划映射关系
	RoleMember = 0
	RoleOwner  = 1
)

var RoleValue = map[string]int64{
	"member": RoleMember,
	"owner":  RoleOwner,
}

var RoleStr = map[int64]string{
	RoleMember: "member",
	RoleOwner:  "owner",
}

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
	ID             int64  `gorm:"primary_key;unique;not null"`
	StableID       string `gorm:"unique"`
	Name           string `gorm:"uniqueIndex:idx_user_org_id"`
	OrganizationID int64  `gorm:"uniqueIndex:idx_user_org_id"`
	Organization   Organization
	Display_Name   string //`gorm:"unique"`
	Role           int64
	//IsBelongToOrg bool `gorm:"default:false"`

	//TODO 哪些字段是user也需要的
	/*
		ExpiryDuration uint `gorm:"default:180"`
		ExpiryDuration uint `gorm:"default:180"`
		EnableMagic    bool `gorm:"default:false"`
		OverrideLocal  bool `gorm:"default:false"`
		Nameservers    StringList
		SplitDns       SplitDNS
	*/
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (user *User) BeforeCreate(tx *gorm.DB) error {
	if user.ID == 0 {
		flakeID, err := snowflake.NewNode(1)
		if err != nil {
			return err
		}
		id := flakeID.Generate().Int64()
		user.ID = id
	}
	user.StableID = GetShortId(user.ID)
	return nil
}

func (user *User) CheckEmpty() bool {
	return user == nil || user.ID == 0
}

// CreateUser creates a new User. Returns error if could not be created
// or another user already exists.
func (h *Mirage) CreateUser(name string, disName string, orgName string, provider string) (*User, error) {
	/*
		err := CheckForFQDNRules(name)
		if err != nil {
			return nil, err
		}
	*/
	user := User{}
	user.Name = name
	user.Display_Name = disName
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var org *Organization
		var trxErr error
		//需要先查询orgName是否存在
		org, trxErr = GetOrgnaizationByNameInTx(tx, orgName, provider)
		// 不存在: 创建组织
		if errors.Is(trxErr, ErrOrgNotFound) {
			org, trxErr = h.CreateOrgnaizationInTx(tx, orgName, provider)
			user.Role = RoleOwner
			// 其他错误, 报错返回
		} else if trxErr == nil && org.ID == 0 {
			trxErr = ErrOrgNotFound
		}
		//user.IsBelongToOrg = true
		if trxErr == nil {
			user.OrganizationID = org.ID
			user.Organization = *org
			trxErr = tx.Create(&user).Error
		}
		if trxErr != nil {
			log.Error().
				Str("func", "CreateUser").
				Err(trxErr).
				Msg("Could not create row")

			return trxErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// DestroyUser destroys a User. Returns error if the User does
// not exist or if there are machines associated with it.
func (h *Mirage) DestroyUser(name, orgName, provider string) error {
	user, err := h.GetUser(name, orgName, provider)
	if err != nil {
		return ErrUserNotFound
	}

	machines, err := h.ListMachinesByUser(user.ID)
	if err != nil {
		return err
	}
	if len(machines) > 0 {
		return ErrUserStillHasNodes
	}

	keys, err := h.ListPreAuthKeys(user.ID)
	if err != nil {
		return err
	}
	for _, key := range keys {
		err = h.DestroyPreAuthKey(key)
		if err != nil {
			return err
		}
	}

	return h.db.Unscoped().Delete(&user).Error
}

// Destroy User By userID, not need to get user from organization infos
func (h *Mirage) DestroyUserByID(userId int64) error {
	if userId <= 0 {
		return ErrUserNotFound
	}

	machines, err := h.ListMachinesByUser(userId)
	if err != nil {
		return err
	}
	if len(machines) > 0 {
		return ErrUserStillHasNodes
	}

	keys, err := h.ListPreAuthKeys(userId)
	if err != nil {
		return err
	}
	for _, key := range keys {
		err = h.DestroyPreAuthKey(key)
		if err != nil {
			return err
		}
	}

	return h.db.Unscoped().Delete(&User{ID: userId}).Error
}

// Force Destroy User By userID, Delete the machines of the user
func (h *Mirage) ForceDestroyUserByID(userId int64) error {
	if userId <= 0 {
		return ErrUserNotFound
	}

	machines, err := h.ListMachinesByUser(userId)
	if err != nil {
		return err
	}
	for _, machine := range machines {
		if err := h.HardDeleteMachine(&machine); err != nil {
			return nil
		}
		h.NotifyNaviOrgNodesChange(machine.User.OrganizationID, "", machine.NodeKey)
	}

	keys, err := h.ListPreAuthKeys(userId)
	if err != nil {
		return err
	}
	for _, key := range keys {
		err = h.DestroyPreAuthKey(key)
		if err != nil {
			return err
		}
	}

	return h.db.Unscoped().Delete(&User{ID: userId}).Error
}

// Update User's node key expiry duration.
func (h *Mirage) UpdateUserKeyExpiry(user *User, newDuration uint) error {
	if user == nil || user.OrganizationID == 0 {
		return ErrOrgNotFound
	}
	return h.UpdateOrgExpiry(user, newDuration)
}

func (h *Mirage) CheckUserExistence(name, orgName, provider string) (bool, error) {
	orgID, err := h.GetOrgnaizationIDByName(orgName, provider)
	if err != nil {
		return false, err
	}
	var count int64
	err = h.db.Model(&User{}).Where(&User{
		Name:           name,
		OrganizationID: orgID,
	}).Count(&count).Error
	return count > 0, err
}

// RenameUser renames a User. Returns error if the User does
// not exist or if another User exists with the new name.
func (h *Mirage) RenameUser(oldName, newName string, orgName string, provider string) error {
	var err error
	oldUser, err := h.GetUser(oldName, orgName, provider)
	if err != nil {
		return err
	}
	/*
		err = CheckForFQDNRules(newName)
		if err != nil {
			return err
		}
	*/
	if isEx, err := h.CheckUserExistence(newName, orgName, provider); err != nil {
		return err
	} else if isEx {
		return ErrUserExists
	}

	oldUser.Name = newName

	if result := h.db.Save(&oldUser); result.Error != nil {
		return result.Error
	}

	return nil
}

// ChangUserRole by userID
func (h *Mirage) ChangUserRole(id tailcfg.UserID, role string) error {
	if r, ok := RoleValue[strings.ToLower(role)]; ok {
		err := h.db.Select("Role").Updates(&User{
			ID:   int64(id),
			Role: r,
		}).Error
		return err
	}
	return ErrUserNotFound
}

// Transfer Owner by userID
func (h *Mirage) TransferOwner(srcId, destId tailcfg.UserID) error {
	return h.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Select("Role").Updates(&User{ID: int64(srcId), Role: RoleMember})
		if result.Error != nil || result.RowsAffected == 0 {
			return ErrChangeUserRole
		}
		result = tx.Select("Role").Updates(&User{ID: int64(destId), Role: RoleOwner})
		if result.Error != nil || result.RowsAffected == 0 {
			return ErrChangeUserRole
		}
		return nil
	})
}

// GetUserByID fetches a user by UserID.
func (h *Mirage) GetUserByID(id tailcfg.UserID) (*User, error) {
	user := User{}
	if result := h.db.Preload("Organization").First(&user, "id = ?", id); errors.Is(
		result.Error,
		gorm.ErrRecordNotFound,
	) {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

func (h *Mirage) ListOrgUsers(orgID int64) ([]User, error) {
	var users []User
	err := h.db.Preload("Organization").Where(&User{
		OrganizationID: orgID,
	}).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (h *Mirage) ListUsersInOrgs(orgID []int64) ([]User, error) {
	var users []User
	err := h.db.Preload("Organization").Where("organization_id in ?", orgID).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetUser fetches a user by name.
func (h *Mirage) GetUser(name, orgName, provider string) (*User, error) {
	org, err := h.GetOrgnaizationRecordByName(orgName, provider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = ErrOrgNotFound
		}
		return nil, err
	}
	user := User{}
	err = h.db.Where(&User{
		Name:           name,
		OrganizationID: org.ID,
	}).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	user.Organization = *org

	return &user, nil
}

/*
func (h *Mirage) GetUser(name string) (*User, error) {
	user := User{}
	if result := h.db.Preload("Organization").First(&user, "name = ?", name); errors.Is(
		result.Error,
		gorm.ErrRecordNotFound,
	) {
		return nil, ErrUserNotFound
	}

	return &user, nil
}
*/

// ListUsers gets all the existing users.
func (h *Mirage) ListUsers() ([]User, error) {
	users := []User{}
	if err := h.db.Preload("Organization").Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// ListMachinesByUser gets all the nodes in a given user.
func (h *Mirage) ListMachinesByUser(userID int64) ([]Machine, error) {
	if userID == 0 {
		return nil, ErrUserNotFound
	}
	machines := []Machine{}
	//TODO 是否需要组织信息
	if err := h.db.Preload("AuthKey").Preload("AuthKey.User").Preload("User").Where(&Machine{UserID: userID}).Find(&machines).Error; err != nil {
		return nil, err
	}

	return machines, nil
}

// SetMachineUser assigns a Machine to a user.
func (h *Mirage) SetMachineUser(machine *Machine, username, orgName, provider string) error {
	/*
		err := CheckForFQDNRules(username)
		if err != nil {
			return err
		}
	*/
	user, err := h.GetUser(username, orgName, provider)
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
		Domain:        n.Organization.MagicDnsDomain,
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
		Domain:        n.Organization.MagicDnsDomain,
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

	nameserversStr := me.Organization.Nameservers

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

	if me.Organization.OverrideLocal {
		dnsConfig.Resolvers = resolvers
	} else {
		dnsConfig.FallbackResolvers = resolvers
	}

	//cgao6: split DNS related here
	dnsConfig.Routes = make(map[string][]*dnstype.Resolver)
	domains := []string{}
	restrictedDNS := me.Organization.SplitDns
	for _, oneRestrictedDNS := range restrictedDNS {
		restrictedResolvers := make(
			[]*dnstype.Resolver,
			len(oneRestrictedDNS.NS),
		)
		for index, nameserverStr := range oneRestrictedDNS.NS {
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
		dnsConfig.Routes[oneRestrictedDNS.Domain] = restrictedResolvers
		domains = append(domains, oneRestrictedDNS.Domain)
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

	dnsConfig.Proxied = me.Organization.EnableMagic

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

	baseDomain := me.Organization.MagicDnsDomain

	return dnsConfig, baseDomain
}

func (h *Mirage) UpdateDNSConfig(user *User, newDNSCfg DNSData) error {
	if user == nil || user.Organization.ID == 0 {
		return ErrOrgNotFound
	}
	org := &user.Organization
	return h.UpdateOrgDNSConfig(org, newDNSCfg)
}

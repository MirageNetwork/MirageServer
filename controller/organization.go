package controller

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-diceware/diceware"
	"gorm.io/gorm"
	"tailscale.com/tailcfg"
)

const (
	ErrOrgExists          = Error("Organization already exists")
	ErrOrgNotFound        = Error("Organization not found")
	ErrCreateOrgParams    = Error("Invalid create organization paramters")
	ErrGetOrgParams       = Error("Invalid get organization paramters")
	ErrDeleteOrgFailed    = Error("Delete Organization Failed")
	ErrDeleteOrgCancelled = Error("Delete Organization Cancelled")
	DefaultExpireTime     = 180
)

type Organization struct {
	ID             int64  `gorm:"primary_key;unique;not null"`
	StableID       string `gorm:"unique"`
	Name           string `gorm:"uniqueIndex:idx_name_provider"`
	Provider       string `gorm:"uniqueIndex:idx_name_provider"`
	ExpiryDuration uint   `gorm:"default:180"`
	EnableMagic    bool   `gorm:"default:false"`
	MagicDnsDomain string
	OverrideLocal  bool `gorm:"default:false"`
	Nameservers    StringList
	SplitDns       SplitDNS
	AclPolicy      *ACLPolicy
	AclRules       []tailcfg.FilterRule `gorm:"-"`
	SshPolicy      *tailcfg.SSHPolicy   `gorm:"-"`
	NaviBanList    NaviBanList
	NaviDeployKey  string
	NaviDeployPub  string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type NaviBanList map[int]struct{}

func (nbl NaviBanList) Value() (driver.Value, error) {
	b, err := json.Marshal(nbl)
	return string(b), err
}

func (nbl *NaviBanList) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, nbl)
	case string:
		return json.Unmarshal([]byte(v), nbl)
	default:
		return fmt.Errorf("cannot parse admin credential: unexpected data type %T", value)
	}
}

func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == 0 {
		flakeID, err := snowflake.NewNode(1)
		if err != nil {
			return err
		}
		id := flakeID.Generate().Int64()
		o.ID = id
	}
	o.StableID = GetShortId(o.ID)
	return nil
}

func (m *Mirage) GenNewMagicDNSDomain(tx *gorm.DB) (string, error) {
	list, err := diceware.Generate(2)
	if err != nil {
		log.Error().Err(err).Msg("Could not generate passphrase")
		return "", err
	}
	tmpMagicDNSDomain := strings.Join(list, "-") + "." + m.cfg.BaseDomain
	for {
		if errors.Is(tx.First(&Organization{}, "magic_dns_domain = ?", tmpMagicDNSDomain).Error, gorm.ErrRecordNotFound) {
			break
		}
		list, err = diceware.Generate(2)
		if err != nil {
			log.Error().Err(err).Msg("Could not generate passphrase")
			return "", err
		}
		tmpMagicDNSDomain = strings.Join(list, "-") + "." + m.cfg.BaseDomain
	}
	return tmpMagicDNSDomain, nil
}

func (m *Mirage) UpdateMagicDNSDomain(orgID int64, netMagicDomain string) error {
	org, err := m.GetOrgnaizationByID(orgID)
	if err != nil {
		return err
	}
	org.MagicDnsDomain = netMagicDomain
	err = m.db.Save(org).Error
	if err != nil {
		return err
	}
	m.setOrgLastStateChangeToNow(orgID)
	return nil
}

func (m *Mirage) CreateOrgnaizationInTx(tx *gorm.DB, name, provider string) (*Organization, error) {
	if len(name) == 0 || len(provider) == 0 {
		return nil, ErrCreateOrgParams
	}
	var count int64
	if err := tx.Model(&Organization{}).Where("name = ? and provider = ?", name, provider).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrOrgExists
	}
	org := Organization{}
	org.Name = name
	org.Provider = provider
	org.ExpiryDuration = DefaultExpireTime
	org.AclPolicy = &ACLPolicy{
		Groups:    make(Groups, 0),
		Hosts:     make(Hosts, 0),
		TagOwners: make(TagOwners, 0),
		ACLs:      make([]ACL, 0),
		Tests:     make([]ACLTest, 0),
		AutoApprovers: AutoApprovers{
			Routes:   make(map[string][]string, 0),
			ExitNode: make([]string, 0),
		},
		SSHs: make([]SSH, 0),
	}

	//cgao6: 添加组织幻域域名roll生成
	newMagicDNSDomain, err := m.GenNewMagicDNSDomain(tx)
	if err != nil {
		log.Error().
			Str("func", "CreateOrgnaization").
			Err(err).
			Msg("Could not generate magic dns domain")
		return nil, err
	}
	org.MagicDnsDomain = newMagicDNSDomain

	if err := tx.Create(&org).Error; err != nil {
		log.Error().
			Str("func", "CreateOrgnaization").
			Err(err).
			Msg("Could not create row")

		return nil, err
	}
	return &org, nil
}

// GetOrgnaizationRecordByName get Organization Info only(not to update the AclRules)
func (m *Mirage) GetOrgnaizationRecordByName(name, provider string) (*Organization, error) {
	var org Organization
	err := m.db.Model(&Organization{}).Where(&Organization{
		Name:     name,
		Provider: provider,
	}).Take(&org).Error
	return &org, err
}

// GetOrgnaizationIDByName get Organization id (the primary key of the db table)
func (m *Mirage) GetOrgnaizationIDByName(name, provider string) (int64, error) {
	var id int64
	err := m.db.Model(&Organization{}).Where(&Organization{
		Name:     name,
		Provider: provider,
	}).Take(&id).Error
	return id, err
}

// GetOrgnaizationByID get Organization Info and update the AclRules
func (m *Mirage) GetOrgnaizationByID(id int64) (*Organization, error) {
	org := &Organization{}
	err := m.db.Where(&Organization{ID: id}).Take(org).Error
	if err != nil {
		return nil, err
	}
	//m.UpdateACLRulesOfOrg(org)
	return org, err
}

// ListOrgnaizations List all the organizations in the database, but it not to generate acl rules
func (m *Mirage) ListOrgnaizations() ([]Organization, error) {
	var orgs []Organization
	err := m.db.Find(&orgs).Error
	if err != nil {
		return nil, err
	}
	return orgs, err
}

func GetOrgnaizationByNameInTx(tx *gorm.DB, name, provider string) (*Organization, error) {
	if len(name) == 0 || len(provider) == 0 {
		return nil, ErrGetOrgParams
	}
	var org Organization
	if err := tx.Where("name = ? and provider = ?", name, provider).Take(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = ErrOrgNotFound
		}
		log.Debug().
			Str("func", "GetOrgnaizationByName").
			Err(err).
			Msg("Could not get row")
		return nil, err
	}

	return &org, nil
}

// 删除组织
// 输入: 组织名 provider名 是否强行删除(删除用户下的machine)
// 返回: 删除前用户数 删除后用户数 错误
func (m *Mirage) DestroyOrgnaization(orgName, provider string, force bool) (int, int, error) {
	var orgID int64
	m.db.Model(&Organization{}).Select("ID").Where("name = ? and provider = ?", orgName, provider).Find(&orgID)
	if orgID == 0 {
		return 0, 0, ErrOrgNotFound
	}

	var deleteFn func(int64) error
	if !force {
		deleteFn = m.DestroyUserByID
	} else {
		deleteFn = m.ForceDestroyUserByID
	}
	//逐个删除用户,遇到删除失败的继续删除下一个
	var userIDs []int64
	before := len(userIDs)
	m.db.Model(&User{}).Select("ID").Where(&User{OrganizationID: orgID}).Find(&userIDs)
	for _, uid := range userIDs {
		deleteFn(uid)
	}
	//检查是否还有用户
	if before > 0 {
		userIDs = []int64{}
		m.db.Model(&User{}).Select("ID").Where(&User{OrganizationID: orgID}).Find(&userIDs)
	}
	after := len(userIDs)
	//没有用户则删除组织
	if after == 0 {
		err := m.db.Unscoped().Delete(&Organization{}, orgID).Error
		if err != nil {
			return before, after, errors.Join(ErrDeleteOrgFailed, err)
		}
	} else {
		return before, after, ErrDeleteOrgCancelled
	}
	return before, after, nil
}

func (m *Mirage) UpdateOrgExpiry(user *User, newDuration uint) error {
	err := m.db.Select("expiry_duration").Updates(&Organization{
		ID:             user.OrganizationID,
		ExpiryDuration: newDuration,
	}).Error
	return err
}

func (m *Mirage) UpdateOrgDNSConfig(org *Organization, newDNSCfg DNSData) error {

	org.EnableMagic = newDNSCfg.MagicDNS
	org.Nameservers = make([]string, 0)
	if len(newDNSCfg.Resolvers) > 0 {
		org.OverrideLocal = true
		org.Nameservers = newDNSCfg.Resolvers
	} else if len(newDNSCfg.FallbackResolvers) > 0 {
		org.OverrideLocal = false
		org.Nameservers = newDNSCfg.FallbackResolvers
	}
	newSplitDns := SplitDNS{}
	for _, domain := range newDNSCfg.Domains {
		if ns, ok := newDNSCfg.Routes[domain]; ok {
			newSplitDns = append(newSplitDns, SplitDNSItem{
				Domain: domain,
				NS:     ns,
			})
		}
	}

	org.SplitDns = newSplitDns
	err := m.db.Select("EnableMagic", "Nameservers", "OverrideLocal", "Nameservers", "OverrideLocal", "SplitDns").Updates(org).Error
	return err
}

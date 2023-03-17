package controller

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"tailscale.com/tailcfg"
)

const (
	ErrOrgExists      = Error("Organization already exists")
	ErrOrgNotFound    = Error("Organization not found")
	DefaultExpireTime = 180
)

type Organization struct {
	ID             int64  `gorm:"primary_key;unique;not null"`
	StableID       string `gorm:"unique"`
	Name           string `gorm:"unique"`
	ExpiryDuration uint   `gorm:"default:180"`
	EnableMagic    bool   `gorm:"default:false"`
	OverrideLocal  bool   `gorm:"default:false"`
	Nameservers    StringList
	SplitDns       SplitDNS
	AclPolicy      *ACLPolicy
	AclRules       []tailcfg.FilterRule `gorm:"-"`
	SshPolicy      *tailcfg.SSHPolicy   `gorm:"-"`

	CreatedAt time.Time
	UpdatedAt time.Time
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

func (m *Mirage) CreateOrgnaization(name string, acl *ACLPolicy) (*Organization, error) {
	tx := m.db.Session(&gorm.Session{})
	return CreateOrgnaizationInTx(tx, name, acl)
}

func CreateOrgnaizationInTx(tx *gorm.DB, name string, acl *ACLPolicy) (*Organization, error) {
	var count int64
	if err := tx.Model(&Organization{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrOrgExists
	}
	org := Organization{}
	org.Name = name
	org.AclPolicy = acl
	org.ExpiryDuration = DefaultExpireTime
	if err := tx.Create(&org).Error; err != nil {
		log.Error().
			Str("func", "CreateOrgnaization").
			Err(err).
			Msg("Could not create row")

		return nil, err
	}
	org.AclRules = tailcfg.FilterAllowAll // default allowall
	org.SshPolicy = &tailcfg.SSHPolicy{}
	return &org, nil
}

func (m *Mirage) GetOrgnaizationByName(name string) (*Organization, error) {
	org, hit, err := GetOrgnaizationByNameInTx(m.db.Session(&gorm.Session{}), name, m.organizationCache)
	if err != nil {
		return nil, err
	}
	if org != nil && !hit {
		m.UpdateACLRulesOfOrg(org)
		m.organizationCache.Set(name, *org, -1)
	}
	return org, err
}

func GetOrgnaizationByNameInTx(tx *gorm.DB, name string, c *cache.Cache) (*Organization, bool, error) {
	oc, ok := c.Get(name)
	if oc != nil && ok {
		if val, ok := oc.(Organization); ok {
			return &val, true, nil
		}
	}
	var org Organization
	if err := tx.Where("name = ?", name).Take(&org).Error; err != nil {
		log.Error().
			Str("func", "GetOrgnaizationByName").
			Err(err).
			Msg("Could not get row")
		return nil, false, err
	}
	org.AclRules = tailcfg.FilterAllowAll // default allowall
	org.SshPolicy = &tailcfg.SSHPolicy{}

	return &org, false, nil
}

func (m *Mirage) DestroyOrgnaization(orgName string) error {
	tx := m.db.Session(&gorm.Session{})
	err := DestroyOrgnaizationInTx(tx, orgName)
	if err == nil {
		m.organizationCache.Delete(orgName)
	}
	return err
}

func DestroyOrgnaizationInTx(tx *gorm.DB, orgName string) error {
	var count int64
	tx.Model(&Organization{}).Where("name = ?", orgName).Count(&count)
	if count == 0 {
		return ErrOrgNotFound
	}
	if result := tx.Unscoped().Delete(&Organization{Name: orgName}); result.Error != nil {
		return result.Error
	}
	return nil
}

func (m *Mirage) UpdateOrgExpiry(user *User, newDuration uint) error {
	err := m.db.Select("expiry_duration").Updates(&Organization{
		ID:             user.OrgId,
		ExpiryDuration: newDuration,
	}).Error
	if err == nil {
		m.organizationCache.Delete(user.OrgName)
	}
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
	org.SplitDns = newDNSCfg.Routes
	err := m.db.Select("EnableMagic", "Nameservers", "OverrideLocal", "Nameservers", "OverrideLocal", "SplitDns").Updates(org).Error
	if err == nil {
		m.organizationCache.Delete(org.Name)
	}
	return err
}

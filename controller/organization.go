package controller

import (
	"errors"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-diceware/diceware"
	"gorm.io/gorm"
	"tailscale.com/tailcfg"
)

const (
	ErrOrgExists       = Error("Organization already exists")
	ErrOrgNotFound     = Error("Organization not found")
	ErrCreateOrgParams = Error("Invalid create organization paramters")
	ErrGetOrgParams    = Error("Invalid get organization paramters")
	DefaultExpireTime  = 180
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
	m.setLastStateChangeToNow()
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
	org.AclPolicy = &ACLPolicy{
		ACLs: []ACL{{
			Action:       "accept",
			Protocol:     "",
			Sources:      []string{"*"},
			Destinations: []string{"*:*"},
		}},
	}
	org.ExpiryDuration = DefaultExpireTime

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

func (m *Mirage) GetOrgnaizationByName(name, provider string) (*Organization, error) {
	org, err := GetOrgnaizationByNameInTx(m.db.Session(&gorm.Session{}), name, provider)
	if err != nil {
		return nil, err
	}
	m.UpdateACLRulesOfOrg(org)
	return org, err
}

func (m *Mirage) GetOrgnaizationIDByName(name, provider string) (int64, error) {
	var id int64
	err := m.db.Model(&Organization{}).Where(&Organization{
		Name:     name,
		Provider: provider,
	}).Take(&id).Error
	return id, err
}

func (m *Mirage) GetOrgnaizationByID(id int64) (*Organization, error) {
	org := &Organization{}
	err := m.db.Where(&Organization{ID: id}).Take(org).Error
	if err != nil {
		return nil, err
	}
	m.UpdateACLRulesOfOrg(org)
	return org, err
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

func (m *Mirage) DestroyOrgnaization(orgName, provider string) error {
	tx := m.db.Session(&gorm.Session{})
	err := DestroyOrgnaizationInTx(tx, orgName, provider)
	return err
}

func DestroyOrgnaizationInTx(tx *gorm.DB, orgName, provider string) error {
	var count int64
	tx.Model(&Organization{}).Where("name = ? and provider = ?", orgName, provider).Count(&count)
	if count == 0 {
		return ErrOrgNotFound
	}
	if result := tx.Unscoped().Delete(&Organization{Name: orgName, Provider: provider}); result.Error != nil {
		return result.Error
	}
	return nil
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
	org.SplitDns = newDNSCfg.Routes
	err := m.db.Select("EnableMagic", "Nameservers", "OverrideLocal", "Nameservers", "OverrideLocal", "SplitDns").Updates(org).Error
	return err
}

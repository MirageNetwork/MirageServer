package controller

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
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

func (m *Mirage) CreateOrgnaization(name string) (*Organization, error) {
	tx := m.db.Session(&gorm.Session{})
	return CreateOrgnaizationInTx(tx, name)
}

func CreateOrgnaizationInTx(tx *gorm.DB, name string) (*Organization, error) {
	var count int64
	if err := tx.Model(&Organization{}).Where("name = ?", name).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrOrgExists
	}
	org := Organization{}
	org.Name = name
	org.ExpiryDuration = DefaultExpireTime
	if err := tx.Create(&org).Error; err != nil {
		log.Error().
			Str("func", "CreateOrgnaization").
			Err(err).
			Msg("Could not create row")

		return nil, err
	}

	return &org, nil
}

func (m *Mirage) GetOrgnaizationByName(name string) (*Organization, error) {
	return GetOrgnaizationByNameInTx(m.db.Session(&gorm.Session{}), name)
}

func GetOrgnaizationByNameInTx(tx *gorm.DB, name string) (*Organization, error) {
	var org Organization
	if err := tx.Where("name = ?", name).Take(&org).Error; err != nil {
		log.Error().
			Str("func", "GetOrgnaizationByName").
			Err(err).
			Msg("Could not get row")
		return nil, err
	}
	return &org, nil
}

func (m *Mirage) DestroyOrgnaization(orgId int64) error {
	tx := m.db.Session(&gorm.Session{})
	return DestroyOrgnaizationInTx(tx, orgId)
}

func DestroyOrgnaizationInTx(tx *gorm.DB, orgId int64) error {
	var count int64
	tx.Model(&Organization{}).Where("id = ?", orgId).Count(&count)
	if count == 0 {
		return ErrOrgNotFound
	}
	if result := tx.Unscoped().Delete(&Organization{ID: orgId}); result.Error != nil {
		return result.Error
	}
	return nil
}

func (m *Mirage) UpdateOrgExpiry(orgId int64, newDuration uint) error {
	return m.db.Select("expiry_duration").Updates(&Organization{
		ID:             orgId,
		ExpiryDuration: newDuration,
	}).Error
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
	return m.db.Select("EnableMagic", "Nameservers", "OverrideLocal", "Nameservers", "OverrideLocal", "SplitDns").Updates(org).Error
}

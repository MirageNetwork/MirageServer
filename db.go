package Mirage

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"tailscale.com/tailcfg"
)

const (
	dbVersion = "1"

	errValueNotFound     = Error("not found")
	ErrCannotParsePrefix = Error("cannot parse prefix")
)

// KV is a key-value store in a psql table. For future use...
type KV struct {
	Key   string
	Value string
}

func (h *Mirage) initDB() error {
	db, err := h.openDB()
	if err != nil {
		return err
	}
	h.db = db

	if h.dbType == Postgres {
		db.Exec(`create extension if not exists "uuid-ossp";`)
	}

	_ = db.Migrator().RenameTable("namespaces", "users")

	err = db.AutoMigrate(&User{})
	if err != nil {
		return err
	}

	_ = db.Migrator().RenameColumn(&Machine{}, "namespace_id", "user_id")
	_ = db.Migrator().RenameColumn(&PreAuthKey{}, "namespace_id", "user_id")

	_ = db.Migrator().RenameColumn(&Machine{}, "ip_address", "ip_addresses")
	_ = db.Migrator().RenameColumn(&Machine{}, "name", "hostname")

	err = db.AutoMigrate(&Route{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&Machine{})
	if err != nil {
		return err
	}

	if db.Migrator().HasColumn(&Machine{}, "given_name") {
		machines := Machines{}
		if err := h.db.Find(&machines).Error; err != nil {
			log.Error().Err(err).Msg("Error accessing db")
		}

		for item, machine := range machines {
			if machine.GivenName == "" {
				normalizedHostname, err := NormalizeToFQDNRules(
					machine.Hostname,
					h.cfg.OIDC.StripEmaildomain,
				)
				if err != nil {
					log.Error().
						Caller().
						Str("hostname", machine.Hostname).
						Err(err).
						Msg("Failed to normalize machine hostname in DB migration")
				}

				err = h.RenameMachine(&machines[item], normalizedHostname)
				if err != nil {
					log.Error().
						Caller().
						Str("hostname", machine.Hostname).
						Err(err).
						Msg("Failed to save normalized machine name in DB migration")
				}
			}
		}
	}

	err = db.AutoMigrate(&KV{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&PreAuthKey{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&PreAuthKeyACLTag{})
	if err != nil {
		return err
	}

	_ = db.Migrator().DropTable("shared_machines")

	err = h.setValue("db_version", dbVersion)

	return err
}

func (h *Mirage) openDB() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	var log logger.Interface
	if h.dbDebug {
		log = logger.Default
	} else {
		log = logger.Default.LogMode(logger.Silent)
	}

	switch h.dbType {
	case Sqlite:
		db, err = gorm.Open(
			sqlite.Open(h.dbString+"?_synchronous=1&_journal_mode=WAL"),
			&gorm.Config{
				DisableForeignKeyConstraintWhenMigrating: true,
				Logger:                                   log,
			},
		)

		db.Exec("PRAGMA foreign_keys=ON")

		// The pure Go SQLite library does not handle locking in
		// the same way as the C based one and we cant use the gorm
		// connection pool as of 2022/02/23.
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetConnMaxIdleTime(time.Hour)

	case Postgres:
		db, err = gorm.Open(postgres.Open(h.dbString), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger:                                   log,
		})
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}

// getValue returns the value for the given key in KV.
func (h *Mirage) getValue(key string) (string, error) {
	var row KV
	if result := h.db.First(&row, "key = ?", key); errors.Is(
		result.Error,
		gorm.ErrRecordNotFound,
	) {
		return "", errValueNotFound
	}

	return row.Value, nil
}

// setValue sets value for the given key in KV.
func (h *Mirage) setValue(key string, value string) error {
	keyValue := KV{
		Key:   key,
		Value: value,
	}

	if _, err := h.getValue(key); err == nil {
		h.db.Model(&keyValue).Where("key = ?", key).Update("value", value)

		return nil
	}

	if err := h.db.Create(keyValue).Error; err != nil {
		return fmt.Errorf("failed to create key value pair in the database: %w", err)
	}

	return nil
}

func (h *Mirage) pingDB(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	db, err := h.db.DB()
	if err != nil {
		return err
	}

	return db.PingContext(ctx)
}

// This is a "wrapper" type around tailscales
// Hostinfo to allow us to add database "serialization"
// methods. This allows us to use a typed values throughout
// the code and not have to marshal/unmarshal and error
// check all over the code.
type HostInfo tailcfg.Hostinfo

func (hi *HostInfo) Scan(destination interface{}) error {
	switch value := destination.(type) {
	case []byte:
		return json.Unmarshal(value, hi)

	case string:
		return json.Unmarshal([]byte(value), hi)

	default:
		return fmt.Errorf("%w: unexpected data type %T", ErrMachineAddressesInvalid, destination)
	}
}

// Value return json value, implement driver.Valuer interface.
func (hi HostInfo) Value() (driver.Value, error) {
	bytes, err := json.Marshal(hi)

	return string(bytes), err
}

type IPPrefix netip.Prefix

func (i *IPPrefix) Scan(destination interface{}) error {
	switch value := destination.(type) {
	case string:
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return err
		}
		*i = IPPrefix(prefix)

		return nil
	default:
		return fmt.Errorf("%w: unexpected data type %T", ErrCannotParsePrefix, destination)
	}
}

// Value return json value, implement driver.Valuer interface.
func (i IPPrefix) Value() (driver.Value, error) {
	prefixStr := netip.Prefix(i).String()

	return prefixStr, nil
}

type IPPrefixes []netip.Prefix

func (i *IPPrefixes) Scan(destination interface{}) error {
	switch value := destination.(type) {
	case []byte:
		return json.Unmarshal(value, i)

	case string:
		return json.Unmarshal([]byte(value), i)

	default:
		return fmt.Errorf("%w: unexpected data type %T", ErrMachineAddressesInvalid, destination)
	}
}

// Value return json value, implement driver.Valuer interface.
func (i IPPrefixes) Value() (driver.Value, error) {
	bytes, err := json.Marshal(i)

	return string(bytes), err
}

type StringList []string

func (i *StringList) Scan(destination interface{}) error {
	switch value := destination.(type) {
	case []byte:
		return json.Unmarshal(value, i)

	case string:
		return json.Unmarshal([]byte(value), i)

	default:
		return fmt.Errorf("%w: unexpected data type %T", ErrMachineAddressesInvalid, destination)
	}
}

// Value return json value, implement driver.Valuer interface.
func (i StringList) Value() (driver.Value, error) {
	bytes, err := json.Marshal(i)

	return string(bytes), err
}

// cgao6: add splitdns type to store dns config into user's db
type SplitDNS map[string][]string

func (i *SplitDNS) Scan(destination interface{}) error {
	switch value := destination.(type) {
	case []byte:
		return json.Unmarshal(value, i)

	case string:
		return json.Unmarshal([]byte(value), i)

	default:
		return fmt.Errorf("%w: unexpected data type %T", ErrMachineAddressesInvalid, destination)
	}
}

// Value return json value, implement driver.Valuer interface.
func (i SplitDNS) Value() (driver.Value, error) {
	bytes, err := json.Marshal(i)

	return string(bytes), err
}

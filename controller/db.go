package controller

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/netip"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"tailscale.com/tailcfg"
)

type DataPool struct {
	db *gorm.DB
}

const (
	dbVersion = "1"

	errValueNotFound     = Error("not found")
	ErrCannotParsePrefix = Error("cannot parse prefix")
)

func (dp *DataPool) DB() *gorm.DB {
	return dp.db
}

func (dp *DataPool) InitCockpitDB() error {
	err := dp.db.AutoMigrate(&SysConfig{})
	if err != nil {
		return err
	}

	err = dp.db.AutoMigrate(&NaviRegion{})
	if err != nil {
		return err
	}
	err = dp.db.AutoMigrate(&NaviNode{})
	if err != nil {
		return err
	}

	return err
}

func (dp *DataPool) InitMirageDB() error {
	err := dp.db.AutoMigrate(&User{})
	if err != nil {
		return err
	}

	err = dp.db.AutoMigrate(&Route{})
	if err != nil {
		return err
	}

	err = dp.db.AutoMigrate(&Machine{})
	if err != nil {
		return err
	}

	err = dp.db.AutoMigrate(&PreAuthKey{})
	if err != nil {
		return err
	}

	/*
		err = dp.db.AutoMigrate(&PreAuthKeyACLTag{})
		if err != nil {
			return err
		}
	*/
	err = dp.db.AutoMigrate(&Organization{})
	if err != nil {
		return err
	}

	return err
}

func (dp *DataPool) OpenDB() error {
	var log logger.Interface
	log = logger.Default.LogMode(logger.Silent)

	db, err := gorm.Open(
		sqlite.Open(AbsolutePathFromConfigPath(DatabasePath)+"?_synchronous=1&_journal_mode=WAL"),
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

	if err != nil {
		return err
	}
	dp.db = db

	return nil
}

func (dp *DataPool) pingDB(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	db, err := dp.db.DB()
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

func (i IPPrefix) String() string {
	return netip.Prefix(i).String()
}

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
type SplitDNS []SplitDNSItem

type SplitDNSItem struct {
	Domain string   `json:"domain"`
	NS     []string `json:"ns"`
}

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

// ACLPolicy struct to json implement
func (a *ACLPolicy) Scan(destination interface{}) error {
	switch value := destination.(type) {
	case []byte:
		return json.Unmarshal(value, a)

	case string:
		return json.Unmarshal([]byte(value), a)

	default:
		return fmt.Errorf("%w: unexpected data type %T", ErrMachineAddressesInvalid, destination)
	}
}

// Value return json value, implement driver.Valuer interface.
func (a ACLPolicy) Value() (driver.Value, error) {
	bytes, err := json.Marshal(a)

	return string(bytes), err
}

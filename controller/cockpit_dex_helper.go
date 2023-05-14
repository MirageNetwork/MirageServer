package controller

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dexidp/dex/pkg/log"
	"github.com/dexidp/dex/server"
	"github.com/dexidp/dex/storage"
	"github.com/dexidp/dex/storage/ent"
	"github.com/dexidp/dex/storage/etcd"
	"github.com/dexidp/dex/storage/kubernetes"
	"github.com/dexidp/dex/storage/memory"
	"github.com/dexidp/dex/storage/sql"
	"github.com/sirupsen/logrus"
)

// Storage holds app's storage configuration.
type DexStorage struct {
	Type   string           `json:"type"`
	Config DexStorageConfig `json:"config"`
}

// StorageConfig is a configuration that can create a storage.
type DexStorageConfig interface {
	Open(logger log.Logger) (storage.Storage, error)
}

var (
	_ DexStorageConfig = (*etcd.Etcd)(nil)
	_ DexStorageConfig = (*kubernetes.Config)(nil)
	_ DexStorageConfig = (*memory.Config)(nil)
	_ DexStorageConfig = (*sql.SQLite3)(nil)
	_ DexStorageConfig = (*sql.Postgres)(nil)
	_ DexStorageConfig = (*sql.MySQL)(nil)
	_ DexStorageConfig = (*ent.SQLite3)(nil)
	_ DexStorageConfig = (*ent.Postgres)(nil)
	_ DexStorageConfig = (*ent.MySQL)(nil)
)

func getORMBasedSQLStorage(normal, entBased DexStorageConfig) func() DexStorageConfig {
	return func() DexStorageConfig {
		switch os.Getenv("DEX_ENT_ENABLED") {
		case "true", "yes":
			return entBased
		default:
			return normal
		}
	}
}

var dexstorages = map[string]func() DexStorageConfig{
	"etcd":       func() DexStorageConfig { return new(etcd.Etcd) },
	"kubernetes": func() DexStorageConfig { return new(kubernetes.Config) },
	"memory":     func() DexStorageConfig { return new(memory.Config) },
	"sqlite3":    getORMBasedSQLStorage(&sql.SQLite3{}, &ent.SQLite3{}),
	"postgres":   getORMBasedSQLStorage(&sql.Postgres{}, &ent.Postgres{}),
	"mysql":      getORMBasedSQLStorage(&sql.MySQL{}, &ent.MySQL{}),
}

// isExpandEnvEnabled returns if os.ExpandEnv should be used for each storage and connector config.
// Disabling this feature avoids surprises e.g. if the LDAP bind password contains a dollar character.
// Returns false if the env variable "DEX_EXPAND_ENV" is a falsy string, e.g. "false".
// Returns true if the env variable is unset or a truthy string, e.g. "true", or can't be parsed as bool.
func isExpandEnvEnabled() bool {
	enabled, err := strconv.ParseBool(os.Getenv("DEX_EXPAND_ENV"))
	if err != nil {
		// Unset, empty string or can't be parsed as bool: Default = true.
		return true
	}
	return enabled
}

// UnmarshalJSON allows Storage to implement the unmarshaler interface to
// dynamically determine the type of the storage config.
func (s *DexStorage) UnmarshalJSON(b []byte) error {
	var store struct {
		Type   string          `json:"type"`
		Config json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(b, &store); err != nil {
		return fmt.Errorf("parse storage: %v", err)
	}
	f, ok := dexstorages[store.Type]
	if !ok {
		return fmt.Errorf("unknown storage type %q", store.Type)
	}

	storageConfig := f()
	if len(store.Config) != 0 {
		data := []byte(store.Config)
		if isExpandEnvEnabled() {
			// Caution, we're expanding in the raw JSON/YAML source. This may not be what the admin expects.
			data = []byte(os.ExpandEnv(string(store.Config)))
		}
		if err := json.Unmarshal(data, storageConfig); err != nil {
			return fmt.Errorf("parse storage config: %v", err)
		}
	}
	*s = DexStorage{
		Type:   store.Type,
		Config: storageConfig,
	}
	return nil
}

// Connector is a magical type that can unmarshal YAML dynamically. The
// Type field determines the connector type, which is then customized for Config.
type Connector struct {
	Type string `json:"type"`
	Name string `json:"name"`
	ID   string `json:"id"`

	Config server.ConnectorConfig `json:"config"`
}

// UnmarshalJSON allows Connector to implement the unmarshaler interface to
// dynamically determine the type of the connector config.
func (c *Connector) UnmarshalJSON(b []byte) error {
	var conn struct {
		Type string `json:"type"`
		Name string `json:"name"`
		ID   string `json:"id"`

		Config json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(b, &conn); err != nil {
		return fmt.Errorf("parse connector: %v", err)
	}
	f, ok := server.ConnectorsConfig[conn.Type]
	if !ok {
		return fmt.Errorf("unknown connector type %q", conn.Type)
	}

	connConfig := f()
	if len(conn.Config) != 0 {
		data := []byte(conn.Config)
		if isExpandEnvEnabled() {
			// Caution, we're expanding in the raw JSON/YAML source. This may not be what the admin expects.
			data = []byte(os.ExpandEnv(string(conn.Config)))
		}
		if err := json.Unmarshal(data, connConfig); err != nil {
			return fmt.Errorf("parse connector config: %v", err)
		}
	}
	*c = Connector{
		Type:   conn.Type,
		Name:   conn.Name,
		ID:     conn.ID,
		Config: connConfig,
	}
	return nil
}

// ToStorageConnector converts an object to storage connector type.
func ToStorageConnector(c Connector) (storage.Connector, error) {
	data, err := json.Marshal(c.Config)
	if err != nil {
		return storage.Connector{}, fmt.Errorf("failed to marshal connector config: %v", err)
	}

	return storage.Connector{
		ID:     c.ID,
		Type:   c.Type,
		Name:   c.Name,
		Config: data,
	}, nil
}

type utcFormatter struct {
	f logrus.Formatter
}

func (f *utcFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return f.f.Format(e)
}
func newLogger(level string, format string) (log.Logger, error) {
	var logLevel logrus.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = logrus.DebugLevel
	case "", "info":
		logLevel = logrus.InfoLevel
	case "error":
		logLevel = logrus.ErrorLevel
	default:
		return nil, fmt.Errorf("log level is not one of the supported values : %s", level)
	}

	var formatter utcFormatter
	switch strings.ToLower(format) {
	case "", "text":
		formatter.f = &logrus.TextFormatter{DisableColors: true}
	case "json":
		formatter.f = &logrus.JSONFormatter{}
	default:
		return nil, fmt.Errorf("log format is not one of the supported values : %s", format)
	}

	return &logrus.Logger{
		Out:       os.Stderr,
		Formatter: &formatter,
		Level:     logLevel,
	}, nil
}

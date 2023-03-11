package conf

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path + "/config.yaml")
	if err != nil {
		f, err = os.Open(path + "/config.example.yaml")
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()

	var cfg *Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

type Config struct {
	BaseURL    string
	JWT        JWTConfig
	Persistent DB
	Databases  map[string]DB
	Providers  Providers
}

func (cfg *Config) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		BaseURL    string        `yaml:"baseUrl"`
		Persistent string        `yaml:"persistent"`
		Databases  map[string]DB `yaml:"databases"`
		JWT        JWTConfig     `yaml:"jwt"`
		Providers  Providers     `yaml:"providers"`
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	cfg.BaseURL = raw.BaseURL
	cfg.Databases = raw.Databases

	db, ok := raw.Databases[raw.Persistent]
	if !ok {
		return errors.New("db not found")
	}

	cfg.Persistent = db

	cfg.JWT = raw.JWT
	cfg.Providers = raw.Providers

	return nil
}

type JWTConfig struct {
	Secret  string
	Timeout time.Duration
	Refresh struct {
		Enabled bool
		Maximum time.Duration
	}
}

func (cfg *JWTConfig) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Secret  string
		Timeout string
		Refresh struct {
			Enabled bool
			Maximum string
		}
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	cfg.Secret = raw.Secret

	if raw.Timeout == "" {
		cfg.Timeout = 1 * time.Hour
	} else {
		timeout, err := time.ParseDuration(raw.Timeout)
		if err != nil {
			return err
		}

		cfg.Timeout = timeout
	}

	cfg.Refresh.Enabled = raw.Refresh.Enabled
	if !raw.Refresh.Enabled {
		cfg.Refresh.Maximum = 0
	} else {

		if raw.Refresh.Maximum == "" {
			cfg.Refresh.Maximum = 1 * time.Hour
		} else {
			max, err := time.ParseDuration(raw.Refresh.Maximum)
			if err != nil {
				return err
			}

			cfg.Refresh.Maximum = max
		}
	}

	return nil
}

type DBDriver int

const (
	SQLite DBDriver = iota
	// MySQL
	// PostgreSQL

	BadgerDB
	// Redis

	InMem
)

func parseDBDriver(driver string) (DBDriver, error) {
	switch driver {
	case "sqlite":
		return SQLite, nil
	case "badger":
		return BadgerDB, nil
	case "inmem":
		return InMem, nil
	default:
		return -1, errors.New("driver not supported")
	}
}

type DB struct {
	Driver   DBDriver
	Name     string
	Host     string
	Port     int
	Username string
	Password string
	InMem    bool
}

func (db *DB) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Driver   string
		Name     string
		Host     string
		Port     int
		Username string
		Password string
		InMem    bool
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	driver, err := parseDBDriver(raw.Driver)
	if err != nil {
		return err
	}

	db.Driver = driver
	db.Name = raw.Name
	db.Host = raw.Host
	db.Port = raw.Port
	db.Username = raw.Username
	db.Password = raw.Password
	db.InMem = raw.InMem

	return nil
}

type Providers struct {
	Google struct {
		Client struct {
			ID     string
			Secret string
		}
	}
}

package model

import (
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type DBDriver string

const (
	MySQL      DBDriver = "mysql"
	PostgreSQL DBDriver = "postgresql"
)

type Config struct {
	BaseURL   string
	DB        DB
	JWT       JWTConfig
	Redis     RedisConfig
	Providers Providers
	Services  struct {
		Edge EdgeConfig
	}
}

type DB struct {
	Driver   DBDriver
	Host     string
	Port     int
	Username string
	Password string
	DBName   string
}

type JWTConfig struct {
	Secret  string
	Timeout time.Duration
	Refresh struct {
		Enabled bool
		Maximum time.Duration
	}
}

func (cfg *JWTConfig) UnmarshalYAML(value yaml.Node) error {
	var tmp struct {
		Secret  string
		Timeout string
		Refresh struct {
			Enabled bool
			Maximum string
		}
	}

	if err := value.Decode(&tmp); err != nil {
		return err
	}

	cfg.Secret = tmp.Secret

	if timeout, err := time.ParseDuration(tmp.Timeout); err != nil {
		cfg.Timeout = 1 * time.Hour
	} else {
		cfg.Timeout = timeout
	}

	cfg.Refresh.Enabled = tmp.Refresh.Enabled
	if !cfg.Refresh.Enabled {
		cfg.Refresh.Maximum = 0
	} else {
		if timeout, err := time.ParseDuration(tmp.Refresh.Maximum); err != nil {
			cfg.Refresh.Maximum = 1 * time.Hour
		} else {
			cfg.Refresh.Maximum = timeout
		}
	}

	return nil
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type Providers struct {
	Google struct {
		Client struct {
			ID     string
			Secret string
		}
	}
}

func (cfg *RedisConfig) Addr() string {
	port := cfg.Port
	if port == 0 {
		port = 6379
	}
	return cfg.Host + ":" + strconv.Itoa(port)
}

type EdgeConfig struct {
	OTP OTPConfig
	JWT JWTConfig
}

type OTPConfig struct {
	Length  int
	Timeout time.Duration
}

func (cfg *OTPConfig) UnmarshalYAML(value *yaml.Node) error {
	var tmp struct {
		Length  int
		Timeout string
	}

	if err := value.Decode(&tmp); err != nil {
		return err
	}

	cfg.Length = tmp.Length

	if timeout, err := time.ParseDuration(tmp.Timeout); err != nil {
		cfg.Timeout = 1 * time.Hour
	} else {
		cfg.Timeout = timeout
	}

	return nil
}

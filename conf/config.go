package conf

import (
	"encoding/json"
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

	// TODO
	if cfg.Persistent.Host == "" {
		cfg.Persistent.Host = path
	}

	return cfg, nil
}

type Config struct {
	Name       string     `yaml:"name"`
	BaseURL    string     `yaml:"baseUrl"`
	JWT        JWT        `yaml:"jwt"`
	Persistent Persistent `yaml:"persistent"`
	EventBus   EventBus   `yaml:"eventBus"`
	Providers  Providers  `yaml:"providers"`
	Test       Test       `yaml:"test"`
}

type JWT struct {
	Secret  string
	Timeout time.Duration
	Refresh struct {
		Enabled bool
		Maximum time.Duration
	}
}

func (cfg *JWT) UnmarshalYAML(value *yaml.Node) error {
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

type PersistentDriver int

const (
	SQLite PersistentDriver = iota
	BadgerDB
	InMem
)

func ParsePersistentDriver(driver string) (PersistentDriver, error) {
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

type Persistent struct {
	Driver   PersistentDriver
	Name     string
	Host     string
	Port     int
	Username string
	Password string
	InMem    bool
}

func (p *Persistent) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Driver   string `yaml:"driver"`
		Name     string `yaml:"name"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		InMem    bool   `yaml:"inmem"`
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	driver, err := ParsePersistentDriver(raw.Driver)
	if err != nil {
		return err
	}

	p.Driver = driver
	p.Name = raw.Name
	p.Host = raw.Host
	p.Port = raw.Port
	p.Username = raw.Username
	p.Password = raw.Password
	p.InMem = raw.InMem

	return nil
}

type TransportProvider int

const NATS TransportProvider = iota

func ParseTransportProvider(provider string) (TransportProvider, error) {
	switch provider {
	case "nats":
		return NATS, nil
	default:
		return -1, errors.New("provider not supported")
	}
}

type EventBus struct {
	Provider TransportProvider
	Host     string
	Port     int
	Users    Users
}

func (e *EventBus) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Provider string `yaml:"provider"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Users    Users  `yaml:"users"`
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	provider, err := ParseTransportProvider(raw.Provider)
	if err != nil {
		return err
	}

	e.Provider = provider
	e.Host = raw.Host
	e.Port = raw.Port
	e.Users = raw.Users

	return nil
}

type Users struct {
	Stream   Stream
	Consumer Consumer
}

type Stream struct {
	Name   string
	Config json.RawMessage
}

func (s *Stream) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Name   string
		Config string
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	s.Name = raw.Name
	s.Config = json.RawMessage(raw.Config)

	return nil
}

type Consumer struct {
	Name   string
	Stream string
	Config json.RawMessage
}

func (c *Consumer) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Name   string
		Stream string
		Config string
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	c.Name = raw.Name
	c.Stream = raw.Stream
	c.Config = json.RawMessage(raw.Config)

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

type Test struct {
	Token string
}

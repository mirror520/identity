package conf

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var (
	Path string
	Port int

	global *Config
)

func G() *Config {
	if global == nil {
		panic("configuration not loaded")
	}

	return global
}

func ReplaceGlobals(cfg *Config) {
	global = cfg
}

func LoadEnv(cli *cli.Context) error {
	path := cli.String("path")
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		path = homeDir + "/.identity"
	}

	Path = path
	Port = cli.Int("port")
	return nil
}

func LoadConfig() (*Config, error) {
	f, err := os.Open(Path + "/config.yaml")
	if err != nil {
		f, err = os.Open(Path + "/config.example.yaml")
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()

	r := NewEnvExpandedReader(f)

	var cfg *Config
	if err := yaml.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

type Config struct {
	Name        string      `yaml:"name"`
	BaseURL     string      `yaml:"baseUrl"`
	JWT         JWT         `yaml:"jwt"`
	Transports  Transports  `yaml:"transports"`
	Persistence Persistence `yaml:"persistence"`
	EventBus    EventBus    `yaml:"eventBus"`
	Providers   Providers   `yaml:"providers"`
	Test        Test        `yaml:"test"`
}

type JWT struct {
	Secret  []byte
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

	cfg.Secret = []byte(raw.Secret)

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

type Transports struct {
	HTTP          RegisterHTTP  `yaml:"http"`
	NATS          RegisterNATS  `yaml:"nats"`
	LoadBalancing LoadBalancing `yaml:"loadBalancing"`
}

type RegisterHTTP struct {
	Enabled  bool
	Internal Instance
	External *Instance
}

func (r *RegisterHTTP) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Enabled  bool      `yaml:"enabled"`
		Internal Instance  `yaml:"internal"`
		External *Instance `yaml:"external"`
	}

	if err := value.Decode(&raw); err != nil {
		return err
	}

	r.Enabled = raw.Enabled
	r.Internal = raw.Internal
	r.External = raw.External

	// default
	if r.Internal.Scheme == "" {
		r.Internal.Scheme = "http"
	}

	if r.Internal.Host == "" {
		r.Internal.Host = "localhost"
	}

	if r.Internal.Port == 0 {
		r.Internal.Port = Port
	}

	return nil
}

type RegisterNATS struct {
	Enabled   bool      `yaml:"enabled"`
	Internal  Instance  `yaml:"internal"`
	External  *Instance `yaml:"external"`
	ReqPrefix string    `yaml:"reqPrefix"`
}

type LoadBalancing struct {
	Enabled bool `yaml:"enabled"`
}

type Instance struct {
	Scheme string `yaml:"scheme"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Health Health `yaml:"health"`
}

func (i *Instance) URL() string {
	return i.Scheme + "://" + i.Host + ":" + strconv.Itoa(i.Port)
}

type Health struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type PersistenceDriver int

const (
	SQLite PersistenceDriver = iota
	BadgerDB
	InMem
)

func ParsePersistenceDriver(driver string) (PersistenceDriver, error) {
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

func (driver PersistenceDriver) String() string {
	switch driver {
	case SQLite:
		return "sqlite"
	case BadgerDB:
		return "badger"
	case InMem:
		return "inmem"
	default:
		return "unknwon"
	}
}

type Persistence struct {
	Driver   PersistenceDriver
	Name     string
	Host     string
	Port     int
	Username string
	Password string
	InMem    bool
}

func (p *Persistence) UnmarshalYAML(value *yaml.Node) error {
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

	driver, err := ParsePersistenceDriver(raw.Driver)
	if err != nil {
		return err
	}

	p.Driver = driver
	p.Name = raw.Name

	p.Host = raw.Host
	if raw.Host == "" {
		p.Host = Path
	}

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

func (p TransportProvider) String() string {
	switch p {
	case NATS:
		return "nats"
	default:
		return ""
	}
}

type EventBus struct {
	Provider TransportProvider
	Users    Users
}

func (e *EventBus) UnmarshalYAML(value *yaml.Node) error {
	var raw struct {
		Provider string `yaml:"provider"`
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

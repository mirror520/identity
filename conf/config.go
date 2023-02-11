package conf

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/mirror520/identity/model"
)

func LoadConfig(path string) (*model.Config, error) {
	f, err := os.Open(path + "/config.yaml")
	if err != nil {
		f, err = os.Open(path + "/config.example.yaml")
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()

	var cfg *model.Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

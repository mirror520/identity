package conf

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("INSTANCE_NAME", "identity")

	cfg, err := LoadConfig("..")
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	assert.Equal("identity", cfg.Name)
	assert.Equal("identity.linyc.idv.tw", cfg.BaseURL)

	assert.Equal(1*time.Hour, cfg.JWT.Timeout)
	assert.True(cfg.JWT.Refresh.Enabled)
	assert.Equal(1*time.Hour+30*time.Minute, cfg.JWT.Refresh.Maximum)

	assert.Equal(BadgerDB, cfg.Persistent.Driver)
	assert.Equal("users", cfg.Persistent.Name)
}

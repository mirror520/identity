package conf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	assert := assert.New(t)

	cfg, err := LoadConfig("../..")
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	assert.Equal("api.linyc.idv.tw", cfg.BaseURL)

	assert.Equal(1*time.Hour, cfg.JWT.Timeout)
	assert.True(cfg.JWT.Refresh.Enabled)
	assert.Equal(1*time.Hour+30*time.Minute, cfg.JWT.Refresh.Maximum)

	assert.Equal(SQLite, cfg.Persistent.Driver)
	assert.Equal("identity", cfg.Persistent.Name)
}

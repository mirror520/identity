package user

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	assert := assert.New(t)

	u := NewUser("user01", "User01", "user01@example.com")
	assert.Equal("user01", u.Username)
	assert.Equal("user01@example.com", u.Email)
	assert.Equal(Registered, u.Status)

	jsonStr, err := json.MarshalIndent(u.Events(), "", "    ")
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	fmt.Println(string(jsonStr))
}

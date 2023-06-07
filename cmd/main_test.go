package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	consul "github.com/hashicorp/consul/api"

	"github.com/mirror520/identity"
	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/persistent/db"
	"github.com/mirror520/identity/user"
)

func TestServiceDiscovery(t *testing.T) {
	assert := assert.New(t)

	cfg := consul.DefaultConfig()

	client, err := consul.NewClient(cfg)
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	session, _, err := client.Session().Create(&consul.SessionEntry{
		TTL: "60s",
	}, nil)
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	query, _, err := client.PreparedQuery().Create(&consul.PreparedQueryDefinition{
		Session: session,
		Service: consul.ServiceQuery{
			Service: "identity",
			Tags:    []string{"nats"},
		},
	}, nil)
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	resp, _, err := client.PreparedQuery().Execute(query, nil)
	if err != nil {
		assert.Fail(err.Error())
		return
	}

	assert.Len(resp.Nodes, 4)
}

type identityTestSuite struct {
	suite.Suite
	svc   identity.Service
	users user.Repository
	token string
}

func (suite *identityTestSuite) SetupSuite() {
	cfg, err := conf.LoadConfig("..")
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.token = "YOUR GOOGLE JWT TOKEN" // Token 需由 Google 簽發
	if cfg.Test.Token != "" {
		suite.token = cfg.Test.Token
	}

	users, err := db.NewUserRepository(cfg.Persistent)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.svc = identity.NewService(users, cfg.Providers)
	suite.users = users
}

func (suite *identityTestSuite) TestRegister() {
	u, err := suite.svc.Register("user01", "User01", "user01@example.com")
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Equal("user01", u.Username)
	suite.Equal("user01@example.com", u.Email)
	suite.Equal(user.Registered, u.Status)

	suite.Equal(user.UserRegistered.String(), u.Events()[0].EventName())
}

func (suite *identityTestSuite) TestRegisterAndVerify() {
	u, err := suite.svc.Register("user02", "User02", "user02@example.com")
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Equal("user02", u.Username)
	suite.Equal("user02@example.com", u.Email)
	suite.Equal(user.Registered, u.Status)
	suite.Equal(user.UserRegistered.String(), u.Events()[0].EventName())

	if err := suite.users.Store(u); err != nil {
		suite.Fail(err.Error())
		return
	}

	u, err = suite.svc.OTPVerify("TODO", u.ID)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Equal("user02", u.Username)
	suite.Equal("user02@example.com", u.Email)
	suite.Equal(user.Activated, u.Status)
	suite.Equal(user.UserActivated.String(), u.Events()[0].EventName())
}

func (suite *identityTestSuite) TestSignInWithGoogle() {
	u, err := suite.svc.SignIn(suite.token, user.GOOGLE)
	if err != nil {
		suite.Error(err)
		suite.T().Skip()
		return
	}

	sid := user.SocialID("100043685676652067799")

	suite.Equal("mirror770109", u.Username)
	suite.Equal(user.Activated, u.Status)
	suite.Equal(sid, u.Accounts[0].SocialID)

	suite.Equal(user.UserRegistered.String(), u.Events()[0].EventName())
	suite.Equal(user.UserSocialAccountAdded.String(), u.Events()[1].EventName())
}

func (suite *identityTestSuite) TearDownSuite() {
	suite.users.Close()
}

func TestIdentityTestSuite(t *testing.T) {
	suite.Run(t, new(identityTestSuite))
}

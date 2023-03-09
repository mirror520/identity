package main

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mirror520/identity"
	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/persistent/inmem"
	"github.com/mirror520/identity/user"
)

type identityTestSuite struct {
	suite.Suite
	svc   identity.Service
	users user.Repository
	token string
}

func (suite *identityTestSuite) SetupSuite() {
	cfg, err := conf.LoadConfig("../..")
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.token = "YOUR GOOGLE JWT TOKEN" // Token 需由 Google 簽發

	users, err := inmem.NewUserRepository()
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.svc = identity.NewService(users, cfg.Providers)
	suite.users = users
}

func (suite *identityTestSuite) TestSignInWithGoogle() {
	u, err := suite.svc.SignIn(suite.token, user.GOOGLE)
	if err != nil {
		suite.Error(err)
		suite.T().Skip()
		return
	}

	suite.Equal("mirror770109", u.Username)
	suite.Equal("100043685676652067799", u.Accounts[0].SocialID)
}

func TestIdentityTestSuite(t *testing.T) {
	suite.Run(t, new(identityTestSuite))
}

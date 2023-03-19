package kv

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mirror520/identity/conf"
	"github.com/mirror520/identity/user"
)

type userRepositoryTestSuite struct {
	suite.Suite
	users user.Repository
	user  *user.User
}

func (suite *userRepositoryTestSuite) SetupSuite() {
	cfg := conf.Persistent{
		Driver: conf.BadgerDB,
		Name:   "identity",
		InMem:  true,
	}

	users, err := NewUserRepository(cfg)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	u := user.NewUser("mirror770109", "Lin, Ying-Chin", "mirror770109@gmail.com")
	u.AddSocialAccount(user.GOOGLE, "100043685676652067799")
	users.Store(u)

	suite.users = users
	suite.user = u
}

func (suite *userRepositoryTestSuite) TestFind() {
	user, err := suite.users.Find(suite.user.ID)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Equal("mirror770109", user.Username)
}

func (suite *userRepositoryTestSuite) TestFindByUsername() {
	user, err := suite.users.FindByUsername(suite.user.Username)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Equal("mirror770109", user.Username)
}

func (suite *userRepositoryTestSuite) TestFindBySocialID() {
	sid := suite.user.Accounts[0].SocialID

	user, err := suite.users.FindBySocialID(sid)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Equal("mirror770109", user.Username)
	suite.Equal(sid, user.Accounts[0].SocialID)
}

func (suite *userRepositoryTestSuite) TearDownSuite() {
	suite.users.Close()

	// os.RemoveAll("identity")
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(userRepositoryTestSuite))
}

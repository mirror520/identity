package db

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mirror520/identity/model/conf"
	"github.com/mirror520/identity/model/user"
)

type userRepositoryTestSuite struct {
	suite.Suite
	users user.Repository
	user  *user.User
}

func (suite *userRepositoryTestSuite) SetupSuite() {
	cfg, err := conf.LoadConfig("../..")
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	users, err := NewUserRepository(cfg.Persistent)
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

func (suite *userRepositoryTestSuite) TestFindBySocialID() {
	sid := "100043685676652067799"

	user, err := suite.users.FindBySocialID(sid)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Equal("mirror770109", user.Username)
	suite.Equal(sid, user.Accounts[0].SocialID)
}

func (suite *userRepositoryTestSuite) TearDownTest() {
	db := suite.users.(DBPersistent).DB()
	db.Exec("DROP TABLE social_accounts")
	db.Exec("DROP TABLE users")
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(userRepositoryTestSuite))
}

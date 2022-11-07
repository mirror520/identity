package identity

import (
	"testing"

	"github.com/jinzhu/configor"
	"github.com/stretchr/testify/suite"

	"github.com/mirror520/jinte/model"
	"github.com/mirror520/jinte/model/user"
	"github.com/mirror520/jinte/persistent/db"
)

type identityTestSuite struct {
	suite.Suite
	svc   Service
	users user.Repository
	token string
}

func (suite *identityTestSuite) SetupSuite() {
	configor.Load(&model.Config, "../../config.yaml")
	suite.token = "YOUR GOOGLE JWT TOKEN" // Token 需由 Google 簽發
}

func (suite *identityTestSuite) SetupTest() {
	users, err := db.NewUserRepository()
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.svc = NewService(users)
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
	suite.Equal(user.SocialAccountID("100043685676652067799"), u.Accounts[0].SocialID)
}

func (suite *identityTestSuite) TearDownTest() {
	db := suite.users.(db.DBPersistent).DB()
	db.Exec("DROP TABLE workspace_members")
	db.Exec("DROP TABLE workspaces")
	db.Exec("DROP TABLE social_accounts")
	db.Exec("DROP TABLE users")
}

func TestIdentityTestSuite(t *testing.T) {
	suite.Run(t, new(identityTestSuite))
}

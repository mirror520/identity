package db

import (
	"testing"

	"github.com/jinzhu/configor"
	"github.com/stretchr/testify/suite"

	"github.com/mirror520/jinte/model"
	"github.com/mirror520/jinte/model/user"
)

type userRepositoryTestSuite struct {
	suite.Suite
	users user.Repository
	user  *user.User
}

func (suite *userRepositoryTestSuite) SetupSuite() {
	configor.Load(&model.Config, "../../config.yaml")
}

func (suite *userRepositoryTestSuite) SetupTest() {
	users, err := NewUserRepository()
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	u := user.NewUser("mirror770109", "Lin, Ying-Chin", "mirror770109@gmail.com")
	u.AddSocialAccount(user.GOOGLE, "100043685676652067799")
	users.Store(u)

	w := u.DefaultWorkspace()
	w.Name = "Mirror's Workspace"
	users.StoreWorkspace(w)

	w1 := u.BuildWorkspace("First Workspace")
	w2 := u.BuildWorkspace("Second Workspace")
	w3 := u.BuildWorkspace("Third Workspace")

	users.StoreWorkspace(w1)
	users.StoreWorkspace(w2)
	users.StoreWorkspace(w3)

	suite.users = users
	suite.user = u
}

func (suite *userRepositoryTestSuite) TestFindBySocialID() {
	sid := user.SocialAccountID("100043685676652067799")

	user, err := suite.users.FindBySocialID(sid)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Equal("mirror770109", user.Username)
	suite.Equal(sid, user.Accounts[0].SocialID)
}

func (suite *userRepositoryTestSuite) TestFindWorkspaces() {
	workspaces, err := suite.users.FindWorkspaces(suite.user.ID)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.Len(workspaces, 4)
}

func (suite *userRepositoryTestSuite) TearDownTest() {
	db := suite.users.(DBPersistent).DB()
	db.Exec("DROP TABLE workspace_members")
	db.Exec("DROP TABLE workspaces")
	db.Exec("DROP TABLE social_accounts")
	db.Exec("DROP TABLE users")
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(userRepositoryTestSuite))
}

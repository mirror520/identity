package policy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type policyTestSuite struct {
	suite.Suite
	policy Policy
}

func (suite *policyTestSuite) SetupSuite() {
	p, err := NewRegoPolicy(context.TODO(), ".")
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.policy = p
}

func (suite *policyTestSuite) TestEvalListUsersWithAdminRole() {
	input := map[string]any{
		"domain": "identity::users",
		"action": "list",
		"claims": map[string]any{
			"roles": []string{"admin"},
		},
	}

	accepted, err := suite.policy.Eval(context.TODO(), input)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.True(accepted)
}

func (suite *policyTestSuite) TestEvalNotListUsersWithUserRole() {
	input := map[string]any{
		"domain": "identity::users",
		"action": "list",
		"claims": map[string]any{
			"roles": []string{"user"},
		},
	}

	accepted, err := suite.policy.Eval(context.TODO(), input)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.False(accepted)
}

func (suite *policyTestSuite) TestEvalUpdateUsersWithUserRoleAndOwner() {
	input := map[string]any{
		"domain":    "identity::users",
		"action":    "update",
		"object":    "mirror520",
		"who_flags": 0b0001,
		"claims": map[string]any{
			"sub":   "mirror520",
			"roles": []string{"user"},
		},
	}

	accepted, err := suite.policy.Eval(context.TODO(), input)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.True(accepted)
}

func (suite *policyTestSuite) TestEvalUpdateUsersWithUserRoleAndNotOwner() {
	input := map[string]any{
		"domain":    "identity::users",
		"action":    "update",
		"object":    "mirror",
		"who_flags": 0b0001,
		"claims": map[string]any{
			"sub":   "mirror520",
			"roles": []string{"user"},
		},
	}

	accepted, err := suite.policy.Eval(context.TODO(), input)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.False(accepted)
}

func (suite *policyTestSuite) TestEvalUpdateUsersWithAdminRoleAndAdmin() {
	input := map[string]any{
		"domain":    "identity::users",
		"action":    "update",
		"who_flags": 0b1000,
		"claims": map[string]any{
			"sub":   "mirror520",
			"roles": []string{"admin"},
		},
	}

	accepted, err := suite.policy.Eval(context.TODO(), input)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.True(accepted)
}

func (suite *policyTestSuite) TestEvalUpdateUsersWithUserRoleAndNotAdmin() {
	input := map[string]any{
		"domain":    "identity::users",
		"action":    "update",
		"who_flags": 0b1000,
		"claims": map[string]any{
			"sub":   "mirror520",
			"roles": []string{"user"},
		},
	}

	accepted, err := suite.policy.Eval(context.TODO(), input)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.False(accepted)
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(policyTestSuite))
}

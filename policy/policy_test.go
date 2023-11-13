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

func (suite *policyTestSuite) TestEvalWithAdmin() {
	input := map[string]any{
		"domain": "identity::users",
		"action": "list",
		"roles": []string{
			"admin",
		},
	}

	accepted, err := suite.policy.Eval(context.TODO(), input)
	if err != nil {
		suite.Fail(err.Error())
		return
	}

	suite.True(accepted)
}

func TestPolicyTestSuite(t *testing.T) {
	suite.Run(t, new(policyTestSuite))
}

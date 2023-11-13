package policy

import (
	"context"
	"os"

	_ "embed"

	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/storage/inmem"
)

type Policy interface {
	Eval(ctx context.Context, input any) (bool, error)
}

//go:embed rbac.rego
var module string

type regoPolicy struct {
	query *rego.PreparedEvalQuery
	store storage.Store
}

func NewRegoPolicy(ctx context.Context, path string) (Policy, error) {
	f, err := os.Open(path + "/data.json")
	if err != nil {
		return nil, err
	}

	store := inmem.NewFromReader(f)

	query, err := rego.New(
		rego.Module("rbac.rego", module),
		rego.Query("data.app.rbac.allow"),
		rego.Store(store),
	).PrepareForEval(ctx)

	if err != nil {
		return nil, err
	}

	return &regoPolicy{
		&query,
		store,
	}, nil
}

func (policy *regoPolicy) Eval(ctx context.Context, input any) (bool, error) {
	results, err := policy.query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, err
	}

	return results.Allowed(), nil
}

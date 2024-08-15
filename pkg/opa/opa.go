package opa

import (
	"context"

	"github.com/open-policy-agent/opa/rego"
)

type OPA struct {
	query rego.PreparedEvalQuery
}

func NewOPA(policyFile string) (*OPA, error) {
	ctx := context.Background()

	query, err := rego.New(
		rego.Query("data.rbac.allow"),
		rego.Load([]string{policyFile}, nil),
	).PrepareForEval(ctx)

	if err != nil {
		return nil, err
	}

	return &OPA{query: query}, nil
}

func (o *OPA) Evaluate(ctx context.Context, input interface{}) (bool, error) {
	results, err := o.query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return false, err
	}

	if len(results) == 0 {
		return false, nil
	}

	allowed, ok := results[0].Expressions[0].Value.(bool)
	if !ok {
		return false, nil
	}

	return allowed, nil
}

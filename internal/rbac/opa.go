package rbac

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"

	"github.com/open-policy-agent/opa/rego"
)

//go:embed rbac.rego
var policyContent string

var opaQuery rego.PreparedEvalQuery

func InitOPA() error {
	ctx := context.Background()

	// 直接使用嵌入的策略内容
	query, err := rego.New(
		rego.Query("data.rbac.allow"),
		rego.Module("rbac.rego", policyContent),
	).PrepareForEval(ctx)

	if err != nil {
		return fmt.Errorf("failed to prepare OPA query: %w", err)
	}

	opaQuery = query
	return nil
}

func evaluateOPAPolicy(input *PermissionInput) (bool, error) {
	ctx := context.Background()

	// 将输入转换为 map[string]interface{}
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return false, fmt.Errorf("failed to marshal input: %w", err)
	}

	log.Printf("inputJSON: %+v\n", string(inputJSON))

	var inputMap map[string]interface{}
	if err := json.Unmarshal(inputJSON, &inputMap); err != nil {
		return false, fmt.Errorf("failed to unmarshal input: %w", err)
	}

	// 评估 OPA 策略
	results, err := opaQuery.Eval(ctx, rego.EvalInput(inputMap))
	if err != nil {
		return false, fmt.Errorf("failed to evaluate OPA policy: %w", err)
	}

	if len(results) == 0 || len(results[0].Expressions) == 0 {
		return false, nil
	}

	allowed, ok := results[0].Expressions[0].Value.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected result type from OPA evaluation")
	}

	return allowed, nil
}

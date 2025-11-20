package loader

import (
	"fmt"
	"os"

	"github.com/palantir/policy-bot/policy"
	"github.com/palantir/policy-bot/policy/common"
	"gopkg.in/yaml.v3"
)

// LoadPolicyEvaluator loads and parses a policy configuration file
func LoadPolicyEvaluator(fileName string) (common.Evaluator, error) {
	policyFile, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load file %s: %w", fileName, err)
	}
	var policyConfig policy.Config
	if err := yaml.UnmarshalStrict(policyFile, &policyConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %s: %w", fileName, err)
	}
	return policy.ParsePolicy(&policyConfig, nil)
}

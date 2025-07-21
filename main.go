package main

import (
	"context"
	"log"
	"os"

	"github.com/palantir/policy-bot/policy"
	"gopkg.in/yaml.v2"
)

func main() {
	file, err := os.ReadFile(".policy.yml")
	if err != nil {
		log.Fatal(err)
	}
	var policyConfig policy.Config
	if err := yaml.UnmarshalStrict(file, &policyConfig); err != nil {
		log.Fatal(err)
	}

	evaluator, err := policy.ParsePolicy(&policyConfig, nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.TODO()

	pullContext := GitHubContext{}
	evaluator.Evaluate(ctx, &pullContext)
}

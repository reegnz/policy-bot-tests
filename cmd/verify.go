package cmd

import (
	"fmt"
	"os"

	"github.com/reegnz/policy-bot-tests/internal/loader"
	"github.com/reegnz/policy-bot-tests/internal/runner"
	"github.com/spf13/cobra"
)

const (
	defaultTestPath   = ".policy-tests"
	defaultPolicyFile = ".policy.yml"
	defaultOutput     = "pretty"
)

var (
	verifyVerbose      int
	verifyFilter       string
	verifyOutputFormat string
	verifyPolicyFile   string
)

// NewVerifyCommand creates the "verify" subcommand
func NewVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [paths...]",
		Short: "Verifies test cases against a policy file",
		Long:  "Verifies test cases against a policy file by loading test cases and evaluating them.",
		RunE:  runVerify,
	}

	cmd.Flags().CountVarP(&verifyVerbose, "verbose", "v", "increase verbosity (can be repeated: -v, -vv, -vvv)")
	cmd.Flags().StringVarP(&verifyFilter, "filter", "f", "", "filter test cases by name using regex")
	cmd.Flags().StringVarP(&verifyOutputFormat, "output", "o", defaultOutput, "output format (pretty, efm)")
	cmd.Flags().StringVarP(&verifyPolicyFile, "policy", "p", defaultPolicyFile, "path to the policy file")

	return cmd
}

func runVerify(cmd *cobra.Command, args []string) error {
	evaluator, err := loader.LoadPolicyEvaluator(verifyPolicyFile)
	if err != nil {
		return fmt.Errorf("failed to load evaluator: %w", err)
	}

	if len(args) == 0 {
		args = []string{defaultTestPath}
	}

	tests, err := loader.LoadTestFiles(args)
	if err != nil {
		return fmt.Errorf("failed to load tests: %w", err)
	}
	if passed := runner.RunTests(evaluator, tests, verifyVerbose, verifyFilter, verifyOutputFormat); !passed {
		os.Exit(1)
	}
	return nil
}

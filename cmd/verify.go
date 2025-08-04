package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/reegnz/policy-bot-tests/internal/loader"
	"github.com/reegnz/policy-bot-tests/internal/runner"
	"github.com/spf13/cobra"
)

var (
	verifyVerbose      int
	verifyFilter       string
	verifyOutputFormat string
)

// NewVerifyCommand creates the "verify" subcommand
func NewVerifyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [directory]",
		Short: "Verifies test cases against a policy file",
		Long:  "Verifies test cases against a policy file by loading test cases and evaluating them.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runVerify,
	}

	cmd.Flags().CountVarP(&verifyVerbose, "verbose", "v", "increase verbosity (can be repeated: -v, -vv, -vvv)")
	cmd.Flags().StringVarP(&verifyFilter, "filter", "f", "", "filter test cases by name using regex")
	cmd.Flags().StringVarP(&verifyOutputFormat, "output", "o", "pretty", "output format (pretty, efm)")

	return cmd
}

func runVerify(cmd *cobra.Command, args []string) error {
	directory := "."
	if len(args) > 0 {
		directory = args[0]
	}

	policyFile := filepath.Join(directory, ".policy.yml")
	testFile := filepath.Join(directory, ".policy-tests.yml")

	evaluator, err := loader.LoadPolicyEvaluator(policyFile)
	if err != nil {
		return fmt.Errorf("failed to load evaluator: %w", err)
	}
	tests, err := loader.LoadTestFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to load tests: %w", err)
	}
	if passed := runner.RunTests(evaluator, tests, verifyVerbose, verifyFilter, verifyOutputFormat); !passed {
		os.Exit(1)
	}
	return nil
}

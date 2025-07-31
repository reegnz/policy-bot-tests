package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/reegnz/policy-bot-tests/internal/loader"
	"github.com/reegnz/policy-bot-tests/internal/runner"
	"github.com/spf13/cobra"
)

// Version information - set by GoReleaser
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

var (
	verbose int
	filter  string
)

// NewRootCommand creates the root command for the CLI
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "policy-bot-tests [directory]",
		Short: "Run tests for policy-bot configurations",
		Long:  "A testing tool for policy-bot configurations that loads test cases and evaluates them against a policy file.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runMain,
	}

	// Add flags
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "increase verbosity (can be repeated: -v, -vv, -vvv)")
	rootCmd.PersistentFlags().StringVarP(&filter, "filter", "f", "", "filter test cases by name using regex")
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, date: %s)", Version, Commit, Date)

	return rootCmd
}

// Execute runs the root command
func Execute() {
	log.SetFlags(0)

	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func runMain(cmd *cobra.Command, args []string) error {
	// Default to current directory if no argument provided
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
	if passed := runner.RunTests(evaluator, tests, verbose, filter); !passed {
		os.Exit(1)
	}
	return nil
}

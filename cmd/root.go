package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// Version information - set by GoReleaser
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// NewRootCommand creates the root command for the CLI
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "policy-bot-tests",
		Short: "A testing tool for policy-bot configurations",
		Long:  "A testing tool for policy-bot configurations that loads test cases and evaluates them against a policy file.",
	}

	rootCmd.AddCommand(NewVerifyCommand())
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, date: %s)", Version, Commit, Date)

	return rootCmd
}

// Execute runs the root command
func Execute() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

package main

import "github.com/reegnz/policy-bot-tests/cmd"

// Version information - set by GoReleaser
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Set version information for the cmd package
	cmd.Version = version
	cmd.Commit = commit
	cmd.Date = date

	// Execute the root command
	cmd.Execute()
}

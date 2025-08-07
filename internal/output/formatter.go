package output

import (
	"log"
	"slices"
	"strings"

	"github.com/palantir/policy-bot/policy/common"
	"github.com/reegnz/policy-bot-tests/internal/models"
)

// PrintTestContext prints the test context information with proper formatting
func PrintTestContext(tc models.TestContext, indent string) {
	log.Printf("%s- Author: %s", indent, tc.Author)
	if len(tc.FilesChanged) > 0 {
		log.Printf("%s- Changed Files:", indent)
		for _, file := range tc.FilesChanged {
			log.Printf("%s  - %s", indent, file)
		}
	}
	if len(tc.Reviews) > 0 {
		log.Printf("%s- Reviews:", indent)
		for _, r := range tc.Reviews {
			log.Printf("%s  - %s (%s)", indent, r.Author, r.State)
		}
	}
	if len(tc.Labels) > 0 {
		log.Printf("%s- Labels:", indent)
		for _, label := range tc.Labels {
			log.Printf("%s  - %s", indent, label)
		}
	}
	if len(tc.Statuses) > 0 {
		log.Printf("%s- Statuses:", indent)
		for k, v := range tc.Statuses {
			log.Printf("%s  - %s: %s", indent, k, v)
		}
	}
	if len(tc.WorkflowRuns) > 0 {
		log.Printf("%s- Workflows:", indent)
		for k, v := range tc.WorkflowRuns {
			log.Printf("%s  - %s: %s", indent, k, strings.Join(v, ", "))
		}
	}
}

// PrintResultTree prints the policy evaluation result tree with proper formatting
func PrintResultTree(result *common.Result, indent string, showSkipped bool) {
	statusIcon := "⚪"
	switch result.Status {
	case common.StatusApproved:
		statusIcon = "✅"
	case common.StatusSkipped:
		statusIcon = "💤"
	case common.StatusPending:
		statusIcon = "🟡"
	case common.StatusDisapproved:
		statusIcon = "❌"
	}

	log.Printf("%s- %s %s: %s\n", indent, statusIcon, result.Name, result.StatusDescription)

	sortedChildren := sortResults(result.Children)

	for _, child := range sortedChildren {
		if child.Status == common.StatusSkipped && !showSkipped {
			continue
		}
		PrintResultTree(child, indent+"  ", showSkipped)
	}
}

// sortResults sorts a slice of results based on their status.
// The sort order is Disapproved > Approved > Pending > Skipped.
func sortResults(results []*common.Result) []*common.Result {
	sorted := slices.Clone(results)
	slices.SortFunc(sorted, func(a, b *common.Result) int {
		// Sorting by the integer value of the status enum achieves the desired order
		return int(b.Status) - int(a.Status)
	})
	return sorted
}

// PrintAssertionResult prints the assertion result with proper formatting
func PrintAssertionResult(assertionResult models.AssertionResult, verbosity int, indent string) {
	// Print evaluation status
	log.Printf("%s- Evaluation status:\n", indent)
	log.Printf("%s  - Expected: %v\n", indent, assertionResult.ExpectedStatus)
	log.Printf("%s  - Actual: %v\n", indent, assertionResult.ActualStatus)

	// Show failing assertions or all if verbosity >= 2
	if verbosity >= 1 || !assertionResult.Success() {
		if verbosity >= 3 || assertionResult.HasMissingApproved() {
			printRuleSection("approved", assertionResult.MissingApproved(), indent)
		}
		if verbosity >= 3 || assertionResult.HasMissingPending() {
			printRuleSection("pending", assertionResult.MissingPending(), indent)
		}
		if verbosity >= 3 || assertionResult.HasMissingSkipped() {
			printRuleSection("skipped", assertionResult.MissingSkipped(), indent)
		}
	}
}

// printRuleSection prints a section of rules with expected and actual lists
func printRuleSection(sectionName string, rules []string, indent string) {
	log.Printf("%s- Missing %s Rules:\n", indent, sectionName)
	for _, rule := range rules {
		log.Printf("%s  - %s\n", indent, rule)
	}
}

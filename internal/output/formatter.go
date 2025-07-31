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
	log.Printf("%s- Author: %s", indent, tc.PR.Author)
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
	if len(tc.Tags) > 0 {
		log.Printf("%s- Labels:", indent)
		for _, tag := range tc.Tags {
			log.Printf("%s  - %s", indent, tag)
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

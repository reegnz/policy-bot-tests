package runner

import (
	"context"
	"log"
	"maps"
	"regexp"

	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
	"github.com/reegnz/policy-bot-tests/internal/models"
	"github.com/reegnz/policy-bot-tests/internal/output"
)

// RunTests executes test cases against a policy evaluator
func RunTests(evaluator common.Evaluator, tests *models.TestFile, verbosity int, filter string) (passed bool) {
	var filterRegex *regexp.Regexp
	var err error
	if filter != "" {
		filterRegex, err = regexp.Compile(filter)
		if err != nil {
			log.Fatalf("Invalid filter regex: %v", err)
		}
	}

	var filteredCases []models.TestCase
	if filterRegex != nil {
		for _, tc := range tests.TestCases {
			if filterRegex.MatchString(tc.Name) {
				filteredCases = append(filteredCases, tc)
			}
		}
	} else {
		filteredCases = tests.TestCases
	}

	if len(filteredCases) == 0 {
		log.Printf("No test cases matched the filter: %s", filter)
		return true
	}

	log.Printf("Running %d of %d total test case(s)\n", len(filteredCases), len(tests.TestCases))
	passedCount := 0
	for _, tc := range filteredCases {
		if tc.LineNumber > 0 {
			log.Printf("--- Running Test: .policy-tests.yml:%d: %s ---\n", tc.LineNumber, tc.Name)
		} else {
			log.Printf("--- Running Test: %s ---\n", tc.Name)
		}

		mergedContext := MergeContexts(tests.DefaultContext, tc.Context)
		pullContext := NewPullContext(mergedContext)
		result := evaluator.Evaluate(context.Background(), pullContext)

		assertionResult := CheckAssertions(tc.Assert, &result)
		pass := assertionResult.Success()
		if !pass || verbosity >= 1 {
			if verbosity >= 3 {
				log.Println("  - Test Context:")
				output.PrintTestContext(mergedContext, "    ")
			}
			if !pass || verbosity >= 1 {
				output.PrintAssertionResult(assertionResult, verbosity)
			}
			log.Println("  - Policy Evaluation Tree:")
			output.PrintResultTree(&result, "    ", verbosity >= 3)
		}

		if pass {
			passedCount++
			log.Println("\033[32mPASS\033[0m")
		} else {
			log.Println("\033[31mFAIL\033[0m")
		}
	}
	log.Printf("\n--- Summary ---\n%d / %d tests passed.\n", passedCount, len(filteredCases))
	passed = passedCount == len(filteredCases)
	return
}

// CheckAssertions validates test assertions against evaluation results
func CheckAssertions(assert models.TestAssertion, result *common.Result) models.AssertionResult {
	// Check approved and pending rules
	approved, pending, skipped := collectRuleStatuses(result)
	return models.NewAssertionResult(assert, result.Status.String(), approved, pending, skipped)
}

// collectRuleStatuses recursively collects rule statuses from evaluation results
func collectRuleStatuses(result *common.Result) (approved, pending, skipped []string) {
	// If a result has children, it is a logical grouping (e.g. and, or).
	// Recurse into the children to find the individual rule results.
	if len(result.Children) > 0 {
		for _, child := range result.Children {
			a, p, s := collectRuleStatuses(child)
			approved = append(approved, a...)
			pending = append(pending, p...)
			skipped = append(skipped, s...)
		}
		return
	}
	// If a result has no children, it is a leaf node representing a rule.
	switch result.Status {
	case common.StatusApproved:
		return []string{result.Name}, nil, nil
	case common.StatusPending:
		return nil, []string{result.Name}, nil
	case common.StatusSkipped:
		return nil, nil, []string{result.Name}
	}
	return nil, nil, nil
}

// MergeContexts merges test contexts with override precedence
func MergeContexts(base, override models.TestContext) models.TestContext {
	merged := base

	if len(override.FilesChanged) > 0 {
		merged.FilesChanged = override.FilesChanged
	}
	if override.Owner != "" {
		merged.Owner = override.Owner
	}
	if override.Repo != "" {
		merged.Repo = override.Repo
	}
	if override.PR.Author != "" {
		merged.PR.Author = override.PR.Author
	}
	if override.PR.BaseRefName != "" {
		merged.PR.BaseRefName = override.PR.BaseRefName
	}
	if override.PR.HeadRefName != "" {
		merged.PR.HeadRefName = override.PR.HeadRefName
	}

	if len(override.Reviews) > 0 {
		merged.Reviews = override.Reviews
	}
	if len(override.Statuses) > 0 {
		merged.Statuses = override.Statuses
	}
	if len(override.WorkflowRuns) > 0 {
		merged.WorkflowRuns = override.WorkflowRuns
	}
	if len(override.Tags) > 0 {
		merged.Tags = override.Tags
	}
	if len(override.TeamMembers) > 0 {
		maps.Copy(merged.TeamMembers, override.TeamMembers)
	}
	if len(override.OrgMembers) > 0 {
		maps.Copy(merged.OrgMembers, override.OrgMembers)
	}

	return merged
}

// NewPullContext creates a pull.Context from test context data
func NewPullContext(tc models.TestContext) pull.Context {
	reviews := []*pull.Review{}
	for _, r := range tc.Reviews {
		reviews = append(reviews, &pull.Review{
			Author: r.Author,
			State:  pull.ReviewState(r.State),
		})
	}

	files := []*pull.File{}
	for _, f := range tc.FilesChanged {
		files = append(files, &pull.File{Filename: f})
	}

	collaborators := []*pull.Collaborator{}
	seenCollaborators := map[string]bool{}
	for _, members := range tc.TeamMembers {
		for _, member := range members {
			if !seenCollaborators[member] {
				collaborators = append(collaborators, &pull.Collaborator{
					Name: member,
					Permissions: []pull.CollaboratorPermission{
						{
							Permission: pull.PermissionWrite,
							ViaRepo:    true,
						},
					},
				})
				seenCollaborators[member] = true
			}
		}
	}

	return models.NewGitHubContext(tc, reviews, files, collaborators)
}

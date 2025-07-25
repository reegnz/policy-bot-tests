package main

import (
	"context"
	"fmt"
	"log"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/palantir/policy-bot/policy"
	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
	"gopkg.in/yaml.v2"
)

func init() {
	log.SetFlags(0)
}

func main() {
	evaluator, err := loadPolicyEvaluator(".policy.yml")
	if err != nil {
		log.Fatalf("Failed to load evaluator: %v", err)
	}
	tests, err := loadTestFile(".policy-tests.yml")
	if err != nil {
		log.Fatalf("Failed to load tests: %v", err)
	}
	if passed := runTests(evaluator, tests); !passed {
		os.Exit(1)
	}
}

func loadPolicyEvaluator(fileName string) (common.Evaluator, error) {
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

func loadTestFile(fileName string) (*TestFile, error) {
	testFile, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load file %s: %w", fileName, err)
	}
	var tests TestFile
	if err := yaml.Unmarshal(testFile, &tests); err != nil {
		log.Fatalf("Failed to unmarshal .policy-tests.yml: %v", err)
	}
	return &tests, nil
}

func runTests(evaluator common.Evaluator, tests *TestFile) (passed bool) {
	log.Printf("Found %d test case(s)\n", len(tests.TestCases))
	passedCount := 0
	for _, tc := range tests.TestCases {
		log.Printf("--- Running Test: %s ---\n", tc.Name)

		mergedContext := mergeContexts(tests.DefaultContext, tc.Context)
		pullContext := newPullContext(mergedContext)
		result := evaluator.Evaluate(context.Background(), pullContext)

		pass := checkAssertions(tc.Assert, &result)
		log.Println("  - Evaluation Tree:")
		printResultTree(&result, "    ")
		if pass {
			passedCount++
			log.Println("\033[32mPASS\033[0m")
		} else {
			log.Println("\033[31mFAIL\033[0m")
		}
	}
	log.Printf("\n--- Summary ---\n%d / %d tests passed.\n", passedCount, len(tests.TestCases))
	passed = passedCount == len(tests.TestCases)
	return
}

func checkAssertions(assert TestAssertion, result *common.Result) (isSuccess bool) {
	// Check overall status
	if result.Status.String() != assert.EvaluationStatus {
		log.Printf("  - Expected evaluation status '%s', but got '%s'\n", assert.EvaluationStatus, result.Status)
		return
	}
	// Check approved and pending rules
	approved, pending, _ := collectRuleStatuses(result)
	slices.Sort(approved)
	slices.Sort(assert.ApprovedRules)
	if !reflect.DeepEqual(approved, assert.ApprovedRules) {
		log.Printf("  - Approved Rules do not match:\n")
		log.Printf("      Expected: %v\n", strings.Join(assert.ApprovedRules, ", "))
		log.Printf("      Actual:   %v\n", strings.Join(approved, ", "))
		return
	}
	slices.Sort(pending)
	slices.Sort(assert.PendingRules)
	if !reflect.DeepEqual(pending, assert.PendingRules) {
		log.Printf("  - Pending Rules do not match:\n")
		log.Printf("      Expected: %v\n", strings.Join(assert.PendingRules, ", "))
		log.Printf("      Actual:   %v\n", strings.Join(pending, ", "))
		return
	}
	return true
}

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

func printResultTree(result *common.Result, indent string) {
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
	for _, child := range result.Children {
		printResultTree(child, indent+"  ")
	}
}

func mergeContexts(base, override TestContext) TestContext {
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

func newPullContext(tc TestContext) pull.Context {
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

	return &GitHubContext{
		owner:         tc.Owner,
		repo:          tc.Repo,
		statuses:      tc.Statuses,
		workflowRuns:  tc.WorkflowRuns,
		collaborators: collaborators,
		pr: PullRequest{
			author:      tc.PR.Author,
			baseRefName: tc.PR.BaseRefName,
			headRefName: tc.PR.HeadRefName,
		},
		reviews: reviews,
		files:   files,
		labels:  tc.Tags,
		GitHubMembershipContext: GitHubMembershipContext{
			teamMembers: tc.TeamMembers,
			orgMembers:  tc.OrgMembers,
		},
		evalTimestamp: time.Now(),
	}
}

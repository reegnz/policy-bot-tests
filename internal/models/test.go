package models

import "slices"

// TestFile matches the root of the .policy-tests.yml file
type TestFile struct {
	DefaultContext TestContext `yaml:"defaultContext"`
	TestCases      []TestCase  `yaml:"testCases"`
}

// TestCase represents a single test case from the YAML file
type TestCase struct {
	Name       string        `yaml:"name"`
	Context    TestContext   `yaml:"context"`
	Assert     TestAssertion `yaml:"assert"`
	LineNumber int           `yaml:"-"`
}

// TestContext is a simplified version of GitHubContext for easy YAML parsing
type TestContext struct {
	FilesChanged []string            `yaml:"filesChanged"`
	Owner        string              `yaml:"owner"`
	Repo         string              `yaml:"repo"`
	PR           TestPullRequest     `yaml:"pr"`
	Reviews      []TestReview        `yaml:"reviews"`
	Statuses     map[string]string   `yaml:"statuses"`
	WorkflowRuns map[string][]string `yaml:"workflowRuns"`
	Tags         []string            `yaml:"tags"`
	TeamMembers  map[string][]string `yaml:"teamMembers"`
	OrgMembers   map[string][]string `yaml:"orgMembers"`
}

// TestPullRequest is a simplified version of a PR for YAML parsing
type TestPullRequest struct {
	Author      string `yaml:"author"`
	BaseRefName string `yaml:"baseRefName"`
	HeadRefName string `yaml:"headRefName"`
}

// TestReview is a simplified version of a review for YAML parsing
type TestReview struct {
	Author string `yaml:"author"`
	State  string `yaml:"state"`
}

// TestAssertion defines the expected outcomes of a test case
type TestAssertion struct {
	EvaluationStatus string   `yaml:"evaluationStatus"`
	MustBeApproved   []string `yaml:"mustBeApproved"`
	MustBePending    []string `yaml:"mustBePending"`
	MustBeSkipped    []string `yaml:"mustBeSkipped"`
}

// AssertionResult holds the results of test assertions
type AssertionResult struct {
	ActualStatus     string
	ExpectedStatus   string
	ExpectedApproved []string
	ActualApproved   []string
	ExpectedPending  []string
	ActualPending    []string
	ExpectedSkipped  []string
	ActualSkipped    []string
}

// NewAssertionResult creates an AssertionResult by comparing expected assertions with actual rule statuses
func NewAssertionResult(assert TestAssertion, actualStatus string, approved, pending, skipped []string) AssertionResult {
	return AssertionResult{
		ActualStatus:     actualStatus,
		ExpectedStatus:   assert.EvaluationStatus,
		ExpectedApproved: assert.MustBeApproved,
		ActualApproved:   matchingItems(assert.MustBeApproved, approved),
		ExpectedPending:  assert.MustBePending,
		ActualPending:    matchingItems(assert.MustBePending, pending),
		ExpectedSkipped:  assert.MustBeSkipped,
		ActualSkipped:    matchingItems(assert.MustBeSkipped, skipped),
	}
}

// missingItems returns items that are present in expected but not in actual
func missingItems(expected, actual []string) []string {
	var missing []string
	for _, expectedItem := range expected {
		if !slices.Contains(actual, expectedItem) {
			missing = append(missing, expectedItem)
		}
	}
	return missing
}

// matchingItems returns items from expected that are also present in actual
func matchingItems(expected, actual []string) []string {
	var matching []string
	for _, expectedItem := range expected {
		if slices.Contains(actual, expectedItem) {
			matching = append(matching, expectedItem)
		}
	}
	return matching
}

// MissingApproved returns the expected approved rules that are not in the actual approved rules
func (ar AssertionResult) MissingApproved() []string {
	return missingItems(ar.ExpectedApproved, ar.ActualApproved)
}

// MissingPending returns the expected pending rules that are not in the actual pending rules
func (ar AssertionResult) MissingPending() []string {
	return missingItems(ar.ExpectedPending, ar.ActualPending)
}

// MissingSkipped returns the expected skipped rules that are not in the actual skipped rules
func (ar AssertionResult) MissingSkipped() []string {
	return missingItems(ar.ExpectedSkipped, ar.ActualSkipped)
}

// Success returns true if all assertions passed
func (ar AssertionResult) Success() bool {
	return ar.MatchesStatus() &&
		!ar.HasMissingApproved() &&
		!ar.HasMissingPending() &&
		!ar.HasMissingSkipped()
}

// MatchesStatus returns true if the evaluation status matches expected
func (ar AssertionResult) MatchesStatus() bool {
	return ar.ExpectedStatus == ar.ActualStatus
}

// HasMissingApproved returns true if any expected approved rules are missing
func (ar AssertionResult) HasMissingApproved() bool {
	return len(ar.MissingApproved()) > 0
}

// HasMissingPending returns true if any expected pending rules are missing
func (ar AssertionResult) HasMissingPending() bool {
	return len(ar.MissingPending()) > 0
}

// HasMissingSkipped returns true if any expected skipped rules are missing
func (ar AssertionResult) HasMissingSkipped() bool {
	return len(ar.MissingSkipped()) > 0
}

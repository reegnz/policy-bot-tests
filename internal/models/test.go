package models

// TestFile matches the root of the .policy-tests.yml file
type TestFile struct {
	DefaultContext TestContext `yaml:"defaultContext"`
	TestCases      []TestCase  `yaml:"testCases"`
}

// TestCase represents a single test case from the YAML file
type TestCase struct {
	Name    string        `yaml:"name"`
	Context TestContext   `yaml:"context"`
	Assert  TestAssertion `yaml:"assert"`
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
	ApprovedRules    []string `yaml:"approvedRules"`
	PendingRules     []string `yaml:"pendingRules"`
}

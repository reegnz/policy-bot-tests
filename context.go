package main

import (
	"slices"
	"strings"
	"time"

	"github.com/palantir/policy-bot/pull"
)

var (
	_ pull.Context           = &pull.GitHubContext{}
	_ pull.MembershipContext = &GitHubMembershipContext{}
)

type GitHubMembershipContext struct {
	orgMembers  map[string][]string
	teamMembers map[string][]string
}

func (mc *GitHubMembershipContext) IsTeamMember(team, user string) (bool, error) {
	return slices.Contains(mc.teamMembers[team], user), nil
}

func (mc *GitHubMembershipContext) IsOrgMember(org, user string) (bool, error) {
	return slices.Contains(mc.orgMembers[org], user), nil
}

func (mc *GitHubMembershipContext) TeamMembers(team string) ([]string, error) {
	return mc.teamMembers[team], nil
}

func (mc *GitHubMembershipContext) OrganizationMembers(org string) ([]string, error) {
	return mc.orgMembers[org], nil
}

type GitHubContext struct {
	GitHubMembershipContext

	evalTimestamp time.Time
	owner         string
	repo          string
	number        int
	pr            PullRequest

	files         []*pull.File
	commits       []*pull.Commit
	comments      []*pull.Comment
	reviews       []*pull.Review
	reviewers     []*pull.Reviewer
	collaborators []*pull.Collaborator
	labels        []string

	pushedAt     map[string]time.Time
	statuses     map[string]string
	workflowRuns map[string][]string
}

type PullRequest struct {
	author    string
	title     string
	createdAt time.Time
	state     string

	isDraft bool

	headSHA     string
	headRefName string

	baseRefName string

	body pull.Body
}

func (ghc *GitHubContext) EvaluationTimestamp() time.Time {
	return ghc.evalTimestamp
}

func (ghc *GitHubContext) RepositoryOwner() string {
	return ghc.owner
}

func (ghc *GitHubContext) RepositoryName() string {
	return ghc.repo
}

func (ghc *GitHubContext) Number() int {
	return ghc.number
}

func (ghc *GitHubContext) Title() string {
	return ghc.pr.title
}

func (ghc *GitHubContext) Body() (*pull.Body, error) {
	return &ghc.pr.body, nil
}

func (ghc *GitHubContext) Author() string {
	return ghc.pr.author
}

func (ghc *GitHubContext) CreatedAt() time.Time {
	return ghc.pr.createdAt
}

func (ghc *GitHubContext) IsOpen() bool {
	return strings.ToLower(ghc.pr.state) == "open"
}

func (ghc *GitHubContext) IsClosed() bool {
	return strings.ToLower(ghc.pr.state) == "closed"
}

func (ghc *GitHubContext) HeadSHA() string {
	return ghc.pr.headSHA
}

func (ghc *GitHubContext) IsDraft() bool {
	return ghc.pr.isDraft
}

func (ghc *GitHubContext) Branches() (base string, head string) {
	base = ghc.pr.baseRefName
	head = ghc.pr.headRefName
	return
}

func (ghc *GitHubContext) ChangedFiles() ([]*pull.File, error) {
	return ghc.files, nil
}

func (ghc *GitHubContext) Commits() ([]*pull.Commit, error) {
	return ghc.commits, nil
}

func (ghc *GitHubContext) Comments() ([]*pull.Comment, error) {
	return ghc.comments, nil
}

func (ghc *GitHubContext) Reviews() ([]*pull.Review, error) {
	return ghc.reviews, nil
}

func (ghc *GitHubContext) RepositoryCollaborators() ([]*pull.Collaborator, error) {
	return ghc.collaborators, nil
}

func (ghc *GitHubContext) CollaboratorPermission(user string) (pull.Permission, error) {
	// Check if the user is a member of any team defined in the test case.
	for _, members := range ghc.teamMembers {
		if slices.Contains(members, user) {
			return pull.PermissionWrite, nil
		}
	}
	// If the user is not in any team, they have no permissions in this mock context.
	return pull.PermissionNone, nil
}

func (ghc *GitHubContext) RequestedReviewers() ([]*pull.Reviewer, error) {
	return ghc.reviewers, nil
}

func (ghc *GitHubContext) Teams() (map[string]pull.Permission, error) {
	// TODO: figure out if needed
	return nil, nil
}

func (ghc *GitHubContext) LatestStatuses() (map[string]string, error) {
	return ghc.statuses, nil
}

func (ghc *GitHubContext) LatestWorkflowRuns() (map[string][]string, error) {
	return ghc.workflowRuns, nil
}

func (ghc *GitHubContext) Labels() ([]string, error) {
	return ghc.labels, nil
}

func (ghc *GitHubContext) PushedAt(sha string) (time.Time, error) {
	return ghc.pushedAt[sha], nil
}

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

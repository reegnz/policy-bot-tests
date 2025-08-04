package models

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

// GitHubMembershipContext handles team and organization membership queries
type GitHubMembershipContext struct {
	orgMembers  map[string][]string
	teamMembers map[string][]string
}

// NewGitHubMembershipContext creates a NewGitHubMembershipContext
// the teamMembers and orgMembers maps are transformed so the keys are all lowercase
func NewGitHubMembershipContext(teamMembers, orgMembers map[string][]string) *GitHubMembershipContext {
	tm := map[string][]string{}
	for k, v := range teamMembers {
		tm[strings.ToLower(k)] = v
	}

	om := map[string][]string{}
	for k, v := range orgMembers {
		om[strings.ToLower(k)] = v
	}

	return &GitHubMembershipContext{
		teamMembers: tm,
		orgMembers:  om,
	}
}

func (mc *GitHubMembershipContext) IsTeamMember(team, user string) (bool, error) {
	return slices.Contains(mc.teamMembers[strings.ToLower(team)], user), nil
}

func (mc *GitHubMembershipContext) IsOrgMember(org, user string) (bool, error) {
	return slices.Contains(mc.orgMembers[strings.ToLower(org)], user), nil
}

func (mc *GitHubMembershipContext) TeamMembers(team string) ([]string, error) {
	return mc.teamMembers[strings.ToLower(team)], nil
}

func (mc *GitHubMembershipContext) OrganizationMembers(org string) ([]string, error) {
	return mc.orgMembers[strings.ToLower(org)], nil
}

// GitHubContext implements the pull.Context interface for policy evaluation
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

// PullRequest represents a GitHub pull request
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

// GitHubContext interface implementations
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

// NewGitHubContext creates a new GitHubContext from test context data
func NewGitHubContext(tc TestContext, reviews []*pull.Review, files []*pull.File, collaborators []*pull.Collaborator) *GitHubContext {
	return &GitHubContext{
		GitHubMembershipContext: *NewGitHubMembershipContext(tc.TeamMembers, tc.OrgMembers),
		evalTimestamp:           time.Now(),
		owner:                   tc.Owner,
		repo:                    tc.Repo,
		pr: PullRequest{
			author:      tc.Author,
			baseRefName: tc.PR.BaseRefName,
			headRefName: tc.PR.HeadRefName,
		},
		files:         files,
		reviews:       reviews,
		collaborators: collaborators,
		labels:        tc.Labels,
		statuses:      tc.Statuses,
		workflowRuns:  tc.WorkflowRuns,
	}
}

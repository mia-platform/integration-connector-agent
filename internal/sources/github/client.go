// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GitHubClient struct {
	token        string
	clientID     string
	clientSecret string
	accessToken  string
	organization string
	baseURL      string
	httpClient   *http.Client
	authType     string // "token" or "app"
}

// Repository represents a GitHub repository.
// JSON tags must match GitHub API response format exactly.
//
//nolint:tagliatelle // GitHub API uses snake_case, must maintain compatibility
type Repository struct {
	ID                        int64     `json:"id"`
	NodeID                    string    `json:"node_id"`
	Name                      string    `json:"name"`
	FullName                  string    `json:"full_name"`
	Description               string    `json:"description"`
	Private                   bool      `json:"private"`
	Fork                      bool      `json:"fork"`
	HTMLURL                   string    `json:"html_url"`
	CloneURL                  string    `json:"clone_url"`
	GitURL                    string    `json:"git_url"`
	SSHURL                    string    `json:"ssh_url"`
	SVNURL                    string    `json:"svn_url"`
	MirrorURL                 string    `json:"mirror_url"`
	Homepage                  string    `json:"homepage"`
	Language                  string    `json:"language"`
	ForksCount                int       `json:"forks_count"`
	StargazersCount           int       `json:"stargazers_count"`
	WatchersCount             int       `json:"watchers_count"`
	Size                      int       `json:"size"`
	DefaultBranch             string    `json:"default_branch"`
	OpenIssuesCount           int       `json:"open_issues_count"`
	IsTemplate                bool      `json:"is_template"`
	Topics                    []string  `json:"topics"`
	HasIssues                 bool      `json:"has_issues"`
	HasProjects               bool      `json:"has_projects"`
	HasWiki                   bool      `json:"has_wiki"`
	HasPages                  bool      `json:"has_pages"`
	HasDownloads              bool      `json:"has_downloads"`
	HasDiscussions            bool      `json:"has_discussions"`
	Archived                  bool      `json:"archived"`
	Disabled                  bool      `json:"disabled"`
	Visibility                string    `json:"visibility"`
	PushedAt                  time.Time `json:"pushed_at"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	AllowRebaseMerge          bool      `json:"allow_rebase_merge"`
	AllowSquashMerge          bool      `json:"allow_squash_merge"`
	AllowMergeCommit          bool      `json:"allow_merge_commit"`
	AllowAutoMerge            bool      `json:"allow_auto_merge"`
	DeleteBranchOnMerge       bool      `json:"delete_branch_on_merge"`
	AllowUpdateBranch         bool      `json:"allow_update_branch"`
	UseSquashPRTitleAsDefault bool      `json:"use_squash_pr_title_as_default"`
	SquashMergeCommitTitle    string    `json:"squash_merge_commit_title"`
	SquashMergeCommitMessage  string    `json:"squash_merge_commit_message"`
	MergeCommitTitle          string    `json:"merge_commit_title"`
	MergeCommitMessage        string    `json:"merge_commit_message"`
	AllowForking              bool      `json:"allow_forking"`
	WebCommitSignoffRequired  bool      `json:"web_commit_signoff_required"`
	SubscribersCount          int       `json:"subscribers_count"`
	NetworkCount              int       `json:"network_count"`
	License                   struct {
		Key    string `json:"key"`
		Name   string `json:"name"`
		SPDXID string `json:"spdx_id"`
		URL    string `json:"url"`
		NodeID string `json:"node_id"`
	} `json:"license"`
	Owner struct {
		Login             string `json:"login"`
		ID                int64  `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"owner"`
	Organization struct {
		Login             string `json:"login"`
		ID                int64  `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"organization"`
	ReadmeContent string `json:"readme_content,omitempty"` // README.md content
}

//nolint:tagliatelle // GitHub API uses snake_case, must maintain compatibility
type PullRequest struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	User   struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
	} `json:"user"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Head      struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"base"`
}

//nolint:tagliatelle // GitHub API uses snake_case, must maintain compatibility
type WorkflowRun struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Event  string `json:"event"`
	Actor  struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
	} `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	HTMLURL   string    `json:"html_url"`
	JobsURL   string    `json:"jobs_url"`
}

//nolint:tagliatelle // GitHub API uses snake_case, must maintain compatibility
type Issue struct {
	ID     int64  `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
	User   struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
	} `json:"user"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Labels    []struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	} `json:"labels"`
}

func NewGitHubClient(token, organization string) (*GitHubClient, error) {
	return &GitHubClient{
		token:        token,
		organization: organization,
		baseURL:      "https://api.github.com",
		authType:     "token",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func NewGitHubClientWithApp(clientID, clientSecret, organization string) (*GitHubClient, error) {
	client := &GitHubClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		organization: organization,
		baseURL:      "https://api.github.com",
		authType:     "app",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Get access token using client credentials flow
	if err := client.getAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	return client, nil
}

func (c *GitHubClient) getAccessToken() error {
	// Use OAuth 2.0 client credentials flow for GitHub App
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create access token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OAuth token request failed with status %d", resp.StatusCode)
	}

	//nolint:tagliatelle // GitHub API uses snake_case, must maintain compatibility
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error,omitempty"`
		ErrorDesc   string `json:"error_description,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResponse.Error != "" {
		return fmt.Errorf("OAuth error: %s - %s", tokenResponse.Error, tokenResponse.ErrorDesc)
	}

	c.accessToken = tokenResponse.AccessToken
	return nil
}

func (c *GitHubClient) makeRequest(ctx context.Context, endpoint string, result interface{}) error {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header based on auth type
	if c.authType == "app" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	} else {
		req.Header.Set("Authorization", "token "+c.token)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "mia-platform-integration-connector-agent")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d for %s", resp.StatusCode, url)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *GitHubClient) ListRepositories(ctx context.Context) ([]Repository, error) {
	var repositories []Repository
	endpoint := fmt.Sprintf("/orgs/%s/repos?per_page=100", c.organization)

	if err := c.makeRequest(ctx, endpoint, &repositories); err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	return repositories, nil
}

func (c *GitHubClient) ListPullRequests(ctx context.Context, repoName string) ([]PullRequest, error) {
	var pullRequests []PullRequest
	endpoint := fmt.Sprintf("/repos/%s/%s/pulls?state=all&per_page=100", c.organization, repoName)

	if err := c.makeRequest(ctx, endpoint, &pullRequests); err != nil {
		return nil, fmt.Errorf("failed to list pull requests for %s: %w", repoName, err)
	}

	return pullRequests, nil
}

func (c *GitHubClient) ListWorkflowRuns(ctx context.Context, repoName string) ([]WorkflowRun, error) {
	//nolint:tagliatelle // GitHub API uses snake_case, must maintain compatibility
	var response struct {
		WorkflowRuns []WorkflowRun `json:"workflow_runs"`
	}
	endpoint := fmt.Sprintf("/repos/%s/%s/actions/runs?per_page=100", c.organization, repoName)

	if err := c.makeRequest(ctx, endpoint, &response); err != nil {
		return nil, fmt.Errorf("failed to list workflow runs for %s: %w", repoName, err)
	}

	return response.WorkflowRuns, nil
}

func (c *GitHubClient) ListIssues(ctx context.Context, repoName string) ([]Issue, error) {
	var issues []Issue
	endpoint := fmt.Sprintf("/repos/%s/%s/issues?state=all&per_page=100", c.organization, repoName)

	if err := c.makeRequest(ctx, endpoint, &issues); err != nil {
		return nil, fmt.Errorf("failed to list issues for %s: %w", repoName, err)
	}

	return issues, nil
}

func (c *GitHubClient) GetRepositoryReadme(ctx context.Context, repoName string) (string, error) {
	var readme struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	endpoint := fmt.Sprintf("/repos/%s/%s/readme", c.organization, repoName)

	if err := c.makeRequest(ctx, endpoint, &readme); err != nil {
		// README not found is not an error - return empty string
		return "", nil
	}

	// GitHub API returns base64 encoded content
	if readme.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(readme.Content)
		if err != nil {
			return "", fmt.Errorf("failed to decode README content: %w", err)
		}
		return string(decoded), nil
	}

	// If not base64 encoded, return as is
	return readme.Content, nil
}

// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GitLabClient struct {
	token      string
	baseURL    string
	httpClient *http.Client
	group      string // Group/Organization name
}

//nolint:tagliatelle // GitLab API uses snake_case, must maintain compatibility
type Project struct {
	ID                               int64     `json:"id"`
	Name                             string    `json:"name"`
	NameWithNamespace                string    `json:"name_with_namespace"`
	Path                             string    `json:"path"`
	PathWithNamespace                string    `json:"path_with_namespace"`
	Description                      string    `json:"description"`
	DefaultBranch                    string    `json:"default_branch"`
	TagList                          []string  `json:"tag_list"`
	SSHURLToRepo                     string    `json:"ssh_url_to_repo"`
	HTTPURLToRepo                    string    `json:"http_url_to_repo"`
	WebURL                           string    `json:"web_url"`
	ReadmeURL                        string    `json:"readme_url"`
	AvatarURL                        string    `json:"avatar_url"`
	ForksCount                       int       `json:"forks_count"`
	StarCount                        int       `json:"star_count"`
	LastActivityAt                   time.Time `json:"last_activity_at"`
	CreatedAt                        time.Time `json:"created_at"`
	UpdatedAt                        time.Time `json:"updated_at"`
	Visibility                       string    `json:"visibility"`
	IssuesEnabled                    bool      `json:"issues_enabled"`
	MergeRequestsEnabled             bool      `json:"merge_requests_enabled"`
	WikiEnabled                      bool      `json:"wiki_enabled"`
	JobsEnabled                      bool      `json:"jobs_enabled"`
	SnippetsEnabled                  bool      `json:"snippets_enabled"`
	ContainerRegistryEnabled         bool      `json:"container_registry_enabled"`
	ServiceDeskEnabled               bool      `json:"service_desk_enabled"`
	CanCreateMergeRequestIn          bool      `json:"can_create_merge_request_in"`
	IssuesAccessLevel                string    `json:"issues_access_level"`
	RepositoryAccessLevel            string    `json:"repository_access_level"`
	MergeRequestsAccessLevel         string    `json:"merge_requests_access_level"`
	ForkingAccessLevel               string    `json:"forking_access_level"`
	WikiAccessLevel                  string    `json:"wiki_access_level"`
	BuildsAccessLevel                string    `json:"builds_access_level"`
	SnippetsAccessLevel              string    `json:"snippets_access_level"`
	PagesAccessLevel                 string    `json:"pages_access_level"`
	AnalyticsAccessLevel             string    `json:"analytics_access_level"`
	ContainerRegistryAccessLevel     string    `json:"container_registry_access_level"`
	SecurityAndComplianceAccessLevel string    `json:"security_and_compliance_access_level"`
	ReleasesAccessLevel              string    `json:"releases_access_level"`
	EnvironmentsAccessLevel          string    `json:"environments_access_level"`
	FeatureFlagsAccessLevel          string    `json:"feature_flags_access_level"`
	InfrastructureAccessLevel        string    `json:"infrastructure_access_level"`
	MonitorAccessLevel               string    `json:"monitor_access_level"`
	ModelExperimentsAccessLevel      string    `json:"model_experiments_access_level"`
	ModelRegistryAccessLevel         string    `json:"model_registry_access_level"`
	Archived                         bool      `json:"archived"`
	EmptyRepo                        bool      `json:"empty_repo"`
	LicenseURL                       string    `json:"license_url"`
	License                          struct {
		Key       string `json:"key"`
		Name      string `json:"name"`
		Nickname  string `json:"nickname"`
		HTMLURL   string `json:"html_url"`
		SourceURL string `json:"source_url"`
	} `json:"license"`
	Owner struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"owner"`
	Namespace struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Path      string `json:"path"`
		Kind      string `json:"kind"`
		FullPath  string `json:"full_path"`
		ParentID  int64  `json:"parent_id"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"namespace"`
	ReadmeContent string `json:"readme_content,omitempty"` // README content
}

//nolint:tagliatelle // GitLab API uses snake_case, must maintain compatibility
type MergeRequest struct {
	ID          int64  `json:"id"`
	IID         int    `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"`
	MergedBy    struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"merged_by"`
	MergeUser struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"merge_user"`
	MergedAt   time.Time `json:"merged_at"`
	PreparedAt time.Time `json:"prepared_at"`
	ClosedBy   struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"closed_by"`
	ClosedAt       time.Time `json:"closed_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	TargetBranch   string    `json:"target_branch"`
	SourceBranch   string    `json:"source_branch"`
	UserNotesCount int       `json:"user_notes_count"`
	Upvotes        int       `json:"upvotes"`
	Downvotes      int       `json:"downvotes"`
	Author         struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"author"`
	Assignees []struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"assignees"`
	Assignee struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"assignee"`
	Reviewers []struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"reviewers"`
	SourceProjectID int64    `json:"source_project_id"`
	TargetProjectID int64    `json:"target_project_id"`
	Labels          []string `json:"labels"`
	Draft           bool     `json:"draft"`
	WorkInProgress  bool     `json:"work_in_progress"`
	Milestone       struct {
		ID          int64     `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		State       string    `json:"state"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		GroupID     int64     `json:"group_id"`
		ProjectID   int64     `json:"project_id"`
		WebURL      string    `json:"web_url"`
	} `json:"milestone"`
	MergeWhenPipelineSucceeds bool   `json:"merge_when_pipeline_succeeds"`
	MergeStatus               string `json:"merge_status"`
	DetailedMergeStatus       string `json:"detailed_merge_status"`
	SHA                       string `json:"sha"`
	MergeCommitSHA            string `json:"merge_commit_sha"`
	SquashCommitSHA           string `json:"squash_commit_sha"`
	DiscussionLocked          bool   `json:"discussion_locked"`
	ShouldRemoveSourceBranch  bool   `json:"should_remove_source_branch"`
	ForceRemoveSourceBranch   bool   `json:"force_remove_source_branch"`
	Reference                 string `json:"reference"`
	References                struct {
		Short    string `json:"short"`
		Relative string `json:"relative"`
		Full     string `json:"full"`
	} `json:"references"`
	WebURL    string `json:"web_url"`
	TimeStats struct {
		TimeEstimate        int    `json:"time_estimate"`
		TotalTimeSpent      int    `json:"total_time_spent"`
		HumanTimeEstimate   string `json:"human_time_estimate"`
		HumanTotalTimeSpent string `json:"human_total_time_spent"`
	} `json:"time_stats"`
	Squash               bool `json:"squash"`
	TaskCompletionStatus struct {
		Count          int `json:"count"`
		CompletedCount int `json:"completed_count"`
	} `json:"task_completion_status"`
	HasConflicts                bool `json:"has_conflicts"`
	BlockingDiscussionsResolved bool `json:"blocking_discussions_resolved"`
	ApprovalsBeforeMerge        int  `json:"approvals_before_merge"`
}

//nolint:tagliatelle // GitLab API uses snake_case, must maintain compatibility
type Pipeline struct {
	ID         int64     `json:"id"`
	IID        int       `json:"iid"`
	ProjectID  int64     `json:"project_id"`
	SHA        string    `json:"sha"`
	Ref        string    `json:"ref"`
	Status     string    `json:"status"`
	Source     string    `json:"source"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	WebURL     string    `json:"web_url"`
	BeforeSHA  string    `json:"before_sha"`
	Tag        bool      `json:"tag"`
	YamlErrors string    `json:"yaml_errors"`
	User       struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"user"`
	StartedAt      time.Time `json:"started_at"`
	FinishedAt     time.Time `json:"finished_at"`
	CommittedAt    time.Time `json:"committed_at"`
	Duration       int       `json:"duration"`
	QueuedDuration int       `json:"queued_duration"`
	Coverage       string    `json:"coverage"`
	DetailedStatus struct {
		Icon         string `json:"icon"`
		Text         string `json:"text"`
		Label        string `json:"label"`
		Group        string `json:"group"`
		Tooltip      string `json:"tooltip"`
		HasDetails   bool   `json:"has_details"`
		DetailsPath  string `json:"details_path"`
		Illustration struct {
			Image string `json:"image"`
		} `json:"illustration"`
		Favicon string `json:"favicon"`
	} `json:"detailed_status"`
	Name string `json:"name"`
}

//nolint:tagliatelle // GitLab API uses snake_case, must maintain compatibility
type Release struct {
	TagName         string    `json:"tag_name"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DescriptionHTML string    `json:"description_html"`
	CreatedAt       time.Time `json:"created_at"`
	ReleasedAt      time.Time `json:"released_at"`
	Author          struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"author"`
	Commit struct {
		ID             string    `json:"id"`
		ShortID        string    `json:"short_id"`
		Title          string    `json:"title"`
		AuthorName     string    `json:"author_name"`
		AuthorEmail    string    `json:"author_email"`
		AuthoredDate   time.Time `json:"authored_date"`
		CommitterName  string    `json:"committer_name"`
		CommitterEmail string    `json:"committer_email"`
		CommittedDate  time.Time `json:"committed_date"`
		CreatedAt      time.Time `json:"created_at"`
		Message        string    `json:"message"`
		ParentIDs      []string  `json:"parent_ids"`
		WebURL         string    `json:"web_url"`
	} `json:"commit"`
	Milestones []struct {
		ID          int64     `json:"id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		State       string    `json:"state"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		GroupID     int64     `json:"group_id"`
		ProjectID   int64     `json:"project_id"`
		WebURL      string    `json:"web_url"`
	} `json:"milestones"`
	CommitPath  string `json:"commit_path"`
	TagPath     string `json:"tag_path"`
	EvidenceSHA string `json:"evidence_sha"`
	Assets      struct {
		Count   int `json:"count"`
		Sources []struct {
			Format string `json:"format"`
			URL    string `json:"url"`
		} `json:"sources"`
		Links []struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			URL      string `json:"url"`
			External bool   `json:"external"`
			LinkType string `json:"link_type"`
		} `json:"links"`
		EvidenceFilePath string `json:"evidence_file_path"`
	} `json:"assets"`
	UpcomingRelease   bool `json:"upcoming_release"`
	HistoricalRelease bool `json:"historical_release"`
}

func NewGitLabClient(token, baseURL, group string) (*GitLabClient, error) {
	return &GitLabClient{
		token:   token,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		group:   group,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *GitLabClient) makeRequest(ctx context.Context, endpoint string, result interface{}) error {
	fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "mia-platform-integration-connector-agent")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitLab API returned status %d for %s", resp.StatusCode, fullURL)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (c *GitLabClient) ListProjects(ctx context.Context) ([]Project, error) {
	var projects []Project
	endpoint := "/api/v4/projects?per_page=100"

	if err := c.makeRequest(ctx, endpoint, &projects); err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, nil
}

func (c *GitLabClient) ListMergeRequests(ctx context.Context, projectID int64) ([]MergeRequest, error) {
	var mergeRequests []MergeRequest
	endpoint := fmt.Sprintf("/api/v4/projects/%d/merge_requests?state=all&per_page=100", projectID)

	if err := c.makeRequest(ctx, endpoint, &mergeRequests); err != nil {
		return nil, fmt.Errorf("failed to list merge requests for project %d: %w", projectID, err)
	}

	return mergeRequests, nil
}

func (c *GitLabClient) ListPipelines(ctx context.Context, projectID int64) ([]Pipeline, error) {
	var pipelines []Pipeline
	endpoint := fmt.Sprintf("/api/v4/projects/%d/pipelines?per_page=100", projectID)

	if err := c.makeRequest(ctx, endpoint, &pipelines); err != nil {
		return nil, fmt.Errorf("failed to list pipelines for project %d: %w", projectID, err)
	}

	return pipelines, nil
}

func (c *GitLabClient) ListReleases(ctx context.Context, projectID int64) ([]Release, error) {
	var releases []Release
	endpoint := fmt.Sprintf("/api/v4/projects/%d/releases?per_page=100", projectID)

	if err := c.makeRequest(ctx, endpoint, &releases); err != nil {
		return nil, fmt.Errorf("failed to list releases for project %d: %w", projectID, err)
	}

	return releases, nil
}

func (c *GitLabClient) GetProjectReadme(ctx context.Context, projectID int64) (string, error) {
	endpoint := fmt.Sprintf("/api/v4/projects/%d/repository/files/README.md/raw?ref=main", projectID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s%s", c.baseURL, endpoint), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("User-Agent", "mia-platform-integration-connector-agent")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// README not found is not an error - return empty string
		return "", nil
	}

	// GitLab returns raw content directly - use io.ReadAll to safely read all content
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read README content: %w", err)
	}

	return string(buf), nil
}

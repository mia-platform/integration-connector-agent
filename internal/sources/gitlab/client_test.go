// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package gitlab

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitLabClient(t *testing.T) {
	client, err := NewGitLabClient("test-token", "https://gitlab.example.com", "test-group")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "test-token", client.token)
	assert.Equal(t, "https://gitlab.example.com", client.baseURL)
	assert.Equal(t, "test-group", client.group)
}

func TestGitLabClient_ListProjects(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("PRIVATE-TOKEN"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"id": 123,
				"name": "test-project",
				"path": "test-project",
				"path_with_namespace": "test-group/test-project",
				"description": "A test project",
				"default_branch": "main",
				"visibility": "private",
				"web_url": "https://gitlab.example.com/test-group/test-project",
				"http_url_to_repo": "https://gitlab.example.com/test-group/test-project.git",
				"ssh_url_to_repo": "git@gitlab.example.com:test-group/test-project.git",
				"star_count": 5,
				"forks_count": 2,
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-02T00:00:00Z",
				"last_activity_at": "2024-01-03T00:00:00Z",
				"namespace": {
					"id": 456,
					"name": "Test Group",
					"path": "test-group",
					"kind": "group",
					"full_path": "test-group"
				}
			}
		]`))
	}))
	defer server.Close()

	client, err := NewGitLabClient("test-token", server.URL, "test-group")
	require.NoError(t, err)

	projects, err := client.ListProjects(t.Context())
	require.NoError(t, err)
	require.Len(t, projects, 1)

	project := projects[0]
	assert.Equal(t, int64(123), project.ID)
	assert.Equal(t, "test-project", project.Name)
	assert.Equal(t, "test-group/test-project", project.PathWithNamespace)
	assert.Equal(t, "A test project", project.Description)
	assert.Equal(t, "main", project.DefaultBranch)
	assert.Equal(t, "private", project.Visibility)
	assert.Equal(t, 5, project.StarCount)
	assert.Equal(t, 2, project.ForksCount)
}

func TestGitLabClient_ListMergeRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/123/merge_requests", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("PRIVATE-TOKEN"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"id": 456,
				"iid": 1,
				"title": "Test MR",
				"description": "A test merge request",
				"state": "opened",
				"merge_status": "can_be_merged",
				"target_branch": "main",
				"source_branch": "feature",
				"web_url": "https://gitlab.example.com/test-group/test-project/-/merge_requests/1",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-02T00:00:00Z",
				"author": {
					"id": 789,
					"username": "testuser",
					"name": "Test User"
				}
			}
		]`))
	}))
	defer server.Close()

	client, err := NewGitLabClient("test-token", server.URL, "test-group")
	require.NoError(t, err)

	mrs, err := client.ListMergeRequests(t.Context(), 123)
	require.NoError(t, err)
	require.Len(t, mrs, 1)

	mr := mrs[0]
	assert.Equal(t, int64(456), mr.ID)
	assert.Equal(t, 1, mr.IID)
	assert.Equal(t, "Test MR", mr.Title)
	assert.Equal(t, "A test merge request", mr.Description)
	assert.Equal(t, "opened", mr.State)
	assert.Equal(t, "can_be_merged", mr.MergeStatus)
	assert.Equal(t, "main", mr.TargetBranch)
	assert.Equal(t, "feature", mr.SourceBranch)
}

func TestGitLabClient_ListPipelines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/123/pipelines", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("PRIVATE-TOKEN"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"id": 789,
				"iid": 1,
				"project_id": 123,
				"sha": "abc123",
				"ref": "main",
				"status": "success",
				"source": "push",
				"created_at": "2024-01-01T00:00:00Z",
				"updated_at": "2024-01-02T00:00:00Z",
				"web_url": "https://gitlab.example.com/test-group/test-project/-/pipelines/789",
				"user": {
					"id": 456,
					"username": "testuser",
					"name": "Test User"
				}
			}
		]`))
	}))
	defer server.Close()

	client, err := NewGitLabClient("test-token", server.URL, "test-group")
	require.NoError(t, err)

	pipelines, err := client.ListPipelines(t.Context(), 123)
	require.NoError(t, err)
	require.Len(t, pipelines, 1)

	pipeline := pipelines[0]
	assert.Equal(t, int64(789), pipeline.ID)
	assert.Equal(t, 1, pipeline.IID)
	assert.Equal(t, int64(123), pipeline.ProjectID)
	assert.Equal(t, "abc123", pipeline.SHA)
	assert.Equal(t, "main", pipeline.Ref)
	assert.Equal(t, "success", pipeline.Status)
	assert.Equal(t, "push", pipeline.Source)
}

func TestGitLabClient_ListReleases(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/123/releases", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("PRIVATE-TOKEN"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"tag_name": "v1.0.0",
				"name": "Release 1.0.0",
				"description": "First release",
				"created_at": "2024-01-01T00:00:00Z",
				"released_at": "2024-01-01T00:00:00Z",
				"author": {
					"id": 456,
					"username": "testuser",
					"name": "Test User"
				},
				"commit": {
					"id": "abc123",
					"short_id": "abc123",
					"title": "Release commit"
				}
			}
		]`))
	}))
	defer server.Close()

	client, err := NewGitLabClient("test-token", server.URL, "test-group")
	require.NoError(t, err)

	releases, err := client.ListReleases(t.Context(), 123)
	require.NoError(t, err)
	require.Len(t, releases, 1)

	release := releases[0]
	assert.Equal(t, "v1.0.0", release.TagName)
	assert.Equal(t, "Release 1.0.0", release.Name)
	assert.Equal(t, "First release", release.Description)
	assert.Equal(t, int64(456), release.Author.ID)
	assert.Equal(t, "testuser", release.Author.Username)
	assert.Equal(t, "abc123", release.Commit.ID)
}

func TestGitLabClient_GetProjectReadme(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4/projects/123/repository/files/README.md/raw", r.URL.Path)
		assert.Equal(t, "test-token", r.Header.Get("PRIVATE-TOKEN"))

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Test Project\n\nThis is a test project."))
	}))
	defer server.Close()

	client, err := NewGitLabClient("test-token", server.URL, "test-group")
	require.NoError(t, err)

	readme, err := client.GetProjectReadme(t.Context(), 123)
	require.NoError(t, err)
	assert.Equal(t, "# Test Project\n\nThis is a test project.", readme)
}

func TestGitLabClient_GetProjectReadme_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewGitLabClient("test-token", server.URL, "test-group")
	require.NoError(t, err)

	readme, err := client.GetProjectReadme(t.Context(), 123)
	require.NoError(t, err)
	assert.Empty(t, readme)
}

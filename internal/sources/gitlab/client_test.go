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
	assert.Equal(t, "https://gitlab.example.com", client.baseURL.String())
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
	assert.EqualValues(t, 123, project["id"])
	assert.Equal(t, "test-project", project["name"])
	assert.Equal(t, "test-group/test-project", project["path_with_namespace"])
	assert.Equal(t, "A test project", project["description"])
	assert.Equal(t, "main", project["default_branch"])
	assert.Equal(t, "private", project["visibility"])
	assert.EqualValues(t, 5, project["star_count"])
	assert.EqualValues(t, 2, project["forks_count"])
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

	mrs, err := client.ListMergeRequests(t.Context(), "123")
	require.NoError(t, err)
	require.Len(t, mrs, 1)

	mr := mrs[0]
	assert.EqualValues(t, 456, mr["id"])
	assert.EqualValues(t, 1, mr["iid"])
	assert.Equal(t, "Test MR", mr["title"])
	assert.Equal(t, "A test merge request", mr["description"])
	assert.Equal(t, "opened", mr["state"])
	assert.Equal(t, "can_be_merged", mr["merge_status"])
	assert.Equal(t, "main", mr["target_branch"])
	assert.Equal(t, "feature", mr["source_branch"])
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

	pipelines, err := client.ListPipelines(t.Context(), "123")
	require.NoError(t, err)
	require.Len(t, pipelines, 1)

	pipeline := pipelines[0]
	assert.EqualValues(t, 789, pipeline["id"])
	assert.EqualValues(t, 1, pipeline["iid"])
	assert.EqualValues(t, 123, pipeline["project_id"])
	assert.Equal(t, "abc123", pipeline["sha"])
	assert.Equal(t, "main", pipeline["ref"])
	assert.Equal(t, "success", pipeline["status"])
	assert.Equal(t, "push", pipeline["source"])
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

	releases, err := client.ListReleases(t.Context(), "123")
	require.NoError(t, err)
	require.Len(t, releases, 1)

	release := releases[0]
	assert.Equal(t, "v1.0.0", release["tag_name"])
	assert.Equal(t, "Release 1.0.0", release["name"])
	assert.Equal(t, "First release", release["description"])
	assert.Equal(t, map[string]any{"id": float64(456), "username": "testuser", "name": "Test User"}, release["author"])
	assert.Equal(t, map[string]any{"id": "abc123", "short_id": "abc123", "title": "Release commit"}, release["commit"])
}

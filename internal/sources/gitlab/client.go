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
	"net/url"
	"time"
)

type GitLabClient struct {
	token   string
	baseURL *url.URL
	group   string // Group/Organization name
}

func NewGitLabClient(token, baseURL, group string) (*GitLabClient, error) {
	url, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid GitLab base URL: %w", err)
	}

	return &GitLabClient{
		token:   token,
		baseURL: url,
		group:   group,
	}, nil
}

func (c *GitLabClient) makeRequest(ctx context.Context, endpoint *url.URL) ([]map[string]any, error) {
	requestURL := c.baseURL.ResolveReference(endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "mia-platform-integration-connector-agent")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API returned status %d for %s", resp.StatusCode, requestURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	result := []map[string]any{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func (c *GitLabClient) ListProjects(ctx context.Context) ([]map[string]any, error) {
	response, err := c.makeRequest(ctx, &url.URL{
		Path:     "/api/v4/projects",
		RawQuery: "per_page=100",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	return response, nil
}

func (c *GitLabClient) ListMergeRequests(ctx context.Context, projectID string) ([]map[string]any, error) {
	response, err := c.makeRequest(ctx, &url.URL{
		Path:     "/api/v4/projects/" + projectID + "/merge_requests",
		RawQuery: "state=all&per_page=100",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list merge requests for project %s: %w", projectID, err)
	}

	return response, nil
}

func (c *GitLabClient) ListPipelines(ctx context.Context, projectID string) ([]map[string]any, error) {
	response, err := c.makeRequest(ctx, &url.URL{
		Path:     "/api/v4/projects/" + projectID + "/pipelines",
		RawQuery: "per_page=100",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines for project %s: %w", projectID, err)
	}

	return response, nil
}

func (c *GitLabClient) ListReleases(ctx context.Context, projectID string) ([]map[string]any, error) {
	response, err := c.makeRequest(ctx, &url.URL{
		Path:     "/api/v4/projects/" + projectID + "/releases",
		RawQuery: "per_page=100",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list releases for project %s: %w", projectID, err)
	}

	return response, nil
}

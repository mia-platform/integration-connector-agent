// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type ConfluenceClient struct {
	username   string
	apiToken   string
	baseURL    string
	httpClient *http.Client
	log        *logrus.Logger
}

//nolint:tagliatelle // Confluence API uses snake_case/_links, must maintain compatibility
type Space struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Description struct {
		Plain struct {
			Value string `json:"value"`
		} `json:"plain"`
	} `json:"description"`
	Homepage struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Title string `json:"title"`
	} `json:"homepage"`
	CreatedAt time.Time `json:"createdAt"`
	Links     struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

//nolint:tagliatelle // Confluence API uses snake_case/_links, must maintain compatibility
type Page struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Title    string `json:"title"`
	SpaceID  string `json:"spaceId"`
	ParentID string `json:"parentId"`
	AuthorID string `json:"authorId"`
	Version  struct {
		Number    int       `json:"number"`
		Message   string    `json:"message"`
		CreatedAt time.Time `json:"createdAt"`
		AuthorID  string    `json:"authorId"`
	} `json:"version"`
	Body struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
		AtlasDocFormat struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"atlas_doc_format"`
	} `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	Links     struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

//nolint:tagliatelle // Confluence API uses snake_case/_links, must maintain compatibility
type SpaceListResponse struct {
	Results []Space `json:"results"`
	Start   int     `json:"start"`
	Limit   int     `json:"limit"`
	Size    int     `json:"size"`
	Links   struct {
		Next string `json:"next"`
	} `json:"_links"`
}

//nolint:tagliatelle // Confluence API uses snake_case/_links, must maintain compatibility
type PageListResponse struct {
	Results []Page `json:"results"`
	Start   int    `json:"start"`
	Limit   int    `json:"limit"`
	Size    int    `json:"size"`
	Links   struct {
		Next string `json:"next"`
	} `json:"_links"`
}

func NewConfluenceClient(username, apiToken, baseURL string, log *logrus.Logger) (*ConfluenceClient, error) {
	// Ensure baseURL ends without trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &ConfluenceClient{
		username: username,
		apiToken: apiToken,
		baseURL:  baseURL,
		log:      log,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *ConfluenceClient) makeRequest(ctx context.Context, endpoint string, result interface{}) error {
	fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	requestStartTime := time.Now()

	// Debug log before making the request
	c.log.WithFields(logrus.Fields{
		"sourceType":    "confluence",
		"operation":     "api-request",
		"method":        "GET",
		"endpoint":      endpoint,
		"fullURL":       fullURL,
		"clientTimeout": "30s",
	}).Debug("preparing to make Atlassian API request")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		c.log.WithFields(logrus.Fields{
			"sourceType": "confluence",
			"operation":  "api-request",
			"endpoint":   endpoint,
			"error":      err.Error(),
		}).Error("failed to create HTTP request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic authentication
	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mia-platform-integration-connector-agent")

	c.log.WithFields(logrus.Fields{
		"sourceType": "confluence",
		"operation":  "api-request",
		"endpoint":   endpoint,
		"headers": map[string]string{
			"Accept":       "application/json",
			"Content-Type": "application/json",
			"User-Agent":   "mia-platform-integration-connector-agent",
		},
	}).Debug("HTTP request prepared, making call to Atlassian API...")

	// Make the actual HTTP request
	requestCallStartTime := time.Now()
	resp, err := c.httpClient.Do(req)
	requestCallDuration := time.Since(requestCallStartTime)

	if err != nil {
		c.log.WithFields(logrus.Fields{
			"sourceType":       "confluence",
			"operation":        "api-request",
			"endpoint":         endpoint,
			"fullURL":          fullURL,
			"requestDuration":  requestCallDuration.String(),
			"totalRequestTime": time.Since(requestStartTime).String(),
			"error":            err.Error(),
		}).Error("HTTP request to Atlassian API failed")
		return fmt.Errorf("failed to make request to %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	c.log.WithFields(logrus.Fields{
		"sourceType":      "confluence",
		"operation":       "api-request",
		"endpoint":        endpoint,
		"statusCode":      resp.StatusCode,
		"responseStatus":  resp.Status,
		"requestDuration": requestCallDuration.String(),
		"contentLength":   resp.ContentLength,
		"contentType":     resp.Header.Get("Content-Type"),
	}).Debug("received response from Atlassian API")

	if resp.StatusCode != http.StatusOK {
		// Try to read the response body for more error details
		bodyBytes := make([]byte, 1024)
		n, _ := resp.Body.Read(bodyBytes)
		bodyText := string(bodyBytes[:n])

		c.log.WithFields(logrus.Fields{
			"sourceType":        "confluence",
			"operation":         "api-request",
			"endpoint":          endpoint,
			"statusCode":        resp.StatusCode,
			"responseStatus":    resp.Status,
			"requestDuration":   requestCallDuration.String(),
			"totalRequestTime":  time.Since(requestStartTime).String(),
			"errorResponseBody": bodyText,
		}).Error("Atlassian API returned non-200 status code")

		return fmt.Errorf("confluence API returned status %d for %s: %s", resp.StatusCode, fullURL, bodyText)
	}

	// Decode the response
	decodeStartTime := time.Now()
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		decodeDuration := time.Since(decodeStartTime)
		c.log.WithFields(logrus.Fields{
			"sourceType":       "confluence",
			"operation":        "api-request",
			"endpoint":         endpoint,
			"requestDuration":  requestCallDuration.String(),
			"decodeDuration":   decodeDuration.String(),
			"totalRequestTime": time.Since(requestStartTime).String(),
			"error":            err.Error(),
		}).Error("failed to decode JSON response from Atlassian API")
		return fmt.Errorf("failed to decode response: %w", err)
	}

	decodeDuration := time.Since(decodeStartTime)
	totalDuration := time.Since(requestStartTime)

	c.log.WithFields(logrus.Fields{
		"sourceType":       "confluence",
		"operation":        "api-request",
		"endpoint":         endpoint,
		"statusCode":       resp.StatusCode,
		"requestDuration":  requestCallDuration.String(),
		"decodeDuration":   decodeDuration.String(),
		"totalRequestTime": totalDuration.String(),
	}).Debug("successfully completed Atlassian API request")

	return nil
}

func (c *ConfluenceClient) ListSpaces(ctx context.Context) ([]Space, error) {
	return c.ListSpacesWithKeysFilter(ctx, nil)
}

func (c *ConfluenceClient) ListSpacesWithKeysFilter(ctx context.Context, spaceKeys []string) ([]Space, error) {
	listStartTime := time.Now()
	var allSpaces []Space
	start := 0
	limit := 50
	pageCount := 0
	totalProcessed := 0

	c.log.WithFields(logrus.Fields{
		"sourceType":      "confluence",
		"operation":       "list-spaces",
		"spaceKeys":       spaceKeys,
		"pageSize":        limit,
		"serverFiltering": len(spaceKeys) > 0,
	}).Info("starting to list spaces from Confluence API with server-side filtering")

	for {
		pageCount++
		pageStartTime := time.Now()

		params := url.Values{}
		params.Set("start", strconv.Itoa(start))
		params.Set("limit", strconv.Itoa(limit))
		params.Set("status", "current") // Add server-side space keys filtering if specified
		if len(spaceKeys) > 0 {
			// Use the 'keys' parameter for server-side filtering by space keys
			// This searches for spaces with the specified keys
			for _, key := range spaceKeys {
				params.Add("keys", key)
			}
		}

		endpoint := "/wiki/api/v2/spaces?" + params.Encode()

		c.log.WithFields(logrus.Fields{
			"sourceType":      "confluence",
			"operation":       "list-spaces",
			"pageNumber":      pageCount,
			"startIndex":      start,
			"pageLimit":       limit,
			"endpoint":        endpoint,
			"totalProcessed":  totalProcessed,
			"serverFiltering": len(spaceKeys) > 0,
			"spaceKeys":       spaceKeys,
		}).Debug("requesting spaces page from Atlassian API")

		var response SpaceListResponse
		if err := c.makeRequest(ctx, endpoint, &response); err != nil {
			c.log.WithFields(logrus.Fields{
				"sourceType":    "confluence",
				"operation":     "list-spaces",
				"pageNumber":    pageCount,
				"startIndex":    start,
				"totalDuration": time.Since(listStartTime).String(),
				"error":         err.Error(),
			}).Error("failed to retrieve spaces page from Atlassian API")
			return nil, fmt.Errorf("failed to list spaces: %w", err)
		}

		pageDuration := time.Since(pageStartTime)
		c.log.WithFields(logrus.Fields{
			"sourceType":    "confluence",
			"operation":     "list-spaces",
			"pageNumber":    pageCount,
			"receivedCount": len(response.Results),
			"responseStart": response.Start,
			"responseLimit": response.Limit,
			"responseSize":  response.Size,
			"hasNextPage":   response.Links.Next != "",
			"pageDuration":  pageDuration.String(),
		}).Debug("received spaces page from Atlassian API")

		// Server-side filtering is handled by the API parameters, so no client-side filtering needed
		allSpaces = append(allSpaces, response.Results...)
		totalProcessed += len(response.Results)

		// Log received spaces for debugging
		for _, space := range response.Results {
			c.log.WithFields(logrus.Fields{
				"sourceType": "confluence",
				"operation":  "list-spaces",
				"spaceName":  space.Name,
				"spaceKey":   space.Key,
				"spaceId":    space.ID,
			}).Debug("received space from server-side filtered API")
		}

		c.log.WithFields(logrus.Fields{
			"sourceType":          "confluence",
			"operation":           "list-spaces",
			"pageNumber":          pageCount,
			"receivedSpacesCount": len(response.Results),
			"totalSpacesReceived": len(allSpaces),
			"totalProcessed":      totalProcessed,
			"serverSideFiltered":  len(spaceKeys) > 0,
			"spaceKeys":           spaceKeys,
			"pageTotalDuration":   time.Since(pageStartTime).String(),
			"totalElapsedTime":    time.Since(listStartTime).String(),
		}).Info("processed spaces page with server-side filtering")

		// Check if we have more results
		if len(response.Results) < limit || response.Links.Next == "" {
			c.log.WithFields(logrus.Fields{
				"sourceType":      "confluence",
				"operation":       "list-spaces",
				"lastPageResults": len(response.Results),
				"pageLimit":       limit,
				"hasNextLink":     response.Links.Next != "",
				"reason":          "end of pagination reached",
			}).Debug("pagination completed - no more pages available")
			break
		}

		start += limit

		c.log.WithFields(logrus.Fields{
			"sourceType":     "confluence",
			"operation":      "list-spaces",
			"nextStartIndex": start,
			"pageNumber":     pageCount,
			"totalProcessed": totalProcessed,
		}).Debug("moving to next page of spaces")
	}

	totalDuration := time.Since(listStartTime)
	c.log.WithFields(logrus.Fields{
		"sourceType":          "confluence",
		"operation":           "list-spaces",
		"totalPagesProcessed": pageCount,
		"totalSpacesReceived": len(allSpaces),
		"spaceKeys":           spaceKeys,
		"serverSideFiltered":  len(spaceKeys) > 0,
		"totalDuration":       totalDuration.String(),
		"averagePageDuration": fmt.Sprintf("%.2fs", totalDuration.Seconds()/float64(pageCount)),
	}).Info("completed listing spaces from Confluence API with server-side filtering")

	return allSpaces, nil
}

func (c *ConfluenceClient) ListPages(ctx context.Context, spaceKey string) ([]Page, error) {
	var allPages []Page
	start := 0
	limit := 50

	for {
		params := url.Values{}
		params.Set("space-key", spaceKey)
		params.Set("start", strconv.Itoa(start))
		params.Set("limit", strconv.Itoa(limit))
		params.Set("status", "current")
		params.Set("body-format", "storage")

		endpoint := "/wiki/api/v2/pages?" + params.Encode()

		var response PageListResponse
		if err := c.makeRequest(ctx, endpoint, &response); err != nil {
			return nil, fmt.Errorf("failed to list pages for space %s: %w", spaceKey, err)
		}

		allPages = append(allPages, response.Results...)

		// Check if we have more results
		if len(response.Results) < limit || response.Links.Next == "" {
			break
		}

		start += limit
	}

	return allPages, nil
}

func (c *ConfluenceClient) GetPage(ctx context.Context, pageID string) (*Page, error) {
	params := url.Values{}
	params.Set("body-format", "storage")

	endpoint := fmt.Sprintf("/wiki/api/v2/pages/%s?%s", pageID, params.Encode())

	var page Page
	if err := c.makeRequest(ctx, endpoint, &page); err != nil {
		return nil, fmt.Errorf("failed to get page %s: %w", pageID, err)
	}

	return &page, nil
}

func (c *ConfluenceClient) GetSpace(ctx context.Context, spaceKey string) (*Space, error) {
	endpoint := "/wiki/api/v2/spaces/" + spaceKey

	var space Space
	if err := c.makeRequest(ctx, endpoint, &space); err != nil {
		return nil, fmt.Errorf("failed to get space %s: %w", spaceKey, err)
	}

	return &space, nil
}

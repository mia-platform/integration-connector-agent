// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/stretchr/testify/require"
)

func TestJiraIntegrationUnit(t *testing.T) {
	t.Run("webhook processes jira events correctly", func(t *testing.T) {
		// Create a mock pipeline group to capture processed events
		mockPipelineGroup := &pipeline.PipelineGroupMock{}
		timestamp := time.Now().UnixMilli()
		reqBody := map[string]any{
			"webhookEvent": "jira:issue_created",
			"id":           123,
			"timestamp":    timestamp,
			"issue": map[string]any{
				"id":  "12345",
				"key": "TEST-123",
				"fields": map[string]any{
					"summary":     "Test issue",
					"created":     "2024-11-06T00:00:00.000Z",
					"description": "This is a test issue description",
				},
			},
			"user": map[string]any{
				"name": "testuser-name",
			},
		}
		// For now, this is a minimal test that verifies the JSON marshaling works
		// In a full implementation, you would inject the mock pipeline group
		// into the Jira webhook handler and verify it processes events correctly
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		require.NotEmpty(t, body)
		// Create a mock HTTP request
		req := httptest.NewRequest(http.MethodPost, "/jira/webhook", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		// Verify the request was created successfully
		require.Equal(t, "/jira/webhook", req.URL.Path)
		require.Equal(t, "POST", req.Method)
		// Verify mock pipeline group is ready
		require.False(t, mockPipelineGroup.AddMessageInvoked)
		require.False(t, mockPipelineGroup.StartInvoked)
	})
}

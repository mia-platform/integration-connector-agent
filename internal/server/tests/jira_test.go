// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0
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

//go:build integration
// +build integration

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	"github.com/stretchr/testify/require"
)

func TestJiraIntegration(t *testing.T) {
	app, mongoURL, db := setupApp(t, setupServerConfig{
		configPath: "testdata/jira/config.json",
	})
	defer app.Shutdown()

	t.Run("save data on mongo", func(t *testing.T) {
		collJiraIssues := testutils.MongoCollection(t, mongoURL, "jira-issues", db)
		collJira := testutils.MongoCollection(t, mongoURL, "jira", db)

		events := []struct {
			name    string
			reqBody map[string]any

			expectedIssuesResults   []map[string]any
			expectedCollJiraResults []map[string]any
		}{
			{
				name: "create issue 1",
				reqBody: map[string]any{
					"webhookEvent": "jira:issue_created",
					"id":           123,
					"timestamp":    time.Now().UnixMilli(),
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
				},

				expectedIssuesResults: []map[string]any{
					{
						"_eventId":    "12345",
						"key":         "TEST-123",
						"createdAt":   "2024-11-06T00:00:00.000Z",
						"description": "This is a test issue description",
						"summary":     "Test issue",
					},
				},
				expectedCollJiraResults: []map[string]any{
					{
						"_eventId":     "12345",
						"webhookEvent": "jira:issue_created",
						"id":           float64(123),
						"timestamp":    float64(time.Now().UnixMilli()),
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
					},
				},
			},
			{
				name: "update issue 1",
				reqBody: map[string]any{
					"webhookEvent": "jira:issue_updated",
					"id":           124,
					"timestamp":    time.Now().UnixMilli(),
					"issue": map[string]any{
						"id":  "12345",
						"key": "TEST-123",
						"fields": map[string]any{
							"summary":     "Test modified issue",
							"created":     "2024-11-06T00:00:00.000Z",
							"description": "This is a test issue description modified",
						},
					},
					"user": map[string]any{
						"name": "testuser-name",
					},
				},

				expectedIssuesResults: []map[string]any{
					{
						"_eventId":    "12345",
						"key":         "TEST-123",
						"createdAt":   "2024-11-06T00:00:00.000Z",
						"description": "This is a test issue description modified",
						"summary":     "Test modified issue",
					},
				},
				expectedCollJiraResults: []map[string]any{
					{
						"_eventId":     "12345",
						"webhookEvent": "jira:issue_updated",
						"id":           float64(124),
						"timestamp":    float64(time.Now().UnixMilli()),
						"issue": map[string]any{
							"id":  "12345",
							"key": "TEST-123",
							"fields": map[string]any{
								"summary":     "Test modified issue",
								"created":     "2024-11-06T00:00:00.000Z",
								"description": "This is a test issue description modified",
							},
						},
						"user": map[string]any{
							"name": "testuser-name",
						},
					},
				},
			},
			{
				name: "create issue 2",
				reqBody: map[string]any{
					"webhookEvent": "jira:issue_created",
					"id":           125,
					"timestamp":    time.Now().UnixMilli(),
					"issue": map[string]any{
						"id":  "12346",
						"key": "TEST-456",
						"fields": map[string]any{
							"summary":     "Test second issue",
							"created":     "2024-11-10T00:00:00.000Z",
							"description": "This is the second issue",
						},
					},
					"user": map[string]any{
						"name": "testuser-name",
					},
				},

				expectedIssuesResults: []map[string]any{
					{
						"_eventId":    "12345",
						"key":         "TEST-123",
						"createdAt":   "2024-11-06T00:00:00.000Z",
						"description": "This is a test issue description modified",
						"summary":     "Test modified issue",
					},
					{
						"_eventId":    "12346",
						"key":         "TEST-456",
						"createdAt":   "2024-11-10T00:00:00.000Z",
						"description": "This is the second issue",
						"summary":     "Test second issue",
					},
				},
				expectedCollJiraResults: []map[string]any{
					{
						"_eventId":     "12345",
						"webhookEvent": "jira:issue_updated",
						"id":           float64(124),
						"timestamp":    float64(time.Now().UnixMilli()),
						"issue": map[string]any{
							"id":  "12345",
							"key": "TEST-123",
							"fields": map[string]any{
								"summary":     "Test modified issue",
								"created":     "2024-11-06T00:00:00.000Z",
								"description": "This is a test issue description modified",
							},
						},
						"user": map[string]any{
							"name": "testuser-name",
						},
					},
					{
						"_eventId":     "12346",
						"webhookEvent": "jira:issue_created",
						"id":           float64(125),
						"timestamp":    float64(time.Now().UnixMilli()),
						"issue": map[string]any{
							"id":  "12346",
							"key": "TEST-456",
							"fields": map[string]any{
								"summary":     "Test second issue",
								"created":     "2024-11-10T00:00:00.000Z",
								"description": "This is the second issue",
							},
						},
						"user": map[string]any{
							"name": "testuser-name",
						},
					},
				},
			},
			{
				name: "delete issue 1",
				reqBody: map[string]any{
					"webhookEvent": "jira:issue_deleted",
					"id":           125,
					"timestamp":    time.Now().UnixMilli(),
					"issue": map[string]any{
						"id":  "12346",
						"key": "TEST-123",
					},
					"user": map[string]any{
						"name": "testuser-name",
					},
				},

				expectedIssuesResults: []map[string]any{
					{
						"_eventId":    "12345",
						"key":         "TEST-123",
						"createdAt":   "2024-11-06T00:00:00.000Z",
						"description": "This is a test issue description modified",
						"summary":     "Test modified issue",
					},
				},
				expectedCollJiraResults: []map[string]any{
					{
						"_eventId":     "12345",
						"webhookEvent": "jira:issue_updated",
						"id":           float64(124),
						"timestamp":    float64(time.Now().UnixMilli()),
						"issue": map[string]any{
							"id":  "12345",
							"key": "TEST-123",
							"fields": map[string]any{
								"summary":     "Test modified issue",
								"created":     "2024-11-06T00:00:00.000Z",
								"description": "This is a test issue description modified",
							},
						},
						"user": map[string]any{
							"name": "testuser-name",
						},
					},
				},
			},
			{
				name: "another event type is saved only on jira collection",
				reqBody: map[string]any{
					"webhookEvent": "issuelink_created",
					"issueLink": map[string]any{
						"id":                 876,
						"sourceIssueId":      222,
						"destinationIssueId": 333,
						"issueLinkType": map[string]any{
							"id":                111,
							"name":              "Link name",
							"outwardName":       "executes Test",
							"inwardName":        "is executed by",
							"isSubTaskLinkType": false,
							"isSystemLinkType":  false,
						},
						"systemLink": false,
					},
				},

				expectedIssuesResults: []map[string]any{
					{
						"_eventId":    "12345",
						"key":         "TEST-123",
						"createdAt":   "2024-11-06T00:00:00.000Z",
						"description": "This is a test issue description modified",
						"summary":     "Test modified issue",
					},
				},
				expectedCollJiraResults: []map[string]any{
					{
						"_eventId":     "12345",
						"webhookEvent": "jira:issue_updated",
						"id":           float64(124),
						"timestamp":    float64(time.Now().UnixMilli()),
						"issue": map[string]any{
							"id":  "12345",
							"key": "TEST-123",
							"fields": map[string]any{
								"summary":     "Test modified issue",
								"created":     "2024-11-06T00:00:00.000Z",
								"description": "This is a test issue description modified",
							},
						},
						"user": map[string]any{
							"name": "testuser-name",
						},
					},
					{
						"_eventId":     "876",
						"webhookEvent": "issuelink_created",
						"issueLink": map[string]any{
							"id":                 float64(876),
							"sourceIssueId":      float64(222),
							"destinationIssueId": float64(333),
							"issueLinkType": map[string]any{
								"id":                float64(111),
								"name":              "Link name",
								"outwardName":       "executes Test",
								"inwardName":        "is executed by",
								"isSubTaskLinkType": false,
								"isSystemLinkType":  false,
							},
							"systemLink": false,
						},
					},
				},
			},
		}

		for _, tc := range events {
			t.Log(tc.name)
			body, err := json.Marshal(tc.reqBody)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/jira/webhook", bytes.NewBuffer(body))

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)

			findAllDocuments(t, collJiraIssues, tc.expectedIssuesResults)
			findAllDocuments(t, collJira, tc.expectedCollJiraResults)
		}
	})
}

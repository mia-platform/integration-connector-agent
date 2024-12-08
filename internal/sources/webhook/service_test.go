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

package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	fakesink "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestSetupServiceWithConfig(t *testing.T) {
	logger, _ := test.NewNullLogger()
	defaultWebhookEndpoint := "/webhook-path"

	type testItem struct {
		config *Configuration
		req    func(t *testing.T) *http.Request

		expectedStatusCode int
		expectedBody       func(t *testing.T, body io.ReadCloser)
	}
	tests := map[string]testItem{
		"expose the correct API - empty body": {
			config: &Configuration{
				WebhookPath: defaultWebhookEndpoint,
				Events:      &Events{},
			},
			req: func(t *testing.T) *http.Request {
				t.Helper()
				return httptest.NewRequest(http.MethodPost, defaultWebhookEndpoint, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		"fails validation": {
			config: &Configuration{
				WebhookPath: defaultWebhookEndpoint,
				Authentication: HMAC{
					Secret:     "SECRET",
					HeaderName: "X-Hub-Signature",
				},
				Events: &Events{},
			},
			req: func(t *testing.T) *http.Request {
				t.Helper()
				return httptest.NewRequest(http.MethodPost, defaultWebhookEndpoint, nil)
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body io.ReadCloser) {
				t.Helper()
				expectedBody := utils.HTTPError{}
				require.NoError(t, json.NewDecoder(body).Decode(&expectedBody))
				require.Equal(t, utils.HTTPError{
					Error:   "Validation Error",
					Message: NoSignatureHeaderButSecretError,
				}, expectedBody)
			},
		},
		"expose the correct default path API": {
			config: &Configuration{
				WebhookPath: defaultWebhookEndpoint,
				Events: &Events{
					Supported: map[string]Event{
						"jira:issue_updated": {
							FieldID:   "issue.id",
							Operation: entities.Write,
						},
					},
					EventTypeFieldPath: "webhookEvent",
				},
			},
			req: func(t *testing.T) *http.Request {
				t.Helper()

				jiraIssue := map[string]any{
					"id": 1,
					"issue": map[string]any{
						"id":  "1",
						"key": "ISSUE-KEY",
					},
					"webhookEvent": "jira:issue_updated",
				}
				reqBody, err := json.Marshal(jiraIssue)
				require.NoError(t, err)

				return httptest.NewRequest(http.MethodPost, defaultWebhookEndpoint, bytes.NewBuffer(reqBody))
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			app, router := testutils.GetTestRouter(t)

			proc := &processors.Processors{}
			s := fakesink.New(nil)
			p1, err := pipeline.New(logger, proc, s)
			require.NoError(t, err)

			pg := pipeline.NewGroup(logger, p1)

			err = SetupService(context.TODO(), router, test.config, pg)
			require.NoError(t, err)

			res, err := app.Test(test.req(t))
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, test.expectedStatusCode, res.StatusCode)

			if test.expectedBody != nil {
				test.expectedBody(t, res.Body)
			}
		})
	}
}

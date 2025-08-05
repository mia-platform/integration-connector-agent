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
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	fakesink "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupServiceWithConfig(t *testing.T) {
	t.Parallel()
	logger, _ := test.NewNullLogger()
	defaultWebhookEndpoint := "/webhook-path"

	type testItem struct {
		config             Configuration[*fakeAuthentication]
		req                func(t *testing.T) *http.Request
		expectedStatusCode int
		expectedBody       func(t *testing.T, body io.ReadCloser)
	}

	testCases := map[string]testItem{
		"expose the correct API - empty body": {
			config: Configuration[*fakeAuthentication]{
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
			config: Configuration[*fakeAuthentication]{
				WebhookPath:    defaultWebhookEndpoint,
				Authentication: &fakeAuthentication{checkErr: errors.New("failed to authenticate")},
				Events:         &Events{},
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
				assert.Equal(t, utils.HTTPError{
					Error:   "Validation Error",
					Message: "failed to authenticate",
				}, expectedBody)
			},
		},
		"expose the correct default path API": {
			config: Configuration[*fakeAuthentication]{
				WebhookPath: defaultWebhookEndpoint,
				Events: &Events{
					Supported: map[string]Event{
						"jira:issue_updated": {
							GetFieldID: GetPrimaryKeyByPath("issue.id"),
							Operation:  entities.Write,
						},
					},
					GetEventType: GetEventTypeByPath("webhookEvent"),
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
		"handle x-www-form-urlencoded payload": {
			config: Configuration[*fakeAuthentication]{
				WebhookPath: defaultWebhookEndpoint,
				Events: &Events{
					Supported: map[string]Event{
						"jira:issue_updated": {
							GetFieldID: GetPrimaryKeyByPath("issue.id"),
							Operation:  entities.Write,
						},
					},
					GetEventType: GetEventTypeByPath("webhookEvent"),
					PayloadKey: ContentTypeConfig{
						fiber.MIMEApplicationForm: "payload",
					},
				},
			},
			req: func(t *testing.T) *http.Request {
				t.Helper()

				form := url.Values{}
				form.Set("payload", `{"webhookEvent":"jira:issue_updated","issue":{"id":"1","key":"ISSUE-KEY"}}`)
				form.Encode()

				req := httptest.NewRequest(http.MethodPost, defaultWebhookEndpoint, strings.NewReader(form.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				return req
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			app, router := testutils.GetTestRouter(t)

			proc := &processors.Processors{}
			s := fakesink.New(nil)
			p1, err := pipeline.New(logger, proc, s)
			require.NoError(t, err)

			pg := pipeline.NewGroup(logger, p1)

			err = SetupService(t.Context(), router, test.config, pg)
			require.NoError(t, err)

			res, err := app.Test(test.req(t))
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, test.expectedStatusCode, res.StatusCode)

			if test.expectedBody != nil {
				test.expectedBody(t, res.Body)
			}
		})
	}
}

func TestExtractBodyFromContentType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		contentType string
		body        string
		events      *Events

		expectedBody  []byte
		expectedError error
	}{
		"JSON content type with valid JSON body": {
			contentType:   "application/json",
			body:          `{"key": "value", "number": 123}`,
			expectedBody:  []byte(`{"key": "value", "number": 123}`),
			expectedError: nil,
		},
		"JSON content type with charset parameter": {
			contentType:   "application/json; charset=utf-8",
			body:          `{"test": true}`,
			expectedBody:  []byte(`{"test": true}`),
			expectedError: nil,
		},
		"JSON content type with trailing semicolon": {
			contentType:   "application/json;",
			body:          `{"test": true}`,
			expectedBody:  []byte(`{"test": true}`),
			expectedError: nil,
		},
		"Form urlencoded content type with payload with json content": {
			contentType: "application/x-www-form-urlencoded",
			body:        "payload=%7B%22key1%22%3A%22value1%22%2C%22key2%22%3A%22value2%22%7D",
			events: &Events{
				PayloadKey: ContentTypeConfig{
					fiber.MIMEApplicationForm: "payload",
				},
			},
			expectedBody:  []byte(`{"key1":"value1","key2":"value2"}`),
			expectedError: nil,
		},
		"Unknown content type raise error": {
			contentType: "text/plain",
			body:        "key=value",
			expectedBody: func() []byte {
				expected := map[string]any{
					"key": "value",
				}
				b, err := json.Marshal(expected)
				require.NoError(t, err)
				return b
			}(),
			expectedError: ErrUnsupportedContentType,
		},
		"Invalid content type header": {
			contentType:   "invalid/content-type-header-with-bad-chars-<>",
			body:          "test",
			expectedBody:  nil,
			expectedError: ErrFailedToParseContentType,
		},
		"empty content type returns body as is": {
			contentType:  "",
			body:         `{}`,
			expectedBody: []byte(`{}`),
		},
		"Empty body with JSON content type": {
			contentType:   "application/json",
			body:          "",
			expectedBody:  []byte(""),
			expectedError: nil,
		},
		"JSON content type with empty object": {
			contentType:   "application/json",
			body:          `{}`,
			expectedBody:  []byte(`{}`),
			expectedError: nil,
		},
		"JSON content type with array": {
			contentType:   "application/json",
			body:          `[1,2,3]`,
			expectedBody:  []byte(`[1,2,3]`),
			expectedError: nil,
		},
		"Malformed form body": {
			contentType: "application/x-www-form-urlencoded",
			events: &Events{
				PayloadKey: ContentTypeConfig{
					fiber.MIMEApplicationForm: "payload",
				},
			},
			body:         `malformed data %`,
			expectedBody: nil,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			app := fiber.New()

			var actualBody []byte
			var actualError error

			var events = &Events{}
			if test.events != nil {
				events = test.events
			}

			app.Post("/test", func(c *fiber.Ctx) error {
				actualBody, actualError = extractBodyFromContentType(c, events)
				return c.SendStatus(200)
			})

			var reqBody io.Reader
			if test.body != "" {
				reqBody = strings.NewReader(test.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/test", reqBody)
			if test.contentType != "" {
				req.Header.Set("Content-Type", test.contentType)
			}

			res, err := app.Test(req)
			require.NoError(t, err)
			defer res.Body.Close()

			if test.expectedError != nil {
				require.Error(t, actualError)
				assert.ErrorIs(t, actualError, test.expectedError)
				assert.Nil(t, actualBody)
			} else {
				assert.NoError(t, actualError)
				if string(test.expectedBody) != "" {
					assert.JSONEq(t, string(test.expectedBody), string(actualBody))
				}
			}
		})
	}
}

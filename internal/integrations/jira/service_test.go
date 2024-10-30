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

package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	swagger "github.com/davidebianchi/gswagger"
	oasfiber "github.com/davidebianchi/gswagger/support/fiber"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/mia-platform/data-connector-agent/internal/entities"
	"github.com/mia-platform/data-connector-agent/internal/utils"
	"github.com/mia-platform/data-connector-agent/internal/writer"
	fakewriter "github.com/mia-platform/data-connector-agent/internal/writer/fake"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestSetupServiceWithConfig(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := logrus.NewEntry(log)

	type testItem struct {
		config Configuration
		req    func(t *testing.T) *http.Request
		writer writer.Writer[entities.PipelineEvent]

		expectedStatusCode int
		expectedBody       func(t *testing.T, body io.ReadCloser)
	}
	tests := map[string]testItem{
		"expose the correct API - empty body": {
			req: func(t *testing.T) *http.Request {
				t.Helper()
				return httptest.NewRequest(http.MethodPost, webhookEndpoint, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		"fails validation": {
			req: func(t *testing.T) *http.Request {
				t.Helper()
				return httptest.NewRequest(http.MethodPost, webhookEndpoint, nil)
			},
			config: Configuration{
				Secret: "SECRET",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body io.ReadCloser) {
				t.Helper()
				expectedBody := utils.HTTPError{}
				require.NoError(t, json.NewDecoder(body).Decode(&expectedBody))
				require.Equal(t, utils.HTTPError{
					Error:   "Validation Error",
					Message: noSignatureHeaderButSecretError,
				}, expectedBody)
			},
		},
		"expose the correct API - updated issue": {
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

				return httptest.NewRequest(http.MethodPost, webhookEndpoint, bytes.NewBuffer(reqBody))
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			app, router := getRouter(t)

			if test.writer == nil {
				test.writer = fakewriter.New()
			}

			err := SetupService(context.TODO(), logger, router, test.config, test.writer)
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

func getRouter(t *testing.T) (*fiber.App, *swagger.Router[fiber.Handler, fiber.Router]) {
	t.Helper()

	app := fiber.New()
	router, err := swagger.NewRouter(oasfiber.NewRouter(app), swagger.Options{
		Openapi: &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test",
				Version: "test-version",
			},
		},
	})
	require.NoError(t, err)

	return app, router
}

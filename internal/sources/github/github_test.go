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

package github

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	fakewriter "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	testCases := map[string]struct {
		config *Config

		expectedConfig *Config
		expectedError  error
	}{
		"with default": {
			config: &Config{},
			expectedConfig: &Config{
				WebhookPath: defaultWebhookPath,
				Authentication: webhook.HMAC{
					HeaderName: authHeaderName,
				},
			},
		},
		"with custom values": {
			config: &Config{
				WebhookPath: "/custom/webhook",
				Authentication: webhook.HMAC{
					HeaderName: "X-Custom-Header",
					Secret:     config.SecretSource("secret"),
				},
			},
			expectedConfig: &Config{
				WebhookPath: "/custom/webhook",
				Authentication: webhook.HMAC{
					HeaderName: "X-Custom-Header",
					Secret:     config.SecretSource("secret"),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectedError != nil {
				require.EqualError(t, err, tc.expectedError.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedConfig, tc.config)
		})
	}

	t.Run("unmarshal config", func(t *testing.T) {
		rawConfig, err := os.ReadFile("testdata/config.json")
		require.NoError(t, err)

		actual := &Config{}
		require.NoError(t, json.Unmarshal(rawConfig, actual))
		require.NoError(t, actual.Validate())

		require.Equal(t, &Config{
			WebhookPath: "/webhook",
			Authentication: webhook.HMAC{
				HeaderName: authHeaderName,
				Secret:     config.SecretSource("SECRET_VALUE"),
			},
		}, actual)
	})
}

func TestAddSourceToRouter(t *testing.T) {
	logger, _ := test.NewNullLogger()

	t.Run("setup webhook", func(t *testing.T) {
		ctx := t.Context()

		rawConfig, err := os.ReadFile("testdata/config.json")
		require.NoError(t, err)
		cfg := config.GenericConfig{}
		require.NoError(t, json.Unmarshal(rawConfig, &cfg))

		app, router := testutils.GetTestRouter(t)

		proc := &processors.Processors{}
		s := fakewriter.New(nil)
		p1, err := pipeline.New(logger, proc, s)
		require.NoError(t, err)

		pg := pipeline.NewGroup(logger, p1)

		id := "12345"
		err = AddSourceToRouter(ctx, cfg, pg, router)
		require.NoError(t, err)

		testCases := []struct {
			eventName   string
			body        string
			contentType string

			expectedPk        entities.PkFields
			expectedOperation entities.Operation
		}{
			{
				eventName:   pullRequestEvent,
				body:        getPullRequestPayload("opened", id),
				contentType: "application/json",

				expectedPk:        entities.PkFields{{Key: "pull_request.id", Value: id}},
				expectedOperation: entities.Write,
			},
			{
				eventName:   pullRequestEvent,
				body:        getPullRequestPayload("closed", id),
				contentType: "application/json",

				expectedPk:        entities.PkFields{{Key: "pull_request.id", Value: id}},
				expectedOperation: entities.Write,
			},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("invoke webhook with %s event", tc.eventName), func(t *testing.T) {
				defer s.ResetCalls()

				req := getWebhookRequest(bytes.NewBufferString(tc.body), tc.eventName)

				resp, err := app.Test(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
				require.Eventually(t, func() bool {
					return len(s.Calls()) == 1
				}, 1*time.Second, 10*time.Millisecond)
				require.Equal(t, fakewriter.Call{
					Operation: tc.expectedOperation,
					Data: &entities.Event{
						PrimaryKeys:   tc.expectedPk,
						Type:          tc.eventName,
						OperationType: tc.expectedOperation,
						OriginalRaw:   []byte(tc.body),
					},
				}, s.Calls().LastCall())
			})

			t.Run(fmt.Sprintf("invoke webhook with %s event with form body", tc.eventName), func(t *testing.T) {
				defer s.ResetCalls()

				form := url.Values{}
				form.Set("payload", tc.body)
				req := getWebhookRequest(bytes.NewBufferString(form.Encode()), tc.eventName)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

				resp, err := app.Test(req)
				require.NoError(t, err)
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
				require.Eventually(t, func() bool {
					return len(s.Calls()) == 1
				}, 1*time.Second, 10*time.Millisecond)
				require.Equal(t, fakewriter.Call{
					Operation: tc.expectedOperation,
					Data: &entities.Event{
						PrimaryKeys:   tc.expectedPk,
						Type:          tc.eventName,
						OperationType: tc.expectedOperation,
						OriginalRaw:   []byte(tc.body),
					},
				}, s.Calls().LastCall())
			})
		}
	})
}

func getWebhookRequest(body *bytes.Buffer, eventType string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	hmac := getHMACValidationHeader("SECRET_VALUE", body.Bytes())
	req.Header.Add(authHeaderName, fmt.Sprintf("sha256=%s", hmac))
	req.Header.Add(githubEventHeader, eventType)
	return req
}

func getHMACValidationHeader(secret string, body []byte) string {
	hasher := hmac.New(sha256.New, []byte(secret))
	hasher.Write(body)
	return hex.EncodeToString(hasher.Sum(nil))
}

func getPullRequestPayload(action, id string) string {
	return fmt.Sprintf(`{
		"action": "%s",
		"pull_request": {
			"id": "%s",
			"title": "Test PR",
			"user": {
				"login": "octocat"
			}
		}
	}`, action, id)
}

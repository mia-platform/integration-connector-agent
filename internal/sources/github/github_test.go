// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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
	webhookhmac "github.com/mia-platform/integration-connector-agent/internal/sources/webhook/hmac"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config         *Config
		expectedConfig *Config
		expectedError  error
	}{
		"with default": {
			config: &Config{
				Configuration: webhook.Configuration[webhookhmac.Authentication]{
					Authentication: webhookhmac.Authentication{
						Secret: "secret",
					},
				},
			},
			expectedConfig: &Config{
				Configuration: webhook.Configuration[webhookhmac.Authentication]{
					WebhookPath: defaultWebhookPath,
					Authentication: webhookhmac.Authentication{
						HeaderName: authHeaderName,
						Secret:     "secret",
					},
					Events: SupportedEvents,
				},
			},
		},
		"with custom values": {
			config: &Config{
				Configuration: webhook.Configuration[webhookhmac.Authentication]{
					WebhookPath: "/custom/webhook",
					Authentication: webhookhmac.Authentication{
						HeaderName: "X-Custom-Header",
						Secret:     "secret",
					},
				},
			},
			expectedConfig: &Config{
				Configuration: webhook.Configuration[webhookhmac.Authentication]{
					WebhookPath: "/custom/webhook",
					Authentication: webhookhmac.Authentication{
						HeaderName: "X-Custom-Header",
						Secret:     "secret",
					},
					Events: SupportedEvents,
				},
			},
		},
		"unmarshal from file": {
			config: func() *Config {
				rawConfig, err := os.ReadFile("testdata/config.json")
				require.NoError(t, err)

				actual := &Config{}
				require.NoError(t, json.Unmarshal(rawConfig, actual))
				return actual
			}(),
			expectedConfig: &Config{
				Configuration: webhook.Configuration[webhookhmac.Authentication]{
					WebhookPath: "/webhook",
					Authentication: webhookhmac.Authentication{
						HeaderName: authHeaderName,
						Secret:     "SECRET_VALUE",
					},
					Events: SupportedEvents,
				},
			},
		},
		"empty config return default": {
			config: &Config{},
			expectedConfig: &Config{
				Configuration: webhook.Configuration[webhookhmac.Authentication]{
					WebhookPath: defaultWebhookPath,
					Authentication: webhookhmac.Authentication{
						HeaderName: authHeaderName,
					},
					Events: SupportedEvents,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tc.config.Validate()
			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expectedConfig, tc.config)
		})
	}
}

func TestAddSourceToRouter(t *testing.T) {
	logger, _ := test.NewNullLogger()

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
					OriginalRaw:   getExpectedPayloadWithEventType(tc.body, tc.eventName),
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
					OriginalRaw:   getExpectedPayloadWithEventType(tc.body, tc.eventName),
				},
			}, s.Calls().LastCall())
		})
	}
}

func getWebhookRequest(body *bytes.Buffer, eventType string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	hmac := getHMACValidationHeader("SECRET_VALUE", body.Bytes())
	req.Header.Add(authHeaderName, "sha256="+hmac)
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

// getExpectedPayloadWithEventType injects eventType into a JSON payload to match what the webhook processor creates
func getExpectedPayloadWithEventType(originalPayload, eventType string) []byte {
	// Parse the original payload
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(originalPayload), &jsonData); err != nil {
		panic(err) // This should never happen in tests
	}

	// Add eventType field
	jsonData["eventType"] = eventType

	// Marshal back to bytes
	enhanced, err := json.Marshal(jsonData)
	if err != nil {
		panic(err) // This should never happen in tests
	}

	return enhanced
}

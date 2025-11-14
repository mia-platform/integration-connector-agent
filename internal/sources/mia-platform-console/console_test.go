// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package console

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
		config *Config

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
						Secret: "secret",
					},
				},
			},
			expectedConfig: &Config{
				Configuration: webhook.Configuration[webhookhmac.Authentication]{
					WebhookPath: "/custom/webhook",
					Authentication: webhookhmac.Authentication{
						HeaderName: authHeaderName,
						Secret:     "secret",
					},
					Events: SupportedEvents,
				},
			},
		},
		"unmarshaled from file": {
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
				assert.NoError(t, err)
			}
			assert.NotNil(t, tc.config.Authentication.CustomValidator)
			tc.config.Authentication.CustomValidator = nil
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
	s := fakewriter.New(nil, logger)
	p1, err := pipeline.New(logger, proc, s)
	require.NoError(t, err)

	pg := pipeline.NewGroup(logger, p1)

	err = AddSourceToRouter(ctx, cfg, pg, router)
	require.NoError(t, err)

	tenantID := "my-tenant-id"
	projectID := "my-prj-id"

	testCases := []struct {
		eventName string
		body      string

		expectedPk        entities.PkFields
		expectedOperation entities.Operation
	}{
		{
			eventName: projectCreatedEvent,
			body:      fmt.Sprintf(`{"eventName":"%s","payload":{"projectId":"%s","tenantId":"%s"}}`, projectCreatedEvent, projectID, tenantID),

			expectedPk: entities.PkFields{
				entities.PkField{Key: "tenantId", Value: tenantID},
				entities.PkField{Key: "projectId", Value: projectID},
			},
		},
		{
			eventName: serviceCreatedEvent,
			body:      fmt.Sprintf(`{"eventName":"%s","payload":{"projectId":"%s","tenantId":"%s","serviceName":"my-service"}}`, serviceCreatedEvent, projectID, tenantID),

			expectedPk: entities.PkFields{
				entities.PkField{Key: "tenantId", Value: tenantID},
				entities.PkField{Key: "projectId", Value: projectID},
				entities.PkField{Key: "serviceName", Value: "my-service"},
			},
		},
		{
			eventName: configurationSavedEvent,
			body:      fmt.Sprintf(`{"eventName":"%s","payload":{"tenantId":"%s","projectId":"%s","revisionName":"my-revision"}}`, configurationSavedEvent, tenantID, projectID),

			expectedPk: entities.PkFields{
				entities.PkField{Key: "tenantId", Value: tenantID},
				entities.PkField{Key: "projectId", Value: projectID},
				entities.PkField{Key: "revisionName", Value: "my-revision"},
			},
		},
		{
			eventName: tagCreatedEvent,
			body:      fmt.Sprintf(`{"eventName":"%s","payload":{"projectId":"%s","tenantId":"%s","tagName":"my-tag"}}`, tagCreatedEvent, projectID, tenantID),

			expectedPk: entities.PkFields{
				entities.PkField{Key: "tenantId", Value: tenantID},
				entities.PkField{Key: "projectId", Value: projectID},
				entities.PkField{Key: "tagName", Value: "my-tag"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("invoke webhook with %s event", tc.eventName), func(t *testing.T) {
			defer s.ResetCalls()

			req := getWebhookRequest(bytes.NewBufferString(tc.body))

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
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
}

func getWebhookRequest(body *bytes.Buffer) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	hmac := getHMACValidationHeader("SECRET_VALUE", body.Bytes())
	req.Header.Add(authHeaderName, "sha256="+hmac)
	return req
}

func getHMACValidationHeader(secret string, body []byte) string {
	hasher := sha256.New()
	hasher.Write(body)
	hasher.Write([]byte(secret))
	return hex.EncodeToString(hasher.Sum(nil))
}

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
	"context"
	"encoding/json"
	"os"
	"testing"

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
					HeaderName: defaultAuthHeaderName,
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
				HeaderName: defaultAuthHeaderName,
				Secret:     config.SecretSource("SECRET_VALUE"),
			},
		}, actual)
	})
}

func TestGetWebhookConfig(t *testing.T) {
	testCases := map[string]struct {
		config *Config

		expectedConfig *webhook.Configuration
		expectedError  string
	}{
		"valid config without authentication": {
			config: &Config{
				WebhookPath: "/webhook",
			},
			expectedConfig: &webhook.Configuration{
				WebhookPath:    "/webhook",
				Authentication: webhook.HMAC{},
				Events:         &DefaultSupportedEvents,
			},
		},
		"valid config with authentication": {
			config: &Config{
				WebhookPath: "/webhook",
				Authentication: webhook.HMAC{
					HeaderName: "X-Custom-Header",
					Secret:     config.SecretSource("secret"),
				},
			},
			expectedConfig: &webhook.Configuration{
				WebhookPath: "/webhook",
				Authentication: webhook.HMAC{
					HeaderName: "X-Custom-Header",
					Secret:     config.SecretSource("secret"),
				},
				Events: &DefaultSupportedEvents,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			webhookConfig, err := tc.config.getWebhookConfig()
			require.NoError(t, err)

			if tc.expectedError != "" {
				require.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedConfig, webhookConfig)
		})
	}
}

func TestAddSourceToRouter(t *testing.T) {
	logger, _ := test.NewNullLogger()

	t.Run("setup webhook", func(t *testing.T) {
		ctx := context.Background()

		rawConfig, err := os.ReadFile("testdata/config.json")
		require.NoError(t, err)
		cfg := config.GenericConfig{}
		require.NoError(t, json.Unmarshal(rawConfig, &cfg))

		_, router := testutils.GetTestRouter(t)

		proc := &processors.Processors{}
		s := fakewriter.New(nil)
		p1, err := pipeline.New(logger, proc, s)
		require.NoError(t, err)

		pg := pipeline.NewGroup(logger, p1)

		err = AddSourceToRouter(ctx, cfg, pg, router)
		require.NoError(t, err)
	})
}

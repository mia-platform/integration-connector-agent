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

package vm

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	testCases := map[string]struct {
		config *Config

		expectedConfig *Config
		expectedError  error
	}{
		"with default": {
			config: &Config{},
			expectedConfig: &Config{
				Configuration: webhook.Configuration{
					WebhookPath: defaultWebhookPath,
					Authentication: webhook.Authentication{
						HeaderName: defaultAuthHeaderName,
					},
					Events: &DefaultSupportedEvents,
				},
			},
		},
		"with custom values": {
			config: &Config{
				Configuration: webhook.Configuration{
					WebhookPath: "/custom/webhook",
					Authentication: webhook.Authentication{
						HeaderName: "X-Custom-Header",
						Secret:     config.SecretSource("secret"),
					},
					Events: &webhook.Events{
						EventTypeFieldPath: "customEventPath",
						Supported: map[string]webhook.Event{
							"event1": {
								Operation: entities.Write,
								FieldID:   "id",
							},
						},
					},
				},
			},
			expectedConfig: &Config{
				Configuration: webhook.Configuration{
					WebhookPath: "/custom/webhook",
					Authentication: webhook.Authentication{
						HeaderName: "X-Custom-Header",
						Secret:     config.SecretSource("secret"),
					},
					Events: &webhook.Events{
						EventTypeFieldPath: "customEventPath",
						Supported: map[string]webhook.Event{
							"event1": {
								Operation: entities.Write,
								FieldID:   "id",
							},
						},
					},
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
			Configuration: webhook.Configuration{
				WebhookPath: "/vm/webhook",
				Authentication: webhook.Authentication{
					HeaderName: defaultAuthHeaderName,
					Secret:     config.SecretSource("SECRET_VALUE"),
				},
				Events: &DefaultSupportedEvents,
			},
		}, actual)
	})
}

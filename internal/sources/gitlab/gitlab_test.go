// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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

package gitlab

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook/hmac"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config      *Config
		expectedErr string
	}{
		"valid webhook only config": {
			config: &Config{},
		},
		"valid import config": {
			config: &Config{
				ImportWebhookPath: "/gitlab/import",
				ImportAuthentication: hmac.Authentication{
					Secret:     config.SecretSource("test-secret"),
					HeaderName: "X-Gitlab-Token",
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.config.Validate()
			if testCase.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestConfigWithDefault(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	cfg = cfg.withDefault()

	assert.Equal(t, defaultWebhookPath, cfg.WebhookPath)
	assert.Equal(t, authHeaderName, cfg.Authentication.HeaderName)
	assert.Equal(t, SupportedEvents, cfg.Events)
	assert.Equal(t, "https://gitlab.com", cfg.BaseURL)
}

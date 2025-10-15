// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

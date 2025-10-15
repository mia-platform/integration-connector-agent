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
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGitHubSource(t *testing.T) {
	t.Run("create GitHub source without import webhook", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"type":        "github",
			"webhookPath": "/github/webhook",
		}
		rawData, err := json.Marshal(rawConfig)
		require.NoError(t, err)

		cfg := config.GenericConfig{
			Type: "github",
			Raw:  rawData,
		}

		log := logrus.New()
		_, router := testutils.GetTestRouter(t)
		pg := &pipeline.PipelineGroupMock{}

		source, err := NewGitHubSource(t.Context(), log, cfg, pg, router)
		require.NoError(t, err)
		require.NotNil(t, source)

		err = source.Close()
		assert.NoError(t, err)
	})

	t.Run("create GitHub source with import webhook but missing configuration", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"type":              "github",
			"webhookPath":       "/github/webhook",
			"importWebhookPath": "/github/import",
		}
		rawData, err := json.Marshal(rawConfig)
		require.NoError(t, err)

		cfg := config.GenericConfig{
			Type: "github",
			Raw:  rawData,
		}

		log := logrus.New()
		_, router := testutils.GetTestRouter(t)
		pg := &pipeline.PipelineGroupMock{}

		_, err = NewGitHubSource(t.Context(), log, cfg, pg, router)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "GitHub authentication is required for import functionality")
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid configuration without import", func(t *testing.T) {
		cfg := &Config{}
		cfg.WebhookPath = "/github/webhook"
		cfg.Events = SupportedEvents

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid configuration with import", func(t *testing.T) {
		cfg := &Config{
			ImportWebhookPath: "/github/import",
		}
		cfg.WebhookPath = "/github/webhook"
		cfg.Events = SupportedEvents
		cfg.ImportAuthentication.Secret = config.SecretSource("test-secret")
		cfg.ImportAuthentication.HeaderName = "X-Hub-Signature-256"

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid configuration - import webhook without authentication", func(t *testing.T) {
		cfg := &Config{
			ImportWebhookPath: "/github/import",
		}
		cfg.WebhookPath = "/github/webhook"
		cfg.Events = SupportedEvents
		// Set secret but no header name to trigger validation error
		cfg.ImportAuthentication.Secret = config.SecretSource("test-secret")
		// HeaderName is intentionally omitted

		err := cfg.Validate()
		assert.Error(t, err)
	})
}

func TestGitHubEventBuilder(t *testing.T) {
	t.Run("create pipeline event from import data", func(t *testing.T) {
		builder := NewGitHubEventBuilder()

		importData := GitHubImportEvent{
			Type:         "repository",
			ID:           123456,
			Name:         "test-repo",
			FullName:     "test-org/test-repo",
			Organization: "test-org",
			Data: Repository{
				ID:       123456,
				Name:     "test-repo",
				FullName: "test-org/test-repo",
			},
		}

		data, err := json.Marshal(importData)
		require.NoError(t, err)

		event, err := builder.GetPipelineEvent(t.Context(), data)
		require.NoError(t, err)
		require.NotNil(t, event)

		assert.Equal(t, "repository", event.GetType())
		assert.Equal(t, entities.Write, event.Operation())
		assert.Equal(t, "123456", event.GetPrimaryKeys()[0].Value)
	})
}

// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package confluence

import (
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfluenceSource(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	t.Run("should create confluence source without import", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"type":        "confluence",
			"webhookPath": "/confluence/webhook",
		}
		rawData, err := json.Marshal(rawConfig)
		require.NoError(t, err)

		cfg := config.GenericConfig{
			Type: "confluence",
			Raw:  rawData,
		}

		pg := pipeline.NewGroup(log)
		_, router := testutils.GetTestRouter(t)

		source, err := NewConfluenceSource(t.Context(), log, cfg, pg, router)
		require.NoError(t, err)
		assert.NotNil(t, source)

		err = source.Close()
		assert.NoError(t, err)
	})

	t.Run("should fail with missing authentication for import", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"type":              "confluence",
			"webhookPath":       "/confluence/webhook",
			"importWebhookPath": "/confluence/import",
		}
		rawData, err := json.Marshal(rawConfig)
		require.NoError(t, err)

		cfg := config.GenericConfig{
			Type: "confluence",
			Raw:  rawData,
		}

		pg := pipeline.NewGroup(log)
		_, router := testutils.GetTestRouter(t)

		_, err = NewConfluenceSource(t.Context(), log, cfg, pg, router)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username")
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("should validate basic config", func(t *testing.T) {
		cfg := &Config{}
		cfg.withDefault()
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("should require username for import", func(t *testing.T) {
		cfg := &Config{
			ImportWebhookPath: "/import",
		}
		cfg.withDefault()
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "username")
	})

	t.Run("should require API token for import", func(t *testing.T) {
		cfg := &Config{
			ImportWebhookPath: "/import",
			Username:          config.SecretSource("test-user"),
		}
		cfg.withDefault()
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API token")
	})

	t.Run("should require base URL for import", func(t *testing.T) {
		cfg := &Config{
			ImportWebhookPath: "/import",
			Username:          config.SecretSource("test-user"),
			APIToken:          config.SecretSource("test-token"),
		}
		cfg.withDefault()
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "baseUrl")
	})
}

func TestConfluenceClient_NewClient(t *testing.T) {
	logger := logrus.New()
	client, err := NewConfluenceClient("test-user", "test-token", "https://test.atlassian.net", logger)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "test-user", client.username)
	assert.Equal(t, "test-token", client.apiToken)
	assert.Equal(t, "https://test.atlassian.net", client.baseURL)
}

func TestConfluenceSource_isItemTypeEnabled(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	t.Run("should return true when no itemTypes specified", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"type":        "confluence",
			"webhookPath": "/confluence/webhook",
		}
		rawData, err := json.Marshal(rawConfig)
		require.NoError(t, err)

		cfg := config.GenericConfig{
			Type: "confluence",
			Raw:  rawData,
		}

		pg := pipeline.NewGroup(log)
		_, router := testutils.GetTestRouter(t)

		source, err := NewConfluenceSource(t.Context(), log, cfg, pg, router)
		require.NoError(t, err)

		confluenceSource := source.(*ConfluenceSource)
		assert.True(t, confluenceSource.isItemTypeEnabled("space"))
		assert.True(t, confluenceSource.isItemTypeEnabled("page"))
		assert.True(t, confluenceSource.isItemTypeEnabled("comment"))
	})

	t.Run("should respect itemTypes configuration", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"type":        "confluence",
			"webhookPath": "/confluence/webhook",
			"itemTypes":   []string{"space", "page"},
		}
		rawData, err := json.Marshal(rawConfig)
		require.NoError(t, err)

		cfg := config.GenericConfig{
			Type: "confluence",
			Raw:  rawData,
		}

		pg := pipeline.NewGroup(log)
		_, router := testutils.GetTestRouter(t)

		source, err := NewConfluenceSource(t.Context(), log, cfg, pg, router)
		require.NoError(t, err)

		confluenceSource := source.(*ConfluenceSource)
		assert.True(t, confluenceSource.isItemTypeEnabled("space"))
		assert.True(t, confluenceSource.isItemTypeEnabled("page"))
		assert.False(t, confluenceSource.isItemTypeEnabled("comment"))
		assert.False(t, confluenceSource.isItemTypeEnabled("attachment"))
	})

	t.Run("should return false for non-configured types", func(t *testing.T) {
		rawConfig := map[string]interface{}{
			"type":        "confluence",
			"webhookPath": "/confluence/webhook",
			"itemTypes":   []string{"space"},
		}
		rawData, err := json.Marshal(rawConfig)
		require.NoError(t, err)

		cfg := config.GenericConfig{
			Type: "confluence",
			Raw:  rawData,
		}

		pg := pipeline.NewGroup(log)
		_, router := testutils.GetTestRouter(t)

		source, err := NewConfluenceSource(t.Context(), log, cfg, pg, router)
		require.NoError(t, err)

		confluenceSource := source.(*ConfluenceSource)
		assert.True(t, confluenceSource.isItemTypeEnabled("space"))
		assert.False(t, confluenceSource.isItemTypeEnabled("page"))
		assert.False(t, confluenceSource.isItemTypeEnabled("comment"))
	})
}

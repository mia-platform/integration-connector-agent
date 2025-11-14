// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package config_test

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/processors/mapper"
	"github.com/mia-platform/integration-connector-agent/internal/sinks/mongo"

	"github.com/stretchr/testify/require"
)

func TestWriterConfig(t *testing.T) {
	cfg, err := config.LoadServiceConfiguration("testdata/all-writer-config.json")
	require.NoError(t, err)

	writers := cfg.Integrations[0].Pipelines[0].Sinks
	require.NotNil(t, writers)

	processors := cfg.Integrations[0].Pipelines[0].Processors
	require.NotNil(t, processors)

	mappedSinks := map[string]config.GenericConfig{}
	for _, writer := range writers {
		mappedSinks[writer.Type] = writer
	}

	mappedProcessors := map[string]config.GenericConfig{}
	for _, p := range processors {
		mappedProcessors[p.Type] = p
	}

	secretValue := "my-secret-env"
	t.Setenv("TEST_SECRET_ENV", secretValue)

	t.Run("mongo", func(t *testing.T) {
		sinkConfig, ok := mappedSinks["mongo"]
		require.True(t, ok)

		mongoConfig, err := config.GetConfig[*mongo.Config](sinkConfig)
		require.NoError(t, err)
		require.Equal(t, &mongo.Config{
			URL:        config.SecretSource(secretValue),
			Collection: "my-collection",
		}, mongoConfig)
	})

	t.Run("mapper", func(t *testing.T) {
		processorCfg, ok := mappedProcessors["mapper"]
		require.True(t, ok)

		mapperConfig, err := config.GetConfig[mapper.Config](processorCfg)
		require.NoError(t, err)
		require.JSONEq(t,
			`{"key": "{{ issue.key }}","summary": "{{ issue.fields.summary }}","createdAt": "{{ issue.fields.created }}","description": "{{ issue.fields.description }}"}`,
			string(mapperConfig.OutputEvent),
		)
	})
}

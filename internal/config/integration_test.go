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

	writers := cfg.Integrations[0].Sinks
	require.NotNil(t, writers)

	processors := cfg.Integrations[0].Processors
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

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

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadServiceConfiguration(t *testing.T) {
	t.Run("invalid configuration not match schema", func(t *testing.T) {
		config, err := LoadServiceConfiguration("./testdata/invalid-config.json")
		require.ErrorContains(t, err, "configuration not valid: json schema validation errors:")
		require.Nil(t, config)
	})

	t.Run("returns configuration", func(t *testing.T) {
		config, err := LoadServiceConfiguration("./testdata/config.json")
		require.NoError(t, err)
		require.Equal(t, &Configuration{
			Integrations: []Integrations{
				{
					Type: "jira",
					Authentication: Authentication{
						Secret: SecretSource{
							FromEnv:  "SECRET_ENV",
							FromFile: "./secret/file",
						},
					},
					Writers: []Writer{
						{
							Type: "mongo",
							URL: SecretSource{
								FromEnv: "MONGO_URL",
							},
							OutputEvent: map[string]any{
								"key":         "{{ issue.key }}",
								"summary":     "{{ issue.fields.summary }}",
								"createdAt":   "{{ issue.fields.created }}",
								"description": "{{ issue.fields.description }}",
							},
						},
					},
				},
			},
		}, config)
	})
}

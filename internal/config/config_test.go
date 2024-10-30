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
	t.Setenv("TEST_LOAD_SERVICE_MONGO_URL", "mongodb://localhost:27017")

	tests := map[string]struct {
		path            string
		expectedError   string
		expectedContent *Configuration
	}{
		"invalid configuration not match schema": {
			path:          "./testdata/invalid-config.json",
			expectedError: "configuration not valid: json schema validation errors:",
		},
		"configuration not found": {
			path:          "./testdata/not-exist",
			expectedError: "configuration not valid: open ./testdata/not-exist: no such file or directory",
		},
		"not json config": {
			path:          "./testdata/invalid-json.json",
			expectedError: "configuration not valid: error validating: unexpected EOF",
		},
		"config is parsed correctly": {
			path: "./testdata/config.json",
			expectedContent: &Configuration{
				Integrations: []Integration{
					{
						Type: "jira",
						Authentication: Authentication{
							Secret: SecretSource("MY_SECRET"),
						},
						Writers: []Writer{
							{
								Type:       "mongo",
								URL:        SecretSource("mongodb://localhost:27017"),
								Collection: "my-collection",
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
			},
		},
		"invalid config if integrations is empty": {
			path:          "./testdata/empty-integrations.json",
			expectedError: "configuration not valid: json schema validation errors:",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			config, err := LoadServiceConfiguration(test.path)
			if test.expectedError != "" {
				require.ErrorContains(t, err, test.expectedError)
				require.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedContent, config)
			}
		})
	}
}

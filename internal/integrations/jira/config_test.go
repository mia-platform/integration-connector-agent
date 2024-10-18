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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfiguration(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		path                  string
		expectedConfiguration *Configuration
		expectedErr           string
	}{
		"wrong file return error": {
			path:        filepath.Join("testdata", "missing"),
			expectedErr: "no such file or directory",
		},
		"configuration is read from valid file": {
			path: filepath.Join("testdata", "valid.json"),
			expectedConfiguration: &Configuration{
				Secret: WebhookSecret{
					FromEnv:  "ENV_NAME",
					FromFile: "file/path",
				},
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			config, err := ReadConfiguration(test.path)
			switch len(test.expectedErr) {
			case 0:
				assert.NoError(t, err)
				assert.Equal(t, test.expectedConfiguration, config)
			default:
				assert.ErrorContains(t, err, test.expectedErr)
				assert.Nil(t, config)
			}
		})
	}
}

func TestWebhookSecret(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configuration  *Configuration
		expectedSecret string
	}{
		"from missing env return empty string": {
			configuration: &Configuration{
				Secret: WebhookSecret{
					FromEnv: "ENV_NAME",
				},
			},
		},
		"from missing file return empty string": {
			configuration: &Configuration{
				Secret: WebhookSecret{
					FromFile: filepath.Join("testdata", "missing"),
				},
			},
		},
		"from missing secret section return emptry string": {
			configuration: &Configuration{},
		},
		"from valid file return secret string": {
			configuration: &Configuration{
				Secret: WebhookSecret{
					FromFile: filepath.Join("testdata", "secret"),
				},
			},
			expectedSecret: "secret-from-file",
		},
		"with both from env has priority": {
			configuration: &Configuration{
				Secret: WebhookSecret{
					FromEnv:  "ENV_NAME",
					FromFile: filepath.Join("testdata", "secret"),
				},
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			secret := test.configuration.WebhookSecret()
			assert.Equal(t, test.expectedSecret, secret)
		})
	}
}

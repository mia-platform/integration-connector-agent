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

package customprocessor

import (
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/stretchr/testify/require"
)

func TestNewPlugin(t *testing.T) {
	testCases := map[string]struct {
		modulePath  string
		expectError error
		message     string
	}{
		"fail to load plugin on invalid path": {
			modulePath:  "./invalid/path/to/plugin.so",
			expectError: ErrPluginLoadFailed,
		},
		"fail to load plugin on empty path": {
			modulePath:  "",
			expectError: ErrPluginLoadFailed,
		},
		"fail to load invalid plugin": {
			modulePath:  "./testdata/invalidplugin.so",
			expectError: ErrPluginLoadFailed,
		},
		"successfully load valid plugin": {
			modulePath: "./testdata/example-valid-plugin.so",
			message:    "WARN: You need to run make test/build-plugin-so to generate the plugin before running tests",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cfg := Config{
				ModulePath: tc.modulePath,
			}

			pluginProcessor, err := New(cfg)
			if tc.expectError != nil {
				require.ErrorIs(t, err, tc.expectError)
				require.Nil(t, pluginProcessor)
				return
			}
			require.NoError(t, err, tc.message)
		})
	}
}

func TestProcess(t *testing.T) {
	inputData := `{
		"key":"123",
		"fields": {
			"summary":"this is the summary",
			"created":"2021-01-01",
			"description":"this is the description",
			"history": { "previous": "something" },
			"changed": "something else"
		}
	}`

	testCases := map[string]struct {
		modulePath   string
		data         string
		message      string
		expectedData map[string]any
	}{
		"successfully invoke plugin process function": {
			modulePath: "./testdata/example-valid-plugin.so",
			message:    "WARN: You need to run make test/build-plugin-so to generate the plugin before running tests",
			data:       inputData,
			expectedData: map[string]any{
				"key": "123",
				"fields": map[string]any{
					"summary":     "this is the summary",
					"created":     "2021-01-01",
					"description": "this is the description",
					"history":     map[string]any{"previous": "something"},
					"changed":     "something else",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cfg := Config{
				ModulePath: tc.modulePath,
			}

			pluginProcessor, err := New(cfg)
			require.NoError(t, err, tc.message)

			event := entities.PipelineEvent(&entities.Event{
				OriginalRaw: []byte(tc.data),
			})

			result, err := pluginProcessor.Process(event)
			require.NoError(t, err)
			require.NotNil(t, result)

			expectedBytes, err := json.Marshal(tc.expectedData)
			require.NoError(t, err)
			require.JSONEq(t, string(expectedBytes), string(result.Data()))
		})
	}
}

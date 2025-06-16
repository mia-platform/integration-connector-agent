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

package hcgp

import (
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

const validPluginPath = "./testdata/mockplugin/mockplugin"

func TestNewProcessor(t *testing.T) {
	testCases := map[string]struct {
		modulePath  string
		initOptions map[string]interface{}
		expectError error
	}{
		"fail to load plugin on invalid path": {
			modulePath:  "./invalid/path/to/plugin",
			expectError: ErrPluginInitialization,
		},
		"fail to load plugin on empty path": {
			modulePath:  "",
			expectError: ErrPluginInitialization,
		},
		"fail to load invalid plugin": {
			modulePath:  "./testdata/invalidplugin",
			expectError: ErrPluginInitialization,
		},
		"successfully load valid plugin": {
			modulePath: validPluginPath,
		},
		"with init options": {
			modulePath: validPluginPath,
			initOptions: map[string]interface{}{
				"option1": "value1",
			},
		},
		"with failing init options": {
			modulePath: validPluginPath,
			initOptions: map[string]interface{}{
				"fail": true,
			},
			expectError: ErrPluginInitialization,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cfg := Config{
				ModulePath:  tc.modulePath,
				InitOptions: tc.initOptions,
			}

			l, _ := test.NewNullLogger()
			pluginProcessor, err := New(l, cfg)
			if tc.expectError != nil {
				require.ErrorIs(t, err, tc.expectError)
				require.Nil(t, pluginProcessor)
				return
			}
			require.NoError(t, err, "WARN: You may need to run make test/build-plugin to generate the plugin before running tests")
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
		expectedData map[string]any
	}{
		"successfully invoke plugin process function": {
			modulePath: validPluginPath,
			data:       inputData,
			expectedData: map[string]any{
				"data": "processed by CustomProcessor",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cfg := Config{
				ModulePath: tc.modulePath,
			}

			l, _ := test.NewNullLogger()
			pluginProcessor, err := New(l, cfg)
			require.NoError(t, err, "WARN: You may need to run make test/build-plugin to generate the plugin before running tests")

			defer pluginProcessor.(*Plugin).Close()

			event := entities.PipelineEvent(&entities.Event{
				OriginalRaw: []byte(tc.data),
				Type:        "whatever",
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

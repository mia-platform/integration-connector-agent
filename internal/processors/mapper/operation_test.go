// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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

package mapper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNewOperation(t *testing.T) {
	defaultTemplate := "{{input.key}}"
	defaultKeys := []string{"input.key"}

	testCases := map[string]struct {
		jsonTemplate string
		keyToUpdate  string
		inputJSON    string

		expectedOperation      operation
		expectedOutputJSON     string
		expectedApplyError     string
		expectedOperationError string
	}{
		"input JSON is not a correct json": {
			keyToUpdate: "output",
			inputJSON:   "{",
			expectedOperation: operation{
				operationType: set,
				valueKeys:     defaultKeys,
			},
			expectedApplyError: fmt.Sprintf("%s: unexpected end of JSON input", errTransform),
		},
		"error without key to update": {
			inputJSON: `{}`,
			expectedOperation: operation{
				operationType: set,
				valueKeys:     defaultKeys,
			},
			expectedApplyError: "error transforming data: output key is empty",
		},
		"with static data": {
			jsonTemplate: "\"static-data\"",
			keyToUpdate:  "output",
			inputJSON:    `{}`,

			expectedOperation: operation{
				operationType: set,
			},
			expectedOutputJSON: `{"output": "static-data"}`,
		},
		"set string": {
			inputJSON:   `{"input": {"key": "my-value"}}`,
			keyToUpdate: "output.key",

			expectedOperation: operation{
				valueKeys:     defaultKeys,
				operationType: set,
			},
			expectedOutputJSON: `{"output": {"key": "my-value"}}`,
		},
		"set object": {
			inputJSON:   `{"input": {"key": {"foo": "my-value"}}}`,
			keyToUpdate: "output.key",

			expectedOperation: operation{
				valueKeys:     defaultKeys,
				operationType: set,
			},
			expectedOutputJSON: `{"output": {"key": {"foo": "my-value"}}}`,
		},
		"set array": {
			keyToUpdate: "output",
			inputJSON:   `{"input": {"key": ["foo", "bar"]}}`,

			expectedOperation: operation{
				valueKeys:     defaultKeys,
				operationType: set,
			},
			expectedOutputJSON: `{"output": ["foo", "bar"]}`,
		},
		"error composite fields": {
			jsonTemplate: "{{key}}-{{field.foo}}",
			keyToUpdate:  "key",
			inputJSON:    `{"key": "key", "field": {"foo": "bar"}}`,

			expectedOperationError: "error creating operation: unsupported combine template: {{key}}-{{field.foo}}",
		},
		"set array of numbers": {
			keyToUpdate:  "output",
			inputJSON:    `{"input": [1, 2, 3]}`,
			jsonTemplate: "{{ input }}",

			expectedOperation: operation{
				valueKeys:     []string{"input"},
				operationType: set,
			},
			expectedOutputJSON: `{"output": [1, 2, 3]}`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			templateStr := tc.jsonTemplate
			if templateStr == "" {
				templateStr = defaultTemplate
			}
			template := gjson.Parse(templateStr)
			actual, err := newOperation(tc.keyToUpdate, template)
			if tc.expectedOperationError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedOperationError)
				return
			}

			expectedOperation := tc.expectedOperation
			expectedOperation.templateValue = template.Value()
			expectedOperation.keyToUpdate = tc.keyToUpdate
			require.Equal(t, expectedOperation, actual)

			t.Run("apply", func(t *testing.T) {
				output, err := actual.apply([]byte(tc.inputJSON), []byte(`{}`))
				if tc.expectedApplyError != "" {
					require.EqualError(t, err, tc.expectedApplyError)
					return
				}
				require.NoError(t, err)
				require.JSONEq(t, tc.expectedOutputJSON, string(output))
			})
		})
	}
}

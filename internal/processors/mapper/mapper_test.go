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

package mapper

import (
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestMapper(t *testing.T) {
	inputData := `{
	"key":"123",
	"fields": {
		"summary":"this is the summary",
		"created":"2021-01-01",
		"description":"this is the description",
		"history": {
			"previous": "something"
		},
		"changed": "something else"
	}
}`

	testCases := map[string]struct {
		model           string
		dataToTransform string

		expectNewError          string
		expectTransformError    string
		expectedTransformedData map[string]any
	}{
		"first level model": {
			model: `{
	"key":"{{ key }}",
	"summary":"{{ fields.summary }}",
	"createdAt":"{{ fields.created }}",
	"description":"{{ fields.description }}"
}`,
			dataToTransform: inputData,
			expectedTransformedData: map[string]any{
				"key":         "123",
				"summary":     "this is the summary",
				"createdAt":   "2021-01-01",
				"description": "this is the description",
			},
		},
		"nested model": {
			model: `{
	"key":"{{ key }}",
	"dataObj": {
		"summary":"{{ fields.summary }}",
		"description":"{{ fields.description }}"
	},
	"history": [
		"{{ fields.history }}",
		"{{ fields.changed }}"
	],
	"createdAt":"{{ fields.created }}"
}`,
			dataToTransform: inputData,

			expectedTransformedData: map[string]any{
				"key": "123",
				"dataObj": map[string]any{
					"summary":     "this is the summary",
					"description": "this is the description",
				},
				"history": []any{
					map[string]any{"previous": "something"},
					"something else",
				},
				"createdAt": "2021-01-01",
			},
		},
		"with static fields": {
			model: `{
	"key":"{{ key }}",
	"state": "public",
	"data": [1, 2, 3]
}`,
			dataToTransform: inputData,

			expectedTransformedData: map[string]any{
				"key":   "123",
				"state": "public",
				"data":  []any{float64(1), float64(2), float64(3)},
			},
		},
		"throws if create operation fails - combined operation": {
			model: `{
	"key": "{{combined}}-{{key}}"
}`,
			expectNewError: "error creating operation: unsupported combine template: {{combined}}-{{key}}",
		},
		"all event in a subfield": {
			model: `{
	"event": "{{@this}}"
}`,
			dataToTransform: inputData,
			expectedTransformedData: map[string]any{
				"event": map[string]any{
					"key": "123",
					"fields": map[string]any{
						"summary":     "this is the summary",
						"created":     "2021-01-01",
						"description": "this is the description",
						"history": map[string]any{
							"previous": "something",
						},
						"changed": "something else",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cfg := Config{
				OutputEvent: []byte(tc.model),
			}

			mapper, err := New(cfg)
			if tc.expectNewError != "" {
				require.EqualError(t, err, tc.expectNewError)
				require.Nil(t, mapper)
				return
			}
			require.NoError(t, err)

			event := entities.PipelineEvent(&entities.Event{
				OriginalRaw: []byte(tc.dataToTransform),
			})

			event, err = mapper.Process(event)
			if tc.expectTransformError != "" {
				require.EqualError(t, err, tc.expectTransformError)
				require.Nil(t, event)
				return
			}
			require.NoError(t, err)

			out, err := json.Marshal(tc.expectedTransformedData)
			require.NoError(t, err)
			require.JSONEq(t, string(out), string(event.Data()))
		})
	}
}

func TestGenerateOperations(t *testing.T) {
	testCases := map[string]struct {
		json string

		expected []operation
	}{
		"generate model with fields at first level": {
			json: `{
			"key":"{{ key }}",
			"summary":"{{ fields.summary }}",
			"createdAt":"{{ fields.created }}",
			"description":"{{ fields.description }}"
		}`,

			expected: []operation{
				{
					keyToUpdate:   "key",
					valueKeys:     []string{"key"},
					operationType: set,
					templateValue: "{{ key }}",
				},
				{
					keyToUpdate:   "summary",
					valueKeys:     []string{"fields.summary"},
					operationType: set,
					templateValue: "{{ fields.summary }}",
				},
				{
					keyToUpdate:   "createdAt",
					valueKeys:     []string{"fields.created"},
					operationType: set,
					templateValue: "{{ fields.created }}",
				},
				{
					keyToUpdate:   "description",
					valueKeys:     []string{"fields.description"},
					operationType: set,
					templateValue: "{{ fields.description }}",
				},
			},
		},
		"array field": {
			json: `{
			"data": ["{{ key }}", "{{ fields.summary }}"]
		}`,

			expected: []operation{
				{
					keyToUpdate:   "data.0",
					valueKeys:     []string{"key"},
					operationType: set,
					templateValue: "{{ key }}",
				},
				{
					keyToUpdate:   "data.1",
					valueKeys:     []string{"fields.summary"},
					operationType: set,
					templateValue: "{{ fields.summary }}",
				},
			},
		},
		"generate model with fields with object and array": {
			json: `{
			"key":"{{ key }}",
			"dataObj": {
				"summary":"{{ fields.summary }}",
				"description":"{{ fields.description }}"
			}
			"history": [
				"{{ fields.bar }}",
				"{{ fields.foo }}"
			],
			"createdAt":"{{ fields.created }}",
		}`,

			expected: []operation{
				{
					keyToUpdate:   "key",
					valueKeys:     []string{"key"},
					operationType: set,
					templateValue: "{{ key }}",
				},
				{
					keyToUpdate:   "dataObj.summary",
					valueKeys:     []string{"fields.summary"},
					operationType: set,
					templateValue: "{{ fields.summary }}",
				},
				{
					keyToUpdate:   "dataObj.description",
					valueKeys:     []string{"fields.description"},
					operationType: set,
					templateValue: "{{ fields.description }}",
				},
				{
					keyToUpdate:   "history.0",
					valueKeys:     []string{"fields.bar"},
					operationType: set,
					templateValue: "{{ fields.bar }}",
				},
				{
					keyToUpdate:   "history.1",
					valueKeys:     []string{"fields.foo"},
					operationType: set,
					templateValue: "{{ fields.foo }}",
				},
				{
					keyToUpdate:   "createdAt",
					valueKeys:     []string{"fields.created"},
					operationType: set,
					templateValue: "{{ fields.created }}",
				},
			},
		},
		"array of numbers": {
			json: `{
	"key":"{{ key }}",
	"state": "public",
	"data": [1, 2, 3]
}`,

			expected: []operation{
				{
					keyToUpdate:   "key",
					valueKeys:     []string{"key"},
					operationType: set,
					templateValue: "{{ key }}",
				},
				{
					keyToUpdate:   "state",
					templateValue: "public",
				},
				{
					keyToUpdate:   "data.0",
					templateValue: float64(1),
				},
				{
					keyToUpdate:   "data.1",
					templateValue: float64(2),
				},
				{
					keyToUpdate:   "data.2",
					templateValue: float64(3),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			parsedJSON := gjson.Parse(tc.json)

			actual, err := generateOperations(parsedJSON)
			require.NoError(t, err)

			require.Equal(t, tc.expected, actual)
		})
	}
}

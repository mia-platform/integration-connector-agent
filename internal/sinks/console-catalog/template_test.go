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

package consolecatalog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplatize(t *testing.T) {
	testCases := []struct {
		name       string
		template   string
		data       []byte
		expected   string
		expectdErr error
	}{
		{
			name:     "no templates",
			template: "no templates here",
			data:     []byte(`{"name": "the-name"}`),
			expected: "no templates here",
		},
		{
			name:     "invalid template",
			template: "{{name",
			data:     []byte(`{"name": "the-name"}`),
			expected: "{{name",
		},
		{
			name:     "simple template",
			template: "{{name}}",
			data:     []byte(`{"name": "the-name"}`),
			expected: "the-name",
		},
		{
			name:     "template with multiple keys",
			template: "{{name}}-{{id}}",
			data:     []byte(`{"name": "the-name", "id": "12345"}`),
			expected: "the-name-12345",
		},
		{
			name:     "template with nested key",
			template: "{{object.name}}",
			data:     []byte(`{"object": {"name": "the-name", "id": "12345"}}`),
			expected: "the-name",
		},
		{
			name:     "simple template with extra words",
			template: "Hello, {{name}}!",
			data:     []byte(`{"name": "World"}`),
			expected: "Hello, World!",
		},
		{
			name:     "simple template with extra words and spaces",
			template: "Hello, {{ name }}!",
			data:     []byte(`{"name": "World"}`),
			expected: "Hello, World!",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := templetize(tc.template, tc.data)
			if tc.expectdErr != nil {
				require.Equal(t, tc.expectdErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "example",
			expected: "example",
		},
		{
			name:     "string with spaces",
			input:    "example string",
			expected: "example-string",
		},
		{
			name:     "string with special characters",
			input:    "example@string!",
			expected: "example-string",
		},
		{
			name:     "string with numbers",
			input:    "example123",
			expected: "example123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with dashes",
			input:    "example-string",
			expected: "example-string",
		},
		{
			name:     "string with underscores",
			input:    "example_string",
			expected: "example_string",
		},
		{
			name:     "string with Uppercase letters",
			input:    "Example String",
			expected: "example-string",
		},
		{
			name:     "string with leading and trailing spaces",
			input:    "   example   ",
			expected: "example",
		},
		{
			name:     "string starting and ending with hyphens",
			input:    "-example string-",
			expected: "example-string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := slugify(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

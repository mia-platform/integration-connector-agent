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
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		path            string
		expectedError   string
		expectedContent string
	}{
		"missing file return error": {
			path:          filepath.Join("testdata", "readfile", "missing"),
			expectedError: "no such file or directory",
		},
		"json file is parsed correctly": {
			path:            filepath.Join("testdata", "readfile", "file.json"),
			expectedContent: `{"key": "value"}`,
		},
		"file without extension is parsed correctly": {
			path:            filepath.Join("testdata", "readfile", "file-without-extension"),
			expectedContent: `{"key": "no-extension"}`,
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			out, err := readFile(test.path)
			switch len(test.expectedError) {
			case 0:
				assert.NoError(t, err)
				assert.Equal(t, test.expectedContent, strings.TrimSpace(string(out)))
			default:
				assert.ErrorContains(t, err, test.expectedError)
			}
		})
	}
}

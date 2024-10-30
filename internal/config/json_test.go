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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadJSON(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		path           string
		expectedError  string
		expectedObject *fakeObj
	}{
		"missing file return error": {
			path:          filepath.Join("testdata", "readjson", "missing"),
			expectedError: "no such file or directory",
		},
		"no json file return error": {
			path:          filepath.Join("testdata", "readjson", "nonjson.toml"),
			expectedError: "invalid character 'o'",
		},
		"file with wrong type return error": {
			path:          filepath.Join("testdata", "readjson", "wrong-data.json"),
			expectedError: "cannot unmarshal number",
		},
		"file with invalid json data return error": {
			path:          filepath.Join("testdata", "readjson", "invalid.json"),
			expectedError: "invalid character '}'",
		},
		"json file is parsed correctly": {
			path: filepath.Join("testdata", "readjson", "file.json"),
			expectedObject: &fakeObj{
				Key: "value",
			},
		},
		"file without extension is parsed correctly": {
			path: filepath.Join("testdata", "readjson", "file-without-extension"),
			expectedObject: &fakeObj{
				Key: "no-extension",
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			obj := new(fakeObj)

			err := ReadJSONFile(test.path, obj)
			switch len(test.expectedError) {
			case 0:
				assert.NoError(t, err)
				assert.Equal(t, test.expectedObject, obj)
			default:
				assert.ErrorContains(t, err, test.expectedError)
			}
		})
	}
}

type fakeObj struct {
	Key string `json:"key"`
}

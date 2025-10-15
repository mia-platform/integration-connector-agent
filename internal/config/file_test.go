// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

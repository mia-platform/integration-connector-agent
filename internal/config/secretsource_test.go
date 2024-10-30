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

func TestSecret(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		source         SecretSource
		expectedSecret string
	}{
		"from missing env return empty string": {
			source: SecretSource{
				FromEnv: "ENV_NAME",
			},
		},
		"from missing file return empty string": {
			source: SecretSource{
				FromFile: filepath.Join("testdata", "secretsource", "missing"),
			},
		},
		"from missing secret section return emptry string": {
			source: SecretSource{},
		},
		"from valid file return secret string": {
			source: SecretSource{
				FromFile: filepath.Join("testdata", "secretsource", "secret"),
			},
			expectedSecret: "secret-from-file",
		},
		"with both from env has priority": {
			source: SecretSource{
				FromEnv:  "ENV_NAME",
				FromFile: filepath.Join("testdata", "secretsource", "secret"),
			},
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			secret := test.source.Secret()
			assert.Equal(t, test.expectedSecret, secret)
		})
	}
}

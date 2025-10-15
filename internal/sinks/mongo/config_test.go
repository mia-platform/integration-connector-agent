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

package mongo

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/config"

	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	testCases := map[string]struct {
		config Config

		expectedError string
	}{
		"without URI": {
			config: Config{},

			expectedError: "url is required",
		},
		"without collection": {
			config: Config{
				URL: config.SecretSource("mongodb://localhost:27017"),
			},

			expectedError: "collection is required",
		},
		"valid config": {
			config: Config{
				URL:        config.SecretSource("mongodb://localhost:27017"),
				Collection: "test",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectedError != "" {
				require.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}

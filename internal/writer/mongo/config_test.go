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

package mongo

import (
	"fmt"
	"testing"

	"github.com/mia-platform/data-connector-agent/internal/config"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	testCases := map[string]struct {
		config Config

		expectedError  string
		expectedConfig Config
	}{
		"without URI": {
			config: Config{},

			expectedError: fmt.Sprintf("%s: URI is empty", config.ErrConfigNotValid),
		},
		"without collection": {
			config: Config{
				URI: config.SecretSource("mongodb://localhost:27017"),
			},

			expectedError: fmt.Sprintf("%s: collection is empty", config.ErrConfigNotValid),
		},
		"without output event": {
			config: Config{
				URI:        config.SecretSource("mongodb://localhost:27017"),
				Collection: "test",
			},

			expectedError: fmt.Sprintf("%s: output event not set", config.ErrConfigNotValid),
		},
		"default ID field not found in output event": {
			config: Config{
				URI:         config.SecretSource("mongodb://localhost:27017"),
				Collection:  "test",
				OutputEvent: map[string]any{},
			},

			expectedError: fmt.Sprintf(`%s: ID field "_id" not found in output event`, config.ErrConfigNotValid),
		},
		"custom ID field not found in output event": {
			config: Config{
				URI:         config.SecretSource("mongodb://localhost:27017"),
				Collection:  "test",
				OutputEvent: map[string]any{},
				IDField:     "custom_id",
			},

			expectedError: fmt.Sprintf(`%s: ID field "custom_id" not found in output event`, config.ErrConfigNotValid),
		},
		"set default ID field if empty": {
			config: Config{
				URI:        config.SecretSource("mongodb://localhost:27017"),
				Collection: "test",
				OutputEvent: map[string]any{
					"_id": "my-id",
				},
			},

			expectedConfig: Config{
				URI:        config.SecretSource("mongodb://localhost:27017"),
				Collection: "test",
				OutputEvent: map[string]any{
					"_id": "my-id",
				},
				IDField: "_id",
			},
		},
		"set custom ID field if empty": {
			config: Config{
				URI:        config.SecretSource("mongodb://localhost:27017"),
				Collection: "test",
				OutputEvent: map[string]any{
					"custom_id": "my-id",
				},
				IDField: "custom_id",
			},

			expectedConfig: Config{
				URI:        config.SecretSource("mongodb://localhost:27017"),
				Collection: "test",
				OutputEvent: map[string]any{
					"custom_id": "my-id",
				},
				IDField: "custom_id",
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
			require.Equal(t, tc.expectedConfig, tc.config)
		})
	}
}

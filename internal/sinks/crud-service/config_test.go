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

package crudservice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidate(t *testing.T) {
	testCases := []struct {
		name        string
		config      *Config
		expectedErr error
	}{
		{
			name:        "valid config",
			config:      &Config{URL: "http://example.com"},
			expectedErr: nil,
		},
		{
			name:        "invalid URL",
			config:      &Config{URL: "zzz:////\n\ninvalid-url"},
			expectedErr: ErrInvalidURL,
		},
		{
			name:        "empty URL",
			config:      &Config{URL: ""},
			expectedErr: ErrURLNotSet,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("injects default primary key", func(t *testing.T) {
		config := &Config{URL: "http://example.com"}
		err := config.Validate()
		require.NoError(t, err)
		require.Equal(t, "_eventId", config.PrimaryKey)
	})
}

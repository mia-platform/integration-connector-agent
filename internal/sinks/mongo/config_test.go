// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

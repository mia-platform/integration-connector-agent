// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

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
		require.Equal(t, DefaultPrimaryKey, config.PrimaryKey)
	})
}

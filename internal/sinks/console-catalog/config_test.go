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

func TestConfigValidate(t *testing.T) {
	testCases := []struct {
		name                 string
		config               *Config
		expectedErr          error
		expectedMissingField string
	}{
		{
			name: "valid config",
			config: &Config{
				URL:              "http://example.com",
				TenantID:         "tenant-id",
				ItemType:         "item-type",
				ClientID:         "client-id",
				ClientSecret:     "client-secret",
				ItemIDTemplate:   "item-id-template",
				ItemNameTemplate: "item-name-template",
			},
			expectedErr: nil,
		},
		{
			name:        "invalid URL",
			config:      &Config{URL: "zzz:////\n\ninvalid-url"},
			expectedErr: ErrInvalidURL,
		},
		{
			name:        "missing URL",
			config:      &Config{},
			expectedErr: ErrURLNotSet,
		},
		{
			name: "missing tenant ID",
			config: &Config{
				URL:      "http://example.com",
				ItemType: "item-type",
				ClientID: "client-id",
			},
			expectedErr:          ErrMissingField,
			expectedMissingField: "tenantId",
		},
		{
			name: "missing item type",
			config: &Config{
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ClientID: "client-id",
			},
			expectedErr:          ErrMissingField,
			expectedMissingField: "itemType",
		},
		{
			name: "missing client ID",
			config: &Config{
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ItemType: "item-type",
			},
			expectedErr:          ErrMissingField,
			expectedMissingField: "clientId",
		},
		{
			name: "missing client secret",
			config: &Config{
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ItemType: "item-type",
				ClientID: "client-id",
			},
			expectedErr:          ErrMissingField,
			expectedMissingField: "clientSecret",
		},
		{
			name: "configuration is valid even when missing item ID template",
			config: &Config{
				URL:              "http://example.com",
				TenantID:         "tenant-id",
				ItemType:         "item-type",
				ClientID:         "client-id",
				ClientSecret:     "client-secret",
				ItemNameTemplate: "item-name-template",
			},
		},
		{
			name: "missing item name template",
			config: &Config{
				URL:            "http://example.com",
				TenantID:       "tenant-id",
				ItemType:       "item-type",
				ClientID:       "client-id",
				ClientSecret:   "client-secret",
				ItemIDTemplate: "item-id-template",
			},
			expectedErr:          ErrMissingField,
			expectedMissingField: "itemNameTemplate",
		},
		{
			name: "invalid lifecycle status",
			config: &Config{
				URL:                 "http://example.com",
				TenantID:            "tenant-id",
				ItemType:            "item-type",
				ClientID:            "client-id",
				ClientSecret:        "client-secret",
				ItemIDTemplate:      "item-id-template",
				ItemNameTemplate:    "item-name-template",
				ItemLifecycleStatus: "invalid-status",
			},
			expectedErr: ErrInvalidLifecycleStatus,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				if tc.expectedMissingField != "" {
					require.ErrorContains(t, err, tc.expectedMissingField)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

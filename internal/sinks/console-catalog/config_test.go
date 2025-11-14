// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package consolecatalog

import (
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/sinks/console-catalog/consoleclient"
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
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ItemTypeDefinitionRef: consoleclient.ItemTypeDefinitionRef{
					Name:      "item-type",
					Namespace: "default",
				},
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
				URL: "http://example.com",
				ItemTypeDefinitionRef: consoleclient.ItemTypeDefinitionRef{
					Name:      "item-type",
					Namespace: "default",
				},
				ClientID: "client-id",
			},
			expectedErr:          ErrMissingField,
			expectedMissingField: "tenantId",
		},
		{
			name: "missing item type name",
			config: &Config{
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ItemTypeDefinitionRef: consoleclient.ItemTypeDefinitionRef{
					Name:      "",
					Namespace: "default",
				},
				ClientID: "client-id",
			},
			expectedErr:          ErrMissingField,
			expectedMissingField: "itemTypeDefinitionRef.name",
		},
		{
			name: "missing item type namespace",
			config: &Config{
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ItemTypeDefinitionRef: consoleclient.ItemTypeDefinitionRef{
					Name:      "item-type",
					Namespace: "",
				},
				ClientID: "client-id",
			},
			expectedErr:          ErrMissingField,
			expectedMissingField: "itemTypeDefinitionRef.namespace",
		},
		{
			name: "missing client ID",
			config: &Config{
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ItemTypeDefinitionRef: consoleclient.ItemTypeDefinitionRef{
					Name:      "item-type",
					Namespace: "default",
				},
			},
			expectedErr:          ErrMissingField,
			expectedMissingField: "clientId",
		},
		{
			name: "missing client secret",
			config: &Config{
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ItemTypeDefinitionRef: consoleclient.ItemTypeDefinitionRef{
					Name:      "item-type",
					Namespace: "default",
				},
				ClientID: "client-id",
			},
			expectedErr:          ErrMissingField,
			expectedMissingField: "clientSecret",
		},
		{
			name: "configuration is valid even when missing item ID template",
			config: &Config{
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ItemTypeDefinitionRef: consoleclient.ItemTypeDefinitionRef{
					Name:      "item-type",
					Namespace: "default",
				},
				ClientID:         "client-id",
				ClientSecret:     "client-secret",
				ItemNameTemplate: "item-name-template",
			},
		},
		{
			name: "missing item name template",
			config: &Config{
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ItemTypeDefinitionRef: consoleclient.ItemTypeDefinitionRef{
					Name:      "item-type",
					Namespace: "default",
				},
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
				URL:      "http://example.com",
				TenantID: "tenant-id",
				ItemTypeDefinitionRef: consoleclient.ItemTypeDefinitionRef{
					Name:      "item-type",
					Namespace: "default",
				},
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

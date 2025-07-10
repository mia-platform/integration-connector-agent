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

package azure

import (
	"encoding/json"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		configData     []byte
		expectedConfig *EventHubConfig
		expectedError  error
	}{
		"valid config with all fields": {
			configData: []byte(`{
	"subscriptionId": "12345",
	"tenantId": "tenant-id",
	"clientId": {
		"fromFile": "testdata/clientid"
	},
	"clientSecret": {
		"fromFile": "testdata/clientsecret"
	},
	"eventHubNamespace": "test-namespace",
	"eventHubName": "test-event-hub",
	"checkpointStorageAccountName": "test-account",
	"checkpointStorageContainerName": "test-container"
}`),
			expectedConfig: &EventHubConfig{
				AuthConfig: AuthConfig{
					SubscriptionID: "12345",
					TenantID:       "tenant-id",
					ClientID:       config.SecretSource("test-id"),
					ClientSecret:   config.SecretSource("test-secret"),
				},
				EventHubNamespace:              "test-namespace.servicebus.windows.net",
				EventHubName:                   "test-event-hub",
				CheckpointStorageAccountName:   "test-account.blob.core.windows.net",
				CheckpointStorageContainerName: "test-container",
			},
		},
		"missing subscription ID": {
			configData:    []byte(`{}`),
			expectedError: ErrMissingSubscriptionID,
		},
		"error for tenant id": {
			configData: []byte(`{
	"subscriptionId": "12345",
	"tenantId": "tenant-id"
}`),
			expectedError: ErrIncompleteAuthConfigForTenantID,
		},
		"error for client id": {
			configData: []byte(`{
	"subscriptionId": "12345",
	"clientId": {
		"fromFile": "testdata/clientid"
	}
}`),
			expectedError: ErrIncompleteAuthConfigForClientID,
		},
		"error for client secret": {
			configData: []byte(`{
	"subscriptionId": "12345",
	"clientSecret": {
		"fromFile": "testdata/clientsecret"
	}
}`),
			expectedError: ErrIncompleteAuthConfigForClientSecret,
		},
		"error for namespace": {
			configData: []byte(`{
	"subscriptionId": "12345",
	"tenantId": "tenant-id",
	"clientId": {
		"fromFile": "testdata/clientid"
	},
	"clientSecret": {
		"fromFile": "testdata/clientsecret"
	}
}`),
			expectedError: ErrEventHubNamespaceRequired,
		},
		"error for event hub name": {
			configData: []byte(`{
	"subscriptionId": "12345",
	"tenantId": "tenant-id",
	"clientId": {
		"fromFile": "testdata/clientid"
	},
	"clientSecret": {
		"fromFile": "testdata/clientsecret"
	},
	"eventHubNamespace": "test-namespace"
}`),
			expectedError: ErrEventHubNameRequired,
		},
		"error for checkpoint account": {
			configData: []byte(`{
	"subscriptionId": "12345",
	"tenantId": "tenant-id",
	"clientId": {
		"fromFile": "testdata/clientid"
	},
	"clientSecret": {
		"fromFile": "testdata/clientsecret"
	},
	"eventHubNamespace": "test-namespace",
	"eventHubName": "test-event-hub"
}`),
			expectedError: ErrCheckpointStorageAccountNameRequired,
		},
		"error for checkpoint container": {
			configData: []byte(`{
	"subscriptionId": "12345",
	"tenantId": "tenant-id",
	"clientId": {
		"fromFile": "testdata/clientid"
	},
	"clientSecret": {
		"fromFile": "testdata/clientsecret"
	},
	"eventHubNamespace": "test-namespace",
	"eventHubName": "test-event-hub",
	"checkpointStorageAccountName": "test-account"
}`),
			expectedError: ErrCheckpointStorageContainerNameRequired,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			cfg := new(EventHubConfig)
			err := json.Unmarshal(test.configData, cfg)
			require.NoError(t, err)

			err = cfg.Validate()
			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expectedConfig, cfg)
		})
	}
}

func TestAzureCredentials(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config        AuthConfig
		expectedError error
	}{
		"credentials with static config": {
			config: AuthConfig{
				SubscriptionID: "12345",
				TenantID:       "tenant-id",
				ClientID:       config.SecretSource("test-id"),
				ClientSecret:   config.SecretSource("test-secret"),
			},
		},
		"default credentials": {
			config: AuthConfig{},
		},
		"error for invalid tenant id": {
			config: AuthConfig{
				SubscriptionID: "12345",
				TenantID:       "invalid!",
				ClientID:       config.SecretSource("test-id"),
				ClientSecret:   config.SecretSource("test-secret"),
			},
			expectedError: ErrAzureClientSecretCredential,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			credentials, err := test.config.AzureTokenProvider()
			if test.expectedError != nil {
				assert.ErrorIs(t, err, test.expectedError)
				return
			}

			assert.NoError(t, err)
			_, ok := credentials.(*azidentity.ChainedTokenCredential)
			require.True(t, ok, "Expected credentials to be of type *azidentity.ChainedTokenCredential")
		})
	}
}

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
	"context"
	"encoding/json"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

var _ azure.ClientInterface = &fakeClient{}

type fakeClient struct {
	t          *testing.T
	resource   *azure.Resource
	apiVersion string
}

func (f *fakeClient) GetByID(_ context.Context, id string, apiVersion string) (*azure.Resource, error) {
	f.t.Helper()
	require.Equal(f.t, f.apiVersion, apiVersion)
	return f.resource, nil
}

func TestProcessor(t *testing.T) {
	t.Parallel()
	l, _ := test.NewNullLogger()

	testCases := map[string]struct {
		client        azure.ClientInterface
		input         entities.PipelineEvent
		expectedAsset *commons.Asset
		expectedError error
	}{
		"request a bucket storage": {
			client: &fakeClient{
				t: t,
				resource: &azure.Resource{
					Name:     "account",
					Type:     "Microsoft.Storage/storageAccounts",
					Tags:     map[string]string{"env": "test"},
					Location: "eastus",
				},
				apiVersion: "2025-01-01",
			},
			input: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{
						Key:   "resourceId",
						Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Storage/storageAccounts/account",
					},
				},
				Type:          azure.EventTypeRecordFromEventHub.String(),
				OperationType: entities.Write,
				OriginalRaw:   []byte(bucketActivityLog),
			},
			expectedAsset: commons.NewAsset("account", "Microsoft.Storage/storageAccounts", commons.AzureAssetProvider).
				WithLocation("eastus").
				WithRelationships([]string{
					"subscription/00000000-0000-0000-0000-000000000000",
					"resourceGroup/group",
				}).
				WithTags(map[string]string{"env": "test"}).
				WithRawData(func() []byte {
					event := new(azure.ActivityLogEventRecord)
					err := json.Unmarshal([]byte(bucketActivityLog), event)
					require.NoError(t, err)
					data, err := json.Marshal(event)
					require.NoError(t, err)
					return data
				}(),
				),
		},
		"request a function": {
			client: &fakeClient{
				t: t,
				resource: &azure.Resource{
					Name:     "function",
					Type:     "Microsoft.Web/sites",
					Tags:     map[string]string{"env": "test"},
					Location: "eastus",
				},
				apiVersion: "2024-11-01",
			},
			input: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{
						Key:   "resourceId",
						Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Web/sites/function",
					},
				},
				Type:          azure.EventTypeRecordFromEventHub.String(),
				OperationType: entities.Write,
				OriginalRaw:   []byte(functionActivityLog),
			},
			expectedAsset: commons.NewAsset("function", "Microsoft.Web/sites", commons.AzureAssetProvider).
				WithLocation("eastus").
				WithRelationships([]string{
					"subscription/00000000-0000-0000-0000-000000000000",
					"resourceGroup/group",
				}).
				WithTags(map[string]string{"env": "test"}).
				WithRawData(func() []byte {
					event := new(azure.ActivityLogEventRecord)
					err := json.Unmarshal([]byte(functionActivityLog), event)
					require.NoError(t, err)
					data, err := json.Marshal(event)
					require.NoError(t, err)
					return data
				}(),
				),
		},
		"request a virtual machine": {
			client: &fakeClient{
				t: t,
				resource: &azure.Resource{
					Name:     "vm",
					Type:     "Microsoft.Compute/virtualMachines",
					Tags:     map[string]string{"env": "test"},
					Location: "eastus",
				},
				apiVersion: "2024-11-01",
			},
			input: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{
						Key:   "resourceId",
						Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Compute/virtualMachines/vm",
					},
				},
				Type:          azure.EventTypeRecordFromEventHub.String(),
				OperationType: entities.Write,
				OriginalRaw:   []byte(virtualMachineActivityLog),
			},
			expectedAsset: commons.NewAsset("vm", "Microsoft.Compute/virtualMachines", commons.AzureAssetProvider).
				WithLocation("eastus").
				WithRelationships([]string{
					"subscription/00000000-0000-0000-0000-000000000000",
					"resourceGroup/group",
				}).
				WithTags(map[string]string{"env": "test"}).
				WithRawData(func() []byte {
					event := new(azure.ActivityLogEventRecord)
					err := json.Unmarshal([]byte(virtualMachineActivityLog), event)
					require.NoError(t, err)
					data, err := json.Marshal(event)
					require.NoError(t, err)
					return data
				}(),
				),
		},
		"event is from import event": {
			input: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{
						Key:   "resourceId",
						Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Storage/storageAccounts/account",
					},
				},
				Type:          azure.EventTypeFromLiveLoad.String(),
				OperationType: entities.Write,
				OriginalRaw:   []byte(liveData),
			},
			expectedAsset: commons.NewAsset("account", "Microsoft.Storage/storageAccounts", commons.AzureAssetProvider).
				WithLocation("eastus").
				WithTags(map[string]string{"env": "test"}).
				WithRelationships([]string{
					"subscription/00000000-0000-0000-0000-000000000000",
					"resourceGroup/group",
				}).
				WithRawData([]byte(liveData)),
		},
		"filter activity log not in success state": {
			input: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{
						Key:   "resourceId",
						Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Storage/storageAccounts/account",
					},
				},
				Type:          azure.EventTypeRecordFromEventHub.String(),
				OperationType: entities.Write,
				OriginalRaw:   []byte(startBucketActivityLog),
			},
			expectedError: entities.ErrDiscardEvent,
		},
		"request unsupported resource": {
			input: &entities.Event{
				PrimaryKeys: entities.PkFields{
					{
						Key:   "resourceId",
						Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Unknown/resurce/name",
					},
				},
				Type:          azure.EventTypeRecordFromEventHub.String(),
				OperationType: entities.Write,
				OriginalRaw:   []byte(unknownActivityLog),
			},
			expectedError: ErrUnsupportedEventSource,
		},
	}

	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			processor, err := New(l, config.AuthOptions{})
			require.NoError(t, err)
			processor.client = test.client
			event, err := processor.Process(test.input)
			if test.expectedError != nil {
				require.ErrorIs(t, err, test.expectedError)
				return
			}

			require.NoError(t, err)
			eventAsset := new(commons.Asset)
			err = json.Unmarshal(event.Data(), eventAsset)
			require.NoError(t, err)
			test.expectedAsset.Timestamp = eventAsset.Timestamp
			require.Equal(t, test.expectedAsset, eventAsset)
		})
	}
}

const liveData = `{
	"id": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Storage/storageAccounts/account",
	"name": "account",
	"type": "Microsoft.Storage/storageAccounts",
	"location": "eastus",
	"tags": {
		"env": "test"
	}
}`

const startBucketActivityLog = `{
	"resourceId": "/SUBSCRIPTIONS/00000000-0000-0000-0000-000000000000/RESOURCEGROUPS/GROUP/PROVIDERS/MICROSOFT.STORAGE/STORAGEACCOUNTS/ACCOUNT",
	"operationName": "MICROSOFT.STORAGE/STORAGEACCOUNTS/WRITE",
	"category": "Administrative",
	"resultType": "Start",
	"resultSignature": "Started.",
	"identity": {
		"authorization": {
			"scope": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Storage/storageAccounts/account",
			"action": "Microsoft.Storage/storageAccounts/write",
			"evidence": {
				"role": "Contributor",
				"roleAssignmentScope": "/subscriptions/00000000-0000-0000-0000-000000000000",
				"roleAssignmentId": "00000000000000000000000000000000",
				"roleDefinitionId": "00000000000000000000000000000000",
				"principalId": "00000000000000000000000000000000",
				"principalType": "ServicePrincipal"
			}
		}
	},
	"level": "Information",
	"properties": {
		"eventCategory": "Administrative",
		"entity": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Storage/storageAccounts/account",
		"message": "Microsoft.Storage/storageAccounts/write",
		"hierarchy": "00000000-0000-0000-0000-000000000000/00000000-0000-0000-0000-000000000000"
	},
	"tenantId": "00000000-0000-0000-0000-000000000000"
}`

const bucketActivityLog = `{
	"resourceId": "/SUBSCRIPTIONS/00000000-0000-0000-0000-000000000000/RESOURCEGROUPS/GROUP/PROVIDERS/MICROSOFT.STORAGE/STORAGEACCOUNTS/ACCOUNT",
	"operationName": "MICROSOFT.STORAGE/STORAGEACCOUNTS/WRITE",
	"category": "Administrative",
	"resultType": "Succeeded",
	"resultSignature": "Succeeded.",
	"identity": {
		"authorization": {
			"scope": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Storage/storageAccounts/account",
			"action": "Microsoft.Storage/storageAccounts/write",
			"evidence": {
				"role": "Contributor",
				"roleAssignmentScope": "/subscriptions/00000000-0000-0000-0000-000000000000",
				"roleAssignmentId": "00000000000000000000000000000000",
				"roleDefinitionId": "00000000000000000000000000000000",
				"principalId": "00000000000000000000000000000000",
				"principalType": "ServicePrincipal"
			}
		}
	},
	"level": "Information",
	"properties": {
		"eventCategory": "Administrative",
		"entity": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Storage/storageAccounts/account",
		"message": "Microsoft.Storage/storageAccounts/write",
		"hierarchy": "00000000-0000-0000-0000-000000000000/00000000-0000-0000-0000-000000000000"
	},
	"tenantId": "00000000-0000-0000-0000-000000000000"
}`

const virtualMachineActivityLog = `{
	"resourceId": "/SUBSCRIPTIONS/00000000-0000-0000-0000-000000000000/RESOURCEGROUPS/GROUP/PROVIDERS/MICROSOFT.COMPUTE/VIRTUALMACHINES/VM",
	"operationName": "MICROSOFT.COMPUTE/VIRTUALMACHINES/WRITE",
	"category": "Administrative",
	"resultType": "Succeeded",
	"resultSignature": "Succeeded.",
	"identity": {
		"authorization": {
			"scope": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Compute/virtualMachines/vm",
			"action": "Microsoft.Compute/virtualMachines/write",
			"evidence": {
				"role": "Contributor",
				"roleAssignmentScope": "/subscriptions/00000000-0000-0000-0000-000000000000",
				"roleAssignmentId": "00000000000000000000000000000000",
				"roleDefinitionId": "00000000000000000000000000000000",
				"principalId": "00000000000000000000000000000000",
				"principalType": "ServicePrincipal"
			}
		}
	},
	"level": "Information",
	"properties": {
		"eventCategory": "Administrative",
		"entity": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Compute/virtualMachines/vm",
		"message": "Microsoft.Compute/virtualMachines/write",
		"hierarchy": "00000000-0000-0000-0000-000000000000/00000000-0000-0000-0000-000000000000"
	},
	"tenantId": "00000000-0000-0000-0000-000000000000"
}`

const functionActivityLog = `{
	"resourceId": "/SUBSCRIPTIONS/00000000-0000-0000-0000-000000000000/RESOURCEGROUPS/GROUP/PROVIDERS/MICROSOFT.WEB/SITES/FUNCTION",
	"operationName": "MICROSOFT.WEB/SITES/WRITE",
	"category": "Administrative",
	"resultType": "Succeeded",
	"resultSignature": "Succeeded.",
	"identity": {
		"authorization": {
			"scope": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Web/sites/function",
			"action": "Microsoft.Web/sites/write",
			"evidence": {
				"role": "Contributor",
				"roleAssignmentScope": "/subscriptions/00000000-0000-0000-0000-000000000000",
				"roleAssignmentId": "00000000000000000000000000000000",
				"roleDefinitionId": "00000000000000000000000000000000",
				"principalId": "00000000000000000000000000000000",
				"principalType": "ServicePrincipal"
			}
		}
	},
	"level": "Information",
	"properties": {
		"eventCategory": "Administrative",
		"entity": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Web/sites/function",
		"message": "Microsoft.Web/sites/write",
		"hierarchy": "00000000-0000-0000-0000-000000000000/00000000-0000-0000-0000-000000000000"
	},
	"tenantId": "00000000-0000-0000-0000-000000000000"
}`

const unknownActivityLog = `{
	"resourceId": "/SUBSCRIPTIONS/00000000-0000-0000-0000-000000000000/RESOURCEGROUPS/GROUP/PROVIDERS/MICROSOFT.UNKNOWN/RESURCES/NAME",
	"operationName": "MICROSOFT.UNKNOWN/RESURCES/WRITE",
	"category": "Administrative",
	"resultType": "Succeeded",
	"resultSignature": "Succeeded.",
	"identity": {
		"authorization": {
			"scope": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Unknown/resurces/name",
			"action": "Microsoft.Unknown/resurces/write",
			"evidence": {
				"role": "Contributor",
				"roleAssignmentScope": "/subscriptions/00000000-0000-0000-0000-000000000000",
				"roleAssignmentId": "00000000000000000000000000000000",
				"roleDefinitionId": "00000000000000000000000000000000",
				"principalId": "00000000000000000000000000000000",
				"principalType": "ServicePrincipal"
			}
		}
	},
	"level": "Information",
	"properties": {
		"eventCategory": "Administrative",
		"entity": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/group/providers/Microsoft.Unknown/resurces/name",
		"message": "Microsoft.Unknown/resurces/write",
		"hierarchy": "00000000-0000-0000-0000-000000000000/00000000-0000-0000-0000-000000000000"
	},
	"tenantId": "00000000-0000-0000-0000-000000000000"
}`

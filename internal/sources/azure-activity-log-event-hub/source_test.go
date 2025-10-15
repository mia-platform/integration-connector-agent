// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package azureactivitylogeventhub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPipelineGroup struct {
	Messages []*entities.Event
}

func (t *testPipelineGroup) AddMessage(message entities.PipelineEvent) {
	if event, ok := message.(*entities.Event); ok {
		t.Messages = append(t.Messages, event)
	}
}
func (t *testPipelineGroup) Start(_ context.Context)       {}
func (t *testPipelineGroup) Close(_ context.Context) error { return nil }

func TestConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		config        config.GenericConfig
		expectedError string
	}{
		"valid config": {
			config: config.GenericConfig{
				Raw: []byte(`{
	"subscriptionId":"subscriptionId",
	"eventHubNamespace":"eventHubNamespace",
	"eventHubName":"eventHubName",
	"checkpointStorageAccountName":"checkpointStorageAccountName",
	"checkpointStorageContainerName":"checkpointStorageContainerName"
}`),
			},
		},
		"missing eventHubName": {
			config: config.GenericConfig{
				Raw: []byte(`{
	"subscriptionId":"subscriptionId",
	"eventHubNamespace":"eventHubNamespace",
	"checkpointStorageAccountName":"checkpointStorageAccountName",
	"checkpointStorageContainerName":"checkpointStorageContainerName"
}`),
			},
			expectedError: "eventHubName is required",
		},
		"missing eventHubNamespace": {
			config: config.GenericConfig{
				Raw: []byte(`{
	"subscriptionId":"subscriptionId",
	"eventHubName":"eventHubName",
	"checkpointStorageAccountName":"checkpointStorageAccountName",
	"checkpointStorageContainerName":"checkpointStorageContainerName"
}`),
			},
			expectedError: "eventHubNamespace is required",
		},
		"missing subscriptionId": {
			config: config.GenericConfig{
				Raw: []byte(`{
	"eventHubNamespace":"eventHubNamespace",
	"eventHubName":"eventHubName",
	"checkpointStorageAccountName":"checkpointStorageAccountName",
	"checkpointStorageContainerName":"checkpointStorageContainerName"
}`),
			},
			expectedError: "subscriptionId is required",
		},
		"missing checkpointStorageAccountName": {
			config: config.GenericConfig{
				Raw: []byte(`{
	"subscriptionId":"subscriptionId",
	"eventHubNamespace":"eventHubNamespace",
	"eventHubName":"eventHubName",
	"checkpointStorageContainerName":"checkpointStorageContainerName"
}`),
			},
			expectedError: "checkpointStorageAccountName is required",
		},
		"missing checkpointStorageContainerName": {
			config: config.GenericConfig{
				Raw: []byte(`{
	"subscriptionId":"subscriptionId",
	"eventHubNamespace":"eventHubNamespace",
	"eventHubName":"eventHubName",
	"checkpointStorageAccountName":"checkpointStorageAccountName"
}`),
			},
			expectedError: "checkpointStorageContainerName is required",
		},
	}

	logger, _ := test.NewNullLogger()
	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			cfg, err := configFromGeneric(test.config, &testPipelineGroup{}, logger)
			if len(test.expectedError) > 0 {
				assert.Nil(t, cfg)
				assert.ErrorContains(t, err, test.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.True(t, strings.HasSuffix(cfg.EventHubNamespace, ".servicebus.windows.net"))
			assert.True(t, strings.HasSuffix(cfg.CheckpointStorageAccountName, ".blob.core.windows.net"))
			assert.NotNil(t, cfg.EventConsumer)
		})
	}
}

const listActionEvent = `{
	"resourceId": "/SUBSCRIPTIONS/00000000-0000-0000-0000-000000000000/RESOURCEGROUPS/GROUP/PROVIDERS/MICROSOFT.STORAGE/STORAGEACCOUNTS/ACCOUNT",
	"operationName": "MICROSOFT.STORAGE/STORAGEACCOUNTS/LISTKEYS/ACTION",
	"category": "Administrative",
	"resultType": "Start",
	"resultSignature": "Started.",
	"identity": {
		"authorization": {
			"scope": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mia-demo/providers/Microsoft.Storage/storageAccounts/miademo",
			"action": "Microsoft.Storage/storageAccounts/listKeys/action",
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
		"message": "Microsoft.Storage/storageAccounts/listKeys/action",
		"hierarchy": "00000000-0000-0000-0000-000000000000/00000000-0000-0000-0000-000000000000"
	},
	"tenantId": "00000000-0000-0000-0000-000000000000"
}`

const deleteEvent = `{
	"resourceId": "/SUBSCRIPTIONS/00000000-0000-0000-0000-000000000000/RESOURCEGROUPS/GROUP/PROVIDERS/MICROSOFT.COMPUTE/VIRTUALMACHINESCALESETS/SCALESET",
	"operationName": "MICROSOFT.COMPUTE/VIRTUALMACHINESCALESETS/DELETE/ACTION",
	"category": "Administrative",
	"resultType": "Start",
	"resultSignature": "Started.",
	"identity": {
		"authorization": {
			"scope": "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/group/providers/microsoft.compute/virtualmachinescalesets/scaleset",
			"action": "Microsoft.Compute/virtualMachineScaleSets/delete/action",
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
		"entity": "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/group/providers/microsoft.compute/virtualmachinescalesets/scaleset",
		"message": "Microsoft.Compute/virtualMachineScaleSets/delete/action",
		"hierarchy": "00000000-0000-0000-0000-000000000000/00000000-0000-0000-0000-000000000000"
	},
	"tenantId": "00000000-0000-0000-0000-000000000000"
}`

func TestActivityLogConsumer(t *testing.T) {
	t.Parallel()

	rawEvent := func(stringEvent string) []byte {
		record := new(azure.ActivityLogEventRecord)
		err := json.Unmarshal([]byte(stringEvent), &record)
		require.NoError(t, err)

		data, err := json.Marshal(record)
		require.NoError(t, err)
		return data
	}

	testCases := map[string]struct {
		eventData        *azeventhubs.ReceivedEventData
		expectedMessages []*entities.Event
		expectedError    bool
	}{
		"return write event": {
			eventData: &azeventhubs.ReceivedEventData{
				EventData: azeventhubs.EventData{
					Body: []byte(fmt.Sprintf("{\"records\":[%s]}", listActionEvent)),
				},
			},
			expectedMessages: []*entities.Event{
				{
					PrimaryKeys: []entities.PkField{
						{
							Key:   "resourceId",
							Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/group/providers/microsoft.storage/storageaccounts/account",
						},
					},
					Type:          azure.EventTypeRecordFromEventHub.String(),
					OperationType: entities.Write,
					OriginalRaw:   rawEvent(listActionEvent),
				},
			},
		},
		"return delete event": {
			eventData: &azeventhubs.ReceivedEventData{
				EventData: azeventhubs.EventData{
					Body: []byte(fmt.Sprintf("{\"records\":[%s]}", deleteEvent)),
				},
			},
			expectedMessages: []*entities.Event{
				{
					PrimaryKeys: []entities.PkField{
						{
							Key:   "resourceId",
							Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/group/providers/microsoft.compute/virtualmachinescalesets/scaleset",
						},
					},
					Type:          azure.EventTypeRecordFromEventHub.String(),
					OperationType: entities.Delete,
					OriginalRaw:   rawEvent(deleteEvent),
				},
			},
		},
		"return multiple events": {
			eventData: &azeventhubs.ReceivedEventData{
				EventData: azeventhubs.EventData{
					Body: []byte(fmt.Sprintf("{\"records\":[%s, %s]}", listActionEvent, deleteEvent)),
				},
			},
			expectedMessages: []*entities.Event{
				{
					PrimaryKeys: []entities.PkField{
						{
							Key:   "resourceId",
							Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/group/providers/microsoft.storage/storageaccounts/account",
						},
					},
					Type:          azure.EventTypeRecordFromEventHub.String(),
					OperationType: entities.Write,
					OriginalRaw:   rawEvent(listActionEvent),
				},
				{
					PrimaryKeys: []entities.PkField{
						{
							Key:   "resourceId",
							Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/group/providers/microsoft.compute/virtualmachinescalesets/scaleset",
						},
					},
					Type:          azure.EventTypeRecordFromEventHub.String(),
					OperationType: entities.Delete,
					OriginalRaw:   rawEvent(deleteEvent),
				},
			},
		},
		"message is unparsable": {
			eventData: &azeventhubs.ReceivedEventData{
				EventData: azeventhubs.EventData{
					Body: []byte(`{"records":[{"":""}}`),
				},
			},
		},
	}

	log, _ := test.NewNullLogger()
	for testName, test := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			pg := &testPipelineGroup{}
			consumerFunction := activityLogConsumer(pg, log)

			err := consumerFunction(test.eventData)
			if test.expectedError {
				assert.Error(t, err)
				assert.Nil(t, pg.Messages)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.expectedMessages, pg.Messages)
		})
	}
}

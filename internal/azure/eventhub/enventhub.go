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

package eventhub

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2/checkpoints"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/mia-platform/integration-connector-agent/internal/azure"
	"github.com/sirupsen/logrus"
)

func NewConsumerClient(config azure.EventHubConfig, consumerGroup string) (*azeventhubs.ConsumerClient, error) {
	credentials, err := config.AzureTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credentials for Event Hub access: %w", err)
	}

	namespace := config.EventHubNamespace
	name := config.EventHubName
	client, err := azeventhubs.NewConsumerClient(namespace, name, consumerGroup, credentials, nil)
	if err != nil {
		return nil, enhanceEventHubError(err, namespace, name, consumerGroup)
	}

	return client, nil
}

// enhanceEventHubError provides detailed guidance for Event Hub access errors
func enhanceEventHubError(err error, namespace, hubName, consumerGroup string) error {
	baseErr := fmt.Errorf("failed to create Event Hub consumer client for '%s/%s' with consumer group '%s': %w", namespace, hubName, consumerGroup, err)

	errStr := err.Error()
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "Forbidden") || strings.Contains(errStr, "authorization") {
		return fmt.Errorf("%w\n\nAzure Event Hub Permission Issue: Service principal lacks required permissions.\nRequired permissions for Event Hub namespace '%s':\n  - 'Azure Event Hubs Data Receiver' role\n\nTo fix this:\n1. Go to Azure Portal → Event Hubs → %s → Access Control (IAM)\n2. Click 'Add role assignment'\n3. Select 'Azure Event Hubs Data Receiver' role\n4. Assign to your service principal/managed identity\n\nFor full permissions, you may also need:\n  - 'Azure Event Hubs Data Owner' role (for administrative operations)", baseErr, namespace, namespace)
	}

	if strings.Contains(errStr, "401") || strings.Contains(errStr, "authentication") {
		return fmt.Errorf("%w\n\nAzure Event Hub Authentication Issue: Failed to authenticate with Event Hub.\nPlease verify:\n1. Azure credentials (AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET) are correct\n2. Service principal exists and has not expired\n3. Service principal has access to the Event Hub namespace '%s'", baseErr, namespace)
	}

	if strings.Contains(errStr, "404") || strings.Contains(errStr, "not found") {
		return fmt.Errorf("%w\n\nAzure Event Hub Resource Issue: Event Hub not found.\nPlease verify:\n1. Event Hub namespace '%s' exists\n2. Event Hub '%s' exists within the namespace\n3. Namespace URL is correct (should end with .servicebus.windows.net)\n4. Your service principal has access to the subscription/resource group", baseErr, namespace, hubName)
	}

	return baseErr
}

func RunEventHubProcessor(ctx context.Context, config azure.EventHubConfig, client *azeventhubs.ConsumerClient, logger *logrus.Logger) error {
	store, err := newCheckpointStore(config)
	if err != nil {
		return err
	}

	processor, err := azeventhubs.NewProcessor(client, store, nil)
	if err != nil {
		return err
	}

	go dispatchPartitionClients(ctx, processor, config.EventConsumer, logger)
	return processor.Run(ctx)
}

func newCheckpointStore(config azure.EventHubConfig) (azeventhubs.CheckpointStore, error) {
	credentials, err := config.AzureTokenProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credentials for checkpoint storage: %w", err)
	}

	accountName := config.CheckpointStorageAccountName
	containerName := config.CheckpointStorageContainerName
	storageAccountClient, err := azblob.NewClient(accountName, credentials, nil)
	if err != nil {
		return nil, enhanceStorageClientError(err, accountName)
	}

	store, err := checkpoints.NewBlobStore(storageAccountClient.ServiceClient().NewContainerClient(containerName), nil)
	if err != nil {
		return nil, enhanceCheckpointStoreError(err, accountName, containerName)
	}

	return store, nil
}

// enhanceStorageClientError provides detailed guidance for Azure storage client creation errors
func enhanceStorageClientError(err error, storageAccount string) error {
	baseErr := fmt.Errorf("failed to create Azure storage client for account '%s': %w", storageAccount, err)

	// Check for common error patterns
	errStr := err.Error()
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "AuthorizationPermissionMismatch") {
		return fmt.Errorf("%w\n\nAzure Configuration Issue: The service principal or managed identity lacks required permissions.\nRequired permissions for storage account '%s':\n  - 'Storage Blob Data Contributor' role (recommended)\n  - OR both 'Storage Blob Data Reader' + 'Storage Blob Data Writer' roles\n\nTo fix this:\n1. Go to Azure Portal → Storage Accounts → %s → Access Control (IAM)\n2. Click 'Add role assignment'\n3. Select 'Storage Blob Data Contributor' role\n4. Assign to your service principal/managed identity\n\nAlternatively, ensure your service principal has the correct credentials:\n- AZURE_TENANT_ID: %s\n- AZURE_CLIENT_ID: %s\n- AZURE_CLIENT_SECRET: (configured securely)", baseErr, storageAccount, storageAccount, "<your-tenant-id>", "<your-client-id>")
	}

	if strings.Contains(errStr, "401") || strings.Contains(errStr, "authentication") || strings.Contains(errStr, "credential") {
		return fmt.Errorf("%w\n\nAzure Authentication Issue: Failed to authenticate with Azure.\nPlease verify your Azure credentials:\n1. Check that AZURE_TENANT_ID, AZURE_CLIENT_ID, and AZURE_CLIENT_SECRET are correctly set\n2. Ensure the service principal exists and is active\n3. Verify the client secret has not expired\n4. Confirm the service principal has access to the subscription", baseErr)
	}

	if strings.Contains(errStr, "404") || strings.Contains(errStr, "not found") {
		return fmt.Errorf("%w\n\nAzure Resource Issue: Storage account not found.\nPlease verify:\n1. Storage account '%s' exists\n2. Storage account name is correct (should be full URL: https://accountname.blob.core.windows.net)\n3. Your service principal has access to the subscription containing this storage account", baseErr, storageAccount)
	}

	return baseErr
}

// enhanceCheckpointStoreError provides detailed guidance for checkpoint store creation errors
func enhanceCheckpointStoreError(err error, storageAccount, containerName string) error {
	baseErr := fmt.Errorf("failed to create checkpoint store for container '%s' in storage account '%s': %w", containerName, storageAccount, err)

	errStr := err.Error()
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "AuthorizationPermissionMismatch") {
		return fmt.Errorf("%w\n\nAzure Blob Container Permission Issue: Cannot access container '%s'.\nRequired permissions:\n  - 'Storage Blob Data Contributor' role on storage account '%s'\n  - OR both 'Storage Blob Data Reader' + 'Storage Blob Data Writer' roles\n\nAdditionally verify:\n1. Container '%s' exists in storage account '%s'\n2. Container is accessible (not in a restricted network)\n3. Service principal has the required RBAC roles assigned", baseErr, containerName, storageAccount, containerName, storageAccount)
	}

	if strings.Contains(errStr, "404") || strings.Contains(errStr, "not found") {
		return fmt.Errorf("%w\n\nAzure Container Issue: Container '%s' not found in storage account '%s'.\nTo fix this:\n1. Create the container in Azure Portal or via Azure CLI:\n   az storage container create --name %s --account-name <account-name>\n2. Ensure the container name is correct\n3. Verify your service principal has access to the storage account", baseErr, containerName, storageAccount, containerName)
	}

	return baseErr
}

// enhancePartitionProcessingError provides detailed guidance for partition processing errors
func enhancePartitionProcessingError(err error) error {
	errStr := err.Error()

	// Check for common runtime errors
	if strings.Contains(errStr, "403") || strings.Contains(errStr, "AuthorizationPermissionMismatch") {
		return fmt.Errorf("azure permission error during event processing: %w\n\nThis typically indicates issues with:\n1. Storage Blob permissions for checkpoint storage (need 'Storage Blob Data Contributor')\n2. Event Hub permissions (need 'Azure Event Hubs Data Receiver')\n\nPlease verify all required permissions are correctly assigned to your service principal", err)
	}

	if strings.Contains(errStr, "401") || strings.Contains(errStr, "authentication") {
		return fmt.Errorf("azure authentication error during event processing: %w\n\nThis may indicate:\n1. Expired service principal credentials\n2. Token refresh issues\n3. Network connectivity problems\n\nCheck your Azure credentials and network connectivity", err)
	}

	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline") {
		return fmt.Errorf("azure Event Hub timeout error: %w\n\nThis may indicate:\n1. Network connectivity issues\n2. Event Hub throttling\n3. Resource scaling issues\n\nConsider checking:\n- Network connectivity to Azure\n- Event Hub throughput units\n- Processing performance", err)
	}

	return err
}

func dispatchPartitionClients(ctx context.Context, processor *azeventhubs.Processor, consumer azure.EventConsumer, logger *logrus.Logger) {
	for {
		var processorPartitionClient *azeventhubs.ProcessorPartitionClient
		if processorPartitionClient = processor.NextPartitionClient(ctx); processorPartitionClient == nil {
			break
		}

		go func() {
			if err := processEventsForPartition(ctx, processorPartitionClient, consumer); err != nil {
				enhancedErr := enhancePartitionProcessingError(err)
				logger.WithError(enhancedErr).Error("failed to process events for partitions")
			}
		}()
	}
}

func processEventsForPartition(ctx context.Context, partitionClient *azeventhubs.ProcessorPartitionClient, consumer azure.EventConsumer) error {
	defer partitionClient.Close(ctx)

	for {
		receiveCtx, cancelReceive := context.WithTimeout(ctx, 30*time.Second)
		events, err := partitionClient.ReceiveEvents(receiveCtx, 10, nil)
		cancelReceive()

		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			var eventHubError *azeventhubs.Error

			if errors.As(err, &eventHubError) && eventHubError.Code == azeventhubs.ErrorCodeOwnershipLost {
				return nil
			}

			return err
		}

		for _, event := range events {
			err := consumer(event)
			if err != nil {
				return err
			}

			if err := partitionClient.UpdateCheckpoint(ctx, event, nil); err != nil {
				return err
			}
		}

		if len(events) == 0 {
			continue
		}
	}
}

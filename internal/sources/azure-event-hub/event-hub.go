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

package azureeventhub

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2/checkpoints"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/sirupsen/logrus"
)

type EventConsumer func(event *azeventhubs.ReceivedEventData) error

func SetupEventHub(ctx context.Context, config *Config, logger *logrus.Logger) error {
	var credential *azidentity.DefaultAzureCredential
	var store azeventhubs.CheckpointStore
	var consumerClient *azeventhubs.ConsumerClient
	var processor *azeventhubs.Processor
	var err error

	if credential, err = azureCredential(); err != nil {
		return err
	}

	if store, err = checkpointStore(credential, config.CheckpointStorageAccountName, config.CheckpointStorageContainerName); err != nil {
		return err
	}

	if consumerClient, err = azeventhubs.NewConsumerClient(config.EventHubNamespace, config.EventHubName, azeventhubs.DefaultConsumerGroup, credential, nil); err != nil {
		return err
	}
	defer consumerClient.Close(ctx)

	if processor, err = azeventhubs.NewProcessor(consumerClient, store, nil); err != nil {
		return err
	}

	go dispatchPartitionClients(ctx, processor, config.EventConsumer, logger)

	processorCtx, processorCancel := context.WithCancel(ctx)
	defer processorCancel()

	return processor.Run(processorCtx)
}

func azureCredential() (*azidentity.DefaultAzureCredential, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}
	return credential, nil
}

func checkpointStore(credential *azidentity.DefaultAzureCredential, storageAccountName, storageContainerName string) (azeventhubs.CheckpointStore, error) {
	var blobClient *azblob.Client
	var err error

	if blobClient, err = azblob.NewClient(storageAccountName, credential, nil); err != nil {
		return nil, err
	}

	azBlobContainerClient := blobClient.ServiceClient().NewContainerClient(storageContainerName)
	return checkpoints.NewBlobStore(azBlobContainerClient, nil)
}

func dispatchPartitionClients(ctx context.Context, processor *azeventhubs.Processor, consumer EventConsumer, logger *logrus.Logger) {
	for {
		var processorPartitionClient *azeventhubs.ProcessorPartitionClient
		if processorPartitionClient = processor.NextPartitionClient(ctx); processorPartitionClient == nil {
			break
		}

		go func() {
			if err := processEventsForPartition(ctx, processorPartitionClient, consumer); err != nil {
				logger.WithError(err).Error("failed to process events for partitions")
			}
		}()
	}
}

func processEventsForPartition(ctx context.Context, partitionClient *azeventhubs.ProcessorPartitionClient, consumer EventConsumer) error {
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

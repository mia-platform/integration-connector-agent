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
		return nil, err
	}

	namespace := config.EventHubNamespace
	name := config.EventHubName
	return azeventhubs.NewConsumerClient(namespace, name, consumerGroup, credentials, nil)
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
		return nil, err
	}

	accountName := config.CheckpointStorageAccountName
	containerName := config.CheckpointStorageContainerName
	storageAccountClient, err := azblob.NewClient(accountName, credentials, nil)
	if err != nil {
		return nil, err
	}

	return checkpoints.NewBlobStore(storageAccountClient.ServiceClient().NewContainerClient(containerName), nil)
}

func dispatchPartitionClients(ctx context.Context, processor *azeventhubs.Processor, consumer azure.EventConsumer, logger *logrus.Logger) {
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

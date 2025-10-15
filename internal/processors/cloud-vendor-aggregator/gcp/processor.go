// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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

package gcpaggregator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp/clients/runservice"
	storageclient "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp/clients/storage"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp/services/service"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp/services/storage"
	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"

	"github.com/sirupsen/logrus"
	gcpOptions "google.golang.org/api/option"
)

type GCPCloudVendorAggregator struct {
	logger  *logrus.Logger
	options gcpOptions.ClientOption

	s storageclient.Client
	f runservice.Client
}

func New(logger *logrus.Logger, authOptions config.AuthOptions) (entities.Processor, error) {
	options := gcpOptions.WithCredentialsJSON([]byte(authOptions.CredenialsJSON.String()))

	storageClient, err := storageclient.NewClient(context.Background(), options)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP storage client: %w", err)
	}

	runServiceClient, err := runservice.NewClient(context.Background(), options)
	if err != nil {
		return nil, err
	}

	return &GCPCloudVendorAggregator{
		logger:  logger,
		options: options,

		s: storageClient,
		f: runServiceClient,
	}, nil
}

func (c *GCPCloudVendorAggregator) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	output := input.Clone()

	if input.Operation() == entities.Delete {
		c.logger.Debug("Delete operation detected, skipping processing")
		return output, nil
	}

	assetInventory, err := c.ParseEvent(input)
	if err != nil {
		c.logger.WithError(err).Error("Failed to parse event")
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	processor, err := c.EventDataProcessor(assetInventory)
	if err != nil {
		c.logger.WithError(err).Error("Failed to process event data")
		return nil, fmt.Errorf("failed to process event data: %w", err)
	}

	newData, err := processor.GetData(context.Background(), assetInventory)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get data from processor")
		return nil, fmt.Errorf("failed to get data from processor: %w", err)
	}

	output.WithData(newData)
	return output, nil
}

func (c *GCPCloudVendorAggregator) EventDataProcessor(event gcppubsubevents.IInventoryEvent) (commons.DataAdapter[gcppubsubevents.IInventoryEvent], error) {
	assetType := event.AssetType()

	switch assetType {
	case storage.StorageAssetType:
		return storage.NewGCPRunServiceDataAdapter(c.s), nil
	case service.RunServiceAssetType:
		return service.NewGCPRunServiceDataAdapter(c.f), nil
	default:
		return nil, fmt.Errorf("unsupported asset type: %s", assetType)
	}
}

func (c *GCPCloudVendorAggregator) ParseEvent(event entities.PipelineEvent) (gcppubsubevents.IInventoryEvent, error) {
	eventType := event.GetType()

	switch eventType {
	case gcppubsubevents.ImportEventType:

		var assetInventory gcppubsubevents.InventoryImportEvent
		if err := json.Unmarshal(event.Data(), &assetInventory); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input data: %w", err)
		}
		return assetInventory, nil

	case gcppubsubevents.RealtimeSyncEventType:

		var assetInventory gcppubsubevents.InventoryEvent
		if err := json.Unmarshal(event.Data(), &assetInventory); err != nil {
			return nil, fmt.Errorf("failed to unmarshal input data: %w", err)
		}
		return assetInventory, nil

	default:
		return nil, fmt.Errorf("unsupported event type: %s", eventType)
	}
}

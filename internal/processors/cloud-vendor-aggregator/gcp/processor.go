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
}

func New(logger *logrus.Logger, authOptions config.AuthOptions) entities.Processor {
	return &GCPCloudVendorAggregator{
		logger:  logger,
		options: gcpOptions.WithCredentialsJSON([]byte(authOptions.CredenialsJSON.String())),
	}
}

func (c *GCPCloudVendorAggregator) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	assetInventory := new(gcppubsubevents.InventoryEvent)

	if err := json.Unmarshal(input.Data(), &assetInventory); err != nil {
		return nil, fmt.Errorf("failed to unmarshal input data: %w", err)
	}

	output := input.Clone()

	if assetInventory.Deleted {
		c.logger.Debug("Delete operation detected, skipping processing")
		return output, nil
	}

	processor, client, err := c.EventDataProcessor(assetInventory)
	if err != nil {
		c.logger.WithError(err).Error("Failed to process event data")
		return nil, fmt.Errorf("failed to process event data: %w", err)
	}

	defer client.Close()

	newData, err := processor.GetData(context.Background(), assetInventory)
	if err != nil {
		c.logger.WithError(err).Error("Failed to get data from processor")
		return nil, fmt.Errorf("failed to get data from processor: %w", err)
	}

	output.WithData(newData)
	return output, nil
}

func (c *GCPCloudVendorAggregator) EventDataProcessor(event *gcppubsubevents.InventoryEvent) (commons.DataAdapter[*gcppubsubevents.InventoryEvent], commons.Closable, error) {
	assetType := event.Asset.AssetType
	switch assetType {
	case storage.StorageAssetType:
		client, err := storageclient.NewClient(context.Background(), c.options)
		if err != nil {
			return nil, nil, err
		}
		return storage.NewGCPRunServiceDataAdapter(client), client, nil
	case service.RunServiceAssetType:
		client, err := runservice.NewClient(context.Background(), c.options)
		if err != nil {
			return nil, nil, err
		}
		return service.NewGCPRunServiceDataAdapter(client), client, nil
	default:
		return nil, nil, fmt.Errorf("unsupported asset type: %s", assetType)
	}
}

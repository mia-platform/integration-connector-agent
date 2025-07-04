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
		options: gcpOptions.WithCredentialsJSON([]byte(authOptions.CredenialsJson.String())),
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
		return storage.NewGCPRunServiceDataAdapter(context.Background(), client), client, nil
	case service.RunServiceAssetType:
		client, err := runservice.NewClient(context.Background(), c.options)
		if err != nil {
			return nil, nil, err
		}
		return service.NewGCPRunServiceDataAdapter(context.Background(), client), client, nil
	default:
		return nil, nil, fmt.Errorf("unsupported asset type: %s", assetType)
	}
}

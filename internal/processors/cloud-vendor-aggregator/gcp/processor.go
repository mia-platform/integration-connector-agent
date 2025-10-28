// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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
	if input.GetType() == gcppubsubevents.ImportEventType {
		fmt.Println("gcppubsubevents.ImportEventType")
		return output, nil
	}

	if input.Operation() == entities.Delete {
		c.logger.Debug("Delete operation detected, skipping processing")
		return output, nil
	}

	asset, assetType, err := getAssetInventoryEvent(input.Data())
	if err != nil {
		return output, err
	}
	output.WithData(asset)

	return &entities.Event{
		PrimaryKeys:   output.GetPrimaryKeys(),
		Type:          assetType,
		OperationType: output.Operation(),
		OriginalRaw:   output.Data(),
	}, nil
}

func getAssetInventoryEvent(rawData []byte) ([]byte, string, error) {
	newRawData := new(gcppubsubevents.InventoryEvent)
	if err := json.Unmarshal(rawData, &newRawData); err != nil {
		fmt.Println("failed to unmarshal raw data", err)
		return nil, "", err
	}
	newByteRawData, err := json.Marshal(newRawData.Asset)
	if err != nil {
		fmt.Println("failed to marshal raw data", err)
		return nil, "", err
	}
	return newByteRawData, newRawData.AssetType(), nil
}

func logRawDataInventoryEvent(rawData []byte) {
	newRawData := new(gcppubsubevents.InventoryEvent)
	if err := json.Unmarshal(rawData, &newRawData); err != nil {
		fmt.Println("failed to unmarshal raw data", err)
		return
	}

	pretty, err := json.MarshalIndent(newRawData.Asset, "", "  ")
	if err != nil {
		fmt.Println("failed to marshal raw data", err)
		return
	}

	fmt.Println("logRawDataInventoryEvent:" + string(pretty))
}

func logRawDataInventoryImportEvent(rawData []byte) {
	newRawData := new(gcppubsubevents.InventoryImportEvent)
	if err := json.Unmarshal(rawData, &newRawData); err != nil {
		fmt.Println("failed to unmarshal raw data", err)
		return
	}

	pretty, err := json.MarshalIndent(newRawData, "", "  ")
	if err != nil {
		fmt.Println("failed to marshal raw data", err)
		return
	}

	fmt.Println("logRawDataInventoryImportEvent: " + string(pretty))
}

func logRawData(newDataBytes []byte) {
	newData := new(commons.Asset)
	if err := json.Unmarshal(newDataBytes, &newData); err != nil {
		fmt.Println("failed to unmarshal raw data", err)
		return
	}

	newRawData := new(gcppubsubevents.InventoryImportEvent)
	if err := json.Unmarshal(newData.RawData, &newRawData); err != nil {
		fmt.Println("failed to unmarshal raw data", err)
		return
	}

	pretty, err := json.MarshalIndent(newRawData, "", "  ")
	if err != nil {
		fmt.Println("failed to marshal raw data", err)
		return
	}

	fmt.Println("Raw data:\n" + string(pretty))
}

func logEventData(event entities.PipelineEvent) error {
	// Try to get the JSON representation of the event for debugging and
	// pretty-print it with indentation so logs are easier to read.
	if jsonData, err := event.JSON(); err != nil {
		return fmt.Errorf("failed to get event JSON: %w", err)
	} else {
		if pretty, err := json.MarshalIndent(jsonData, "", "  "); err != nil {
			// Fallback to logging the raw parsed map if MarshalIndent fails
			return fmt.Errorf("failed to marshal event JSON: %w", err)
		} else {
			fmt.Println("Event data:\n" + string(pretty))
		}
	}
	return nil
}

// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package gcpaggregator

import (
	"encoding/json"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	gcppubsubevents "github.com/mia-platform/integration-connector-agent/internal/sources/gcp-pubsub/events"

	"github.com/sirupsen/logrus"
	gcpOptions "google.golang.org/api/option"
)

type GCPCloudVendorAggregator struct {
	logger  *logrus.Logger
	options gcpOptions.ClientOption
}

func New(logger *logrus.Logger, authOptions config.AuthOptions) (entities.Processor, error) {
	options := gcpOptions.WithCredentialsJSON([]byte(authOptions.CredenialsJSON.String()))

	return &GCPCloudVendorAggregator{
		logger:  logger,
		options: options,
	}, nil
}

func (c *GCPCloudVendorAggregator) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	output := input.Clone()
	if input.GetType() != gcppubsubevents.RealtimeSyncEventType {
		c.logger.Debug("Non-RealtimeSyncEventType detected")
		asset, assetType, err := getAssetInventoryImportEvent(input.Data(), c.logger)
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

	c.logger.Debug("RealtimeSyncEventType detected")

	if input.Operation() == entities.Delete {
		c.logger.Debug("Delete operation detected, skipping processing")
		return output, nil
	}

	asset, assetType, err := getAssetInventoryEvent(input.Data(), c.logger)
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

func getAssetInventoryEvent(rawData []byte, logger *logrus.Logger) ([]byte, string, error) {
	newRawData := new(gcppubsubevents.InventoryEvent)
	if err := json.Unmarshal(rawData, &newRawData); err != nil {
		logger.WithError(err).Error("failed to unmarshal raw data")
		return nil, "", err
	}
	newByteRawData, err := json.Marshal(newRawData.Asset)
	if err != nil {
		logger.WithError(err).Error("failed to marshal raw data")
		return nil, "", err
	}
	return newByteRawData, newRawData.AssetType(), nil
}

func getAssetInventoryImportEvent(rawData []byte, logger *logrus.Logger) ([]byte, string, error) {
	newRawData := new(gcppubsubevents.InventoryImportEvent)
	if err := json.Unmarshal(rawData, &newRawData); err != nil {
		logger.WithError(err).Error("failed to unmarshal raw data")
		return nil, "", err
	}
	newByteRawData, err := json.Marshal(newRawData.Data)
	if err != nil {
		logger.WithError(err).Error("failed to marshal raw data")
		return nil, "", err
	}
	return newByteRawData, newRawData.AssetType(), nil
}

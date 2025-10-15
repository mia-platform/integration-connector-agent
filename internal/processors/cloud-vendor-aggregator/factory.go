// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package cloudvendoraggregator

import (
	"github.com/mia-platform/integration-connector-agent/entities"
	awsaggergator "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws"
	azureaggregator "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/azure"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/commons"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	gcpaggregator "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp"

	"github.com/sirupsen/logrus"
)

type CloudVendorAggregator struct {
}

func New(logger *logrus.Logger, cfg config.Config) (entities.Processor, error) {
	switch cfg.CloudVendorName {
	case commons.GCPAssetProvider:
		return gcpaggregator.New(logger, cfg.AuthOptions)
	case commons.AWSAssetProvider:
		return awsaggergator.New(logger, cfg.AuthOptions), nil
	case commons.AzureAssetProvider:
		return azureaggregator.New(logger, cfg.AuthOptions)
	default:
		return nil, config.ErrInvalidCloudVendor
	}
}

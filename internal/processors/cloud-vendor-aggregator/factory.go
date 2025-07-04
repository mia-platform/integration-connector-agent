package cloudvendoraggregator

import (
	"github.com/mia-platform/integration-connector-agent/entities"
	awsaggergator "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/aws"
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
		return gcpaggregator.New(logger, cfg.AuthOptions), nil
	case commons.AWSAssetProvider:
		return awsaggergator.New(logger, cfg.AuthOptions), nil
	default:
		return nil, config.ErrInvalidCloudVendor
	}
}

package cloudvendoraggregator

import (
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	gcpaggregator "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/gcp"

	"github.com/sirupsen/logrus"
)

type CloudVendorAggregator struct {
}

func New(logger *logrus.Logger, cfg config.Config) (entities.Processor, error) {
	switch cfg.CloudVendorName {
	case "gcp":
		return gcpaggregator.New(logger, cfg.AuthOptions), nil
	default:
		return nil, config.ErrInvalidCloudVendor
	}
}

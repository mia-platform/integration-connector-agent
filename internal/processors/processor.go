// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package processors

import (
	"context"
	"errors"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"

	cloudvendoraggregator "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator"
	cloudvendoraggregatorConfig "github.com/mia-platform/integration-connector-agent/internal/processors/cloud-vendor-aggregator/config"
	"github.com/mia-platform/integration-connector-agent/internal/processors/filter"
	"github.com/mia-platform/integration-connector-agent/internal/processors/hcgp"
	"github.com/mia-platform/integration-connector-agent/internal/processors/mapper"

	"github.com/sirupsen/logrus"
)

var (
	ErrProcessorNotSupported = errors.New("processor not supported")
)

const (
	Mapper                = "mapper"
	Filter                = "filter"
	RPC                   = "rpc-plugin"
	CloudVendorAggregator = "cloud-vendor-aggregator"
)

type Processors struct {
	processors []entities.Processor
}

func (p *Processors) Process(ctx context.Context, message entities.PipelineEvent) (entities.PipelineEvent, error) {
	for i, processor := range p.processors {
		processorType := fmt.Sprintf("%T", processor)
		logger := logrus.WithFields(logrus.Fields{
			"processorIndex": i,
			"processorType":  processorType,
			"eventType":      message.GetType(),
			"primaryKeys":    message.GetPrimaryKeys().Map(),
		})

		logger.Debug("starting processor execution")

		var err error
		message, err = processor.Process(message)
		if err != nil {
			if errors.Is(err, entities.ErrDiscardEvent) {
				// Event filtered out - not an error, just pass it through
				logger.WithError(err).Debug("event filtered by processor")
				return nil, err
			}
			logger.WithError(err).Error("processor execution failed")
			return nil, err
		}

		logger.Debug("processor execution completed successfully")
	}

	return message, nil
}

type CloseableProcessor interface {
	Close() error
}

func (p *Processors) Close() error {
	for _, processor := range p.processors {
		if closer, ok := processor.(CloseableProcessor); ok {
			if err := closer.Close(); err != nil {
				return fmt.Errorf("error closing processor %T: %w", processor, err)
			}
		}
	}
	return nil
}

func New(logger *logrus.Logger, cfg config.Processors) (*Processors, error) {
	p := new(Processors)

	for _, processor := range cfg {
		logger.WithFields(logrus.Fields{"type": processor.Type}).Trace("initializing processor")
		switch processor.Type {
		case Mapper:
			config, err := config.GetConfig[mapper.Config](processor)
			if err != nil {
				return nil, err
			}
			m, err := mapper.New(config)
			if err != nil {
				return nil, err
			}
			p.processors = append(p.processors, m)
		case Filter:
			config, err := config.GetConfig[filter.Config](processor)
			if err != nil {
				return nil, err
			}
			f, err := filter.New(config)
			if err != nil {
				return nil, err
			}
			p.processors = append(p.processors, f)
		case RPC:
			config, err := config.GetConfig[hcgp.Config](processor)
			if err != nil {
				return nil, err
			}
			h, err := hcgp.New(logger, config)
			if err != nil {
				return nil, err
			}
			p.processors = append(p.processors, h)
		case CloudVendorAggregator:
			config, err := config.GetConfig[cloudvendoraggregatorConfig.Config](processor)
			if err != nil {
				return nil, err
			}
			if err := config.Validate(); err != nil {
				return nil, fmt.Errorf("invalid cloud vendor aggregator config: %w", err)
			}
			processor, err := cloudvendoraggregator.New(logger, config)
			if err != nil {
				return nil, fmt.Errorf("error creating cloud vendor aggregator processor: %w", err)
			}
			p.processors = append(p.processors, processor)

		default:
			return nil, ErrProcessorNotSupported
		}
	}

	return p, nil
}

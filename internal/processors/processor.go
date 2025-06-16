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

package processors

import (
	"context"
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/processors/filter"
	"github.com/mia-platform/integration-connector-agent/internal/processors/hcgp"
	"github.com/mia-platform/integration-connector-agent/internal/processors/mapper"
	"github.com/sirupsen/logrus"
)

var (
	ErrProcessorNotSupported = fmt.Errorf("processor not supported")
)

const (
	Mapper            = "mapper"
	Filter            = "filter"
	HashicorpGoPlugin = "rpc-plugin"
)

type Processors struct {
	processors []entities.Processor
}

func (p *Processors) Process(_ context.Context, message entities.PipelineEvent) (entities.PipelineEvent, error) {
	for _, processor := range p.processors {
		var err error
		message, err = processor.Process(message)
		if err != nil {
			return nil, err
		}
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
		case HashicorpGoPlugin:
			config, err := config.GetConfig[hcgp.Config](processor)
			if err != nil {
				return nil, err
			}
			h, err := hcgp.New(logger, config)
			if err != nil {
				return nil, err
			}
			p.processors = append(p.processors, h)
		default:
			return nil, ErrProcessorNotSupported
		}
	}

	return p, nil
}

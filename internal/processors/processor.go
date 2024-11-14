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
	"errors"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/processors/mapper"
)

type Processor interface {
	Process(data []byte) ([]byte, error)
}

var (
	ErrProcessorNotSupported = errors.New("processor not supported")
)

const (
	Mapper = "mapper"
)

type Processors struct {
	processors []Processor
}

func (p *Processors) Process(_ context.Context, message entities.PipelineEvent) (entities.PipelineEvent, error) {
	processedData := message.Data()
	for _, processor := range p.processors {
		var err error
		processedData, err = processor.Process(processedData)
		if err != nil {
			return nil, err
		}
	}

	message.WithData(processedData)
	return message, nil
}

func New(cfg config.Processors) (*Processors, error) {
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
		default:
			return nil, ErrProcessorNotSupported
		}
	}

	return p, nil
}

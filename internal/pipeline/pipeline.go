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

package pipeline

import (
	"context"
	"errors"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	"github.com/mia-platform/integration-connector-agent/internal/sinks"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	"github.com/sirupsen/logrus"
)

var (
	ErrWriterNotDefined = errors.New("writer not defined")
)

type Pipeline struct {
	sinks      sinks.Sink[entities.PipelineEvent]
	processors *processors.Processors
	logger     *logrus.Logger

	eventChan chan entities.PipelineEvent
}

func (p Pipeline) AddMessage(data entities.PipelineEvent) {
	p.eventChan <- data
}

func (p Pipeline) Start(ctx context.Context) error {
	if utils.IsNil(p.sinks) {
		return ErrWriterNotDefined
	}

	err := p.runPipeline(ctx)
	if err != nil {
		p.logger.WithError(err).Error("error starting pipeline")
		return err
	}

	return nil
}

func (p Pipeline) Close(ctx context.Context) error {
	if err := p.sinks.Close(ctx); err != nil {
		return err
	}

	return p.processors.Close()
}

func (p Pipeline) runPipeline(ctx context.Context) error {
loop:
	for {
		select {
		case message, open := <-p.eventChan:
			if !open {
				// the chanel has been closed, break the loop
				break loop
			}

			processedMessage, err := p.processors.Process(ctx, message)
			if err != nil {
				if errors.Is(err, entities.ErrDiscardEvent) {
					// the message has been filtered out
					p.logger.WithError(err).WithField("message", message.Data()).Trace("event filtered for pipeline")
					continue
				}
				p.logger.WithError(err).WithField("message", message.Data()).Error("error processing data")
				continue
			}

			if err := p.sinks.WriteData(ctx, processedMessage); err != nil {
				// TODO: manage failure in writing message. DLQ?
				p.logger.WithError(err).WithFields(logrus.Fields{
					"id":               processedMessage.GetPrimaryKeys().Map(),
					"data":             string(processedMessage.Data()),
					"messageOperation": processedMessage.Operation(),
				}).Error("error writing data")
			}

		case <-ctx.Done():
			// context has been cancelled close che channel
			close(p.eventChan)
			return ctx.Err()
		}
	}

	return nil
}

func New(logger *logrus.Logger, p *processors.Processors, sinks sinks.Sink[entities.PipelineEvent]) (IPipeline, error) {
	// TODO: here instead to use a buffer size it should be used a proper queue
	messageChan := make(chan entities.PipelineEvent, 1000000)

	pipeline := &Pipeline{
		sinks:      sinks,
		processors: p,

		eventChan: messageChan,
		logger:    logger,
	}

	return pipeline, nil
}

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
	"reflect"

	"github.com/mia-platform/data-connector-agent/internal/entities"
	"github.com/mia-platform/data-connector-agent/internal/mapper"
	"github.com/mia-platform/data-connector-agent/internal/writer"
	"github.com/sirupsen/logrus"
)

var (
	ErrWriterNotDefined = errors.New("writer not defined")
)

type Pipeline[T entities.PipelineEvent] struct {
	writer writer.Writer[T]
	mapper mapper.Mapper[T]
	logger *logrus.Entry

	eventChan chan T
}

func (p Pipeline[T]) AddMessage(data T) {
	p.eventChan <- data
}

func (p Pipeline[T]) Start(ctx context.Context) error {
	if isNil(p.writer) {
		return ErrWriterNotDefined
	}

loop:
	for {
		select {
		case message, open := <-p.eventChan:
			if !open {
				// the chanel has been closed, break the loop
				break loop
			}
			// _, err := p.mapper.Map(message.RawData())
			// if err != nil {
			// 	p.logger.WithError(err).WithField("message", message).Error("error mapping data")
			// 	continue
			// }
			switch message.Type() {
			case entities.Write:
				if err := p.writer.Write(ctx, message); err != nil {
					// TODO: manage failure in writing message. DLQ?
					p.logger.WithError(err).WithField("message", message).Error("error writing data")
				}
			case entities.Delete:
				if err := p.writer.Delete(ctx, message); err != nil {
					// TODO: manage failure in writing message. DLQ?
					p.logger.WithError(err).WithField("message", message).Error("error deleting data")
				}
			}

		case <-ctx.Done():
			// context has been cancelled close che channel
			close(p.eventChan)
			return ctx.Err()
		}
	}

	return nil
}

func NewPipeline[T entities.PipelineEvent](logger *logrus.Entry, writer writer.Writer[T]) IPipeline[T] {
	// TODO: here instead to use a buffer size it should be used a proper queue
	messageChan := make(chan T, 1000000)

	m := mapper.Mapper[T]{}

	return &Pipeline[T]{
		writer: writer,
		mapper: m,

		eventChan: messageChan,

		logger: logger,
	}
}

func isNil(i any) bool {
	return i == nil || (reflect.ValueOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil())
}

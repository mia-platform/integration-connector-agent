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

package fakewriter

import (
	"context"
	"sync"

	"github.com/mia-platform/integration-connector-agent/entities"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
)

type Config struct {
	Mocks Mocks
}

func (c *Config) Validate() error {
	return nil
}

type Call struct {
	Data      entities.PipelineEvent
	Operation entities.Operation
}

type Calls []Call

func (c Calls) LastCall() Call {
	if len(c) == 0 {
		return Call{}
	}
	return c[len(c)-1]
}

type Mock struct {
	Error error
}

type Mocks []Mock

func (m *Mocks) ReadAndPop() Mock {
	mock := (*m)[0]
	*m = (*m)[1:]

	return mock
}

type Writer struct {
	mtx sync.Mutex

	stub  Calls
	mocks Mocks
}

func New(config *Config) *Writer {
	if config == nil {
		config = &Config{}
	}

	w := &Writer{
		stub:  Calls{},
		mocks: config.Mocks,
	}

	if w.mocks == nil {
		w.mocks = Mocks{}
	}

	return w
}

func (f *Writer) Calls() Calls {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	return f.stub
}

func (f *Writer) ResetCalls() {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	f.stub = Calls{}
}

func (f *Writer) AddMock(mock Mock) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	f.mocks = append(f.mocks, mock)
}

func (f *Writer) WriteData(ctx context.Context, data entities.PipelineEvent) error {
	log := glogrus.FromContext(ctx)

	f.mtx.Lock()
	defer f.mtx.Unlock()

	f.stub = append(f.stub, Call{
		Data:      data,
		Operation: data.Operation(),
	})

	// Log the data being written to fake sink
	log.WithFields(map[string]interface{}{
		"operation":    data.Operation().String(),
		"primaryKeys":  data.GetPrimaryKeys(),
		"eventType":    data.GetType(),
		"dataSize":     len(data.Data()),
	}).Info("Fake sink received data")

	// Log the actual data content for debugging
	if jsonData, err := data.JSON(); err == nil {
		log.WithField("eventData", jsonData).Debug("Fake sink event data details")
	}

	if len(f.mocks) > 0 {
		mock := f.mocks.ReadAndPop()
		return mock.Error
	}
	return nil
}

func (f *Writer) Close(_ context.Context) error {
	return nil
}

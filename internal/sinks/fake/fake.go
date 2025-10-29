// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package fakewriter

import (
	"context"
	"sync"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/sirupsen/logrus"
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
	log *logrus.Logger

	stub  Calls
	mocks Mocks
}

func New(config *Config, log *logrus.Logger) *Writer {
	if config == nil {
		config = &Config{}
	}

	w := &Writer{
		stub:  Calls{},
		mocks: config.Mocks,
		log:   log,
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
	f.mtx.Lock()
	defer f.mtx.Unlock()

	f.stub = append(f.stub, Call{
		Data:      data,
		Operation: data.Operation(),
	})

	// Log the data being written to fake sink
	f.log.WithFields(map[string]interface{}{
		"operation":   data.Operation().String(),
		"primaryKeys": data.GetPrimaryKeys(),
		"eventType":   data.GetType(),
		"dataSize":    len(data.Data()),
	}).Info("Fake sink received data")

	// Log the actual data content for debugging
	if jsonData, err := data.JSON(); err == nil {
		f.log.WithField("eventData", jsonData).Debug("Fake sink event data details")
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

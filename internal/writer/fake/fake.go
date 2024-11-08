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

	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/writer"
)

type Config struct {
	OutputModel map[string]any
}

func (c *Config) Validate() error {
	return nil
}

type Call struct {
	Data      writer.DataWithIdentifier
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

	outputModel map[string]any
}

func New(config *Config) *Writer {
	return &Writer{
		stub:  Calls{},
		mocks: Mocks{},

		outputModel: config.OutputModel,
	}
}

func (f *Writer) Calls() Calls {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	return f.stub
}

func (f *Writer) AddMock(mock Mock) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	f.mocks = append(f.mocks, mock)
}

func (f *Writer) Write(_ context.Context, data entities.PipelineEvent) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	f.stub = append(f.stub, Call{
		Data:      data,
		Operation: entities.Write,
	})

	if len(f.mocks) > 0 {
		mock := f.mocks.ReadAndPop()
		return mock.Error
	}
	return nil
}

func (f *Writer) Delete(_ context.Context, data entities.PipelineEvent) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	f.stub = append(f.stub, Call{
		Data:      data,
		Operation: entities.Delete,
	})

	if len(f.mocks) > 0 {
		mock := f.mocks.ReadAndPop()
		return mock.Error
	}
	return nil
}

func (f *Writer) OutputModel() map[string]any {
	return f.outputModel
}

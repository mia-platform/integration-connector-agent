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

package fake

import (
	"context"

	"github.com/mia-platform/data-connector-agent/internal/entities"
	"github.com/mia-platform/data-connector-agent/internal/writer"
)

type Call struct {
	Data      writer.DataWithIdentifier
	Operation entities.Operation
}

type Stub struct {
	calls []Call
}

type Writer struct {
	Identifier string

	mockCalls []Stub
}

func New() writer.Writer[entities.PipelineEvent] {
	return &Writer{
		mockCalls: []Stub{},
	}
}

func (f *Writer) Write(_ context.Context, data entities.PipelineEvent) error {
	if f.mockCalls == nil {
		f.mockCalls = []Stub{}
	}

	f.mockCalls = append(f.mockCalls, Stub{
		calls: []Call{
			{
				Data:      data,
				Operation: entities.Write,
			},
		},
	})
	return nil
}

func (f *Writer) Delete(_ context.Context, data entities.PipelineEvent) error {
	if f.mockCalls == nil {
		f.mockCalls = []Stub{}
	}

	f.mockCalls = append(f.mockCalls, Stub{
		calls: []Call{
			{
				Data:      data,
				Operation: entities.Delete,
			},
		},
	})
	return nil
}

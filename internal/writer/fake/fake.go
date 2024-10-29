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

	"github.com/mia-platform/data-connector-agent/internal/writer"
)

type OperationType int

const (
	Write OperationType = iota
	Delete
)

type Call struct {
	Data      writer.DataWithIdentifier
	Operation OperationType
}

type Stub struct {
	calls []Call
}

type Fake struct {
	Identifier string

	mockCalls []Stub
}

func (f *Fake) ID() string {
	return f.Identifier
}

func New() writer.Writer[writer.DataWithIdentifier] {
	return &Fake{}
}

func (f *Fake) Write(_ context.Context, data writer.DataWithIdentifier) error {
	f.mockCalls = append(f.mockCalls, Stub{
		calls: []Call{
			{
				Data:      data,
				Operation: Write,
			},
		},
	})
	return nil
}

func (f *Fake) Delete(_ context.Context, data writer.DataWithIdentifier) error {
	f.mockCalls = append(f.mockCalls, Stub{
		calls: []Call{
			{
				Data:      data,
				Operation: Delete,
			},
		},
	})
	return nil
}

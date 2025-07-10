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

package gcppubsub

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/entities"
)

type eventBuilderMock struct {
	getPipelineEventInvoked bool
	GetPipelineEventFunc    func(ctx context.Context, data []byte) (entities.PipelineEvent, error)

	assertData    func(data []byte)
	returnedEvent *entities.Event
	returnedErr   error
}

func (e *eventBuilderMock) GetPipelineEvent(_ context.Context, data []byte) (entities.PipelineEvent, error) {
	e.getPipelineEventInvoked = true
	if e.GetPipelineEventFunc != nil {
		return e.GetPipelineEventFunc(context.Background(), data)
	}

	if e.assertData != nil {
		e.assertData(data)
	}
	return e.returnedEvent, e.returnedErr
}

type pipelineGroupMock struct {
	addMessageInvoked bool
	startInvoked      bool
	closeInvoked      bool

	assertAddMessage func(data entities.PipelineEvent)
	closeErr         error

	Messages []entities.PipelineEvent
}

func (p *pipelineGroupMock) AddMessage(data entities.PipelineEvent) {
	p.addMessageInvoked = true
	if p.assertAddMessage != nil {
		p.assertAddMessage(data)
	}
	p.Messages = append(p.Messages, data)
}

func (p *pipelineGroupMock) Start(_ context.Context) {
	p.startInvoked = true
}

func (p *pipelineGroupMock) Close() error {
	p.closeInvoked = true
	return p.closeErr
}

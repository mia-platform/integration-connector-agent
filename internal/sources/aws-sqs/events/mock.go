// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package awssqsevents

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/entities"
)

type EventBuilderMock struct {
	GetPipelineEventFunc func(ctx context.Context, data []byte) (entities.PipelineEvent, error)

	AssertData    func(data []byte)
	ReturnedEvent *entities.Event
	ReturnedErr   error
}

func (e EventBuilderMock) GetPipelineEvent(ctx context.Context, data []byte) (entities.PipelineEvent, error) {
	if e.GetPipelineEventFunc != nil {
		return e.GetPipelineEventFunc(ctx, data)
	}

	if e.AssertData != nil {
		e.AssertData(data)
	}
	return e.ReturnedEvent, e.ReturnedErr
}

type PipelineGroupMock struct {
	AddMessageInvoked bool
	StartInvoked      bool
	CloseInvoked      bool

	AssertAddMessage func(data entities.PipelineEvent)
	CloseErr         error
}

func (p *PipelineGroupMock) AddMessage(data entities.PipelineEvent) {
	p.AddMessageInvoked = true
	if p.AssertAddMessage != nil {
		p.AssertAddMessage(data)
	}
}

func (p *PipelineGroupMock) Start(_ context.Context) {
	p.StartInvoked = true
}

func (p *PipelineGroupMock) Close(_ context.Context) error {
	p.CloseInvoked = true
	return p.CloseErr
}

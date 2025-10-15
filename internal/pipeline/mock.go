// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package pipeline

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/entities"
)

type PipelineGroupMock struct {
	AddMessageInvoked bool
	StartInvoked      bool
	CloseInvoked      bool

	AssertAddMessage func(data entities.PipelineEvent)
	CloseErr         error

	Messages []entities.PipelineEvent
}

func (p *PipelineGroupMock) AddMessage(data entities.PipelineEvent) {
	p.AddMessageInvoked = true
	if p.AssertAddMessage != nil {
		p.AssertAddMessage(data)
	}
	p.Messages = append(p.Messages, data)
}

func (p *PipelineGroupMock) Start(_ context.Context) {
	p.StartInvoked = true
}

func (p *PipelineGroupMock) Close(_ context.Context) error {
	p.CloseInvoked = true
	return p.CloseErr
}

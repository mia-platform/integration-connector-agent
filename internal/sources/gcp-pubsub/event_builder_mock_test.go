// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

func (e *eventBuilderMock) GetPipelineEvent(ctx context.Context, data []byte) (entities.PipelineEvent, error) {
	e.getPipelineEventInvoked = true
	if e.GetPipelineEventFunc != nil {
		return e.GetPipelineEventFunc(ctx, data)
	}

	if e.assertData != nil {
		e.assertData(data)
	}
	return e.returnedEvent, e.returnedErr
}

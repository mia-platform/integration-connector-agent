// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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

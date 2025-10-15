// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package pipeline

import (
	"context"

	"github.com/mia-platform/integration-connector-agent/entities"
)

type IPipeline interface {
	AddMessage(data entities.PipelineEvent)
	Start(ctx context.Context) error
	Close(ctx context.Context) error
}

type IPipelineGroup interface {
	AddMessage(data entities.PipelineEvent)
	Start(ctx context.Context)
	Close(ctx context.Context) error
}

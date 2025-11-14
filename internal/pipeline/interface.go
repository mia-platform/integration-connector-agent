// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

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

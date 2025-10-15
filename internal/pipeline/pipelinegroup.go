// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/sirupsen/logrus"
)

type Group struct {
	pipelines []IPipeline
	logger    *logrus.Logger
	errors    []error
}

func NewGroup(logger *logrus.Logger, pipelines ...IPipeline) IPipelineGroup {
	return &Group{
		pipelines: pipelines,
		logger:    logger,
		errors:    make([]error, 0),
	}
}

func (pg *Group) Start(ctx context.Context) {
	for _, p := range pg.pipelines {
		go func(p IPipeline) {
			err := p.Start(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					pg.logger.WithError(err).Info("pipeline context cancelled")
					return
				}
				pg.logger.WithError(err).Error("error starting pipeline")
				// TODO: manage error
				panic(err)
			}
		}(p)
	}
}

func (pg *Group) AddMessage(event entities.PipelineEvent) {
	for _, p := range pg.pipelines {
		p.AddMessage(event.Clone())
	}
}

func (pg *Group) Close(ctx context.Context) error {
	for _, p := range pg.pipelines {
		if err := p.Close(ctx); err != nil {
			pg.errors = append(pg.errors, err)
		}
	}

	if len(pg.errors) > 0 {
		return fmt.Errorf("failed closing processors: %s", joinErrors(pg.errors))
	}

	return nil
}

func joinErrors(errors []error) string {
	var sb strings.Builder
	for i, err := range errors {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

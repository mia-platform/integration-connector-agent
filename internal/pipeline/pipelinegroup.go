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

package pipeline

import (
	"context"
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

func NewGroup(logger *logrus.Logger, pipelines ...IPipeline) *Group {
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

func (pg *Group) Close() error {
	for _, p := range pg.pipelines {
		if err := p.Close(); err != nil {
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

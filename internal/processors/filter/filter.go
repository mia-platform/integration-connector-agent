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

package filter

import (
	"fmt"

	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/utils"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/sirupsen/logrus"
)

type Filter struct {
	program    cel.Program
	expression string
}

func (m Filter) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	data, err := input.JSON()
	if err != nil {
		return nil, err
	}

	evalContext := map[string]any{
		"data":      data,
		"eventType": input.GetType(),
	}

	logrus.WithFields(logrus.Fields{
		"eventType":   input.GetType(),
		"primaryKeys": input.GetPrimaryKeys().Map(),
		"expression":  m.expression,
		"evalContext": evalContext,
	}).Debug("evaluating filter expression")

	out, _, err := m.program.Eval(evalContext)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"eventType":   input.GetType(),
			"primaryKeys": input.GetPrimaryKeys().Map(),
			"expression":  m.expression,
			"error":       err.Error(),
		}).Error("filter expression evaluation failed")
		return nil, fmt.Errorf("program evaluation failed: %s", err.Error())
	}

	filterResult := out.Equal(types.False) != types.True

	logrus.WithFields(logrus.Fields{
		"eventType":    input.GetType(),
		"primaryKeys":  input.GetPrimaryKeys().Map(),
		"expression":   m.expression,
		"filterResult": filterResult,
		"evalResult":   out.Value(),
	}).Debug("filter expression evaluation completed")

	if !filterResult {
		originalBody, decodedBody, wasDecoded := utils.TryDecodeBase64Body(input.Data())
		logFields := logrus.Fields{
			"eventType":    input.GetType(),
			"primaryKeys":  input.GetPrimaryKeys().Map(),
			"expression":   m.expression,
			"reason":       "filter_expression_false",
			"evalResult":   out.Value(),
			"originalBody": originalBody,
		}
		if wasDecoded {
			logFields["decodedBody"] = decodedBody
			logFields["wasBase64"] = true
		}
		logrus.WithFields(logFields).Debug("event skipped by filter")
		return nil, entities.ErrDiscardEvent
	}

	return input, nil
}

func New(cfg Config) (*Filter, error) {
	env, err := cel.NewEnv(
		cel.Variable("eventType", cel.StringType),
		cel.Variable("data", cel.MapType(cel.StringType, cel.AnyType)),
	)
	if err != nil {
		return nil, err
	}

	ast, iss := env.Compile(cfg.CELExpression)
	if iss.Err() != nil {
		return nil, iss.Err()
	}

	prg, err := env.Program(ast, cel.EvalOptions(cel.OptOptimize))
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"expression": cfg.CELExpression,
	}).Debug("filter processor initialized")

	return &Filter{
		program:    prg,
		expression: cfg.CELExpression,
	}, nil
}

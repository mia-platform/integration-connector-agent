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

package main

import (
	"encoding/json"

	rpcprocessor "github.com/mia-platform/integration-connector-agent/adapters/rpc-processor"
	"github.com/mia-platform/integration-connector-agent/entities"
)

type CustomProcessor struct {
	logger rpcprocessor.Logger
}

func (g *CustomProcessor) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	g.logger.WithFields(map[string]interface{}{"input": input}).Info("CustomProcessor running process for input")

	output := input.Clone()
	output.WithData([]byte(`{"data":"processed by CustomProcessor"}`))

	return output, nil
}

func (g *CustomProcessor) Init(raw []byte) error {
	// Here you can initialize your processor with the provided configuration
	// For example, you might want to set up connections, load resources, etc.
	var config map[string]interface{}
	if err := json.Unmarshal(raw, &config); err != nil {
		g.logger.WithError(err).Error("Failed to unmarshal configuration")
		return err
	}

	g.logger.WithFields(map[string]interface{}{"config": config}).Info("CustomProcessor initialized with config")
	return nil
}

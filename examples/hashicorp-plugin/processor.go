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
	"github.com/hashicorp/go-hclog"
	"github.com/mia-platform/integration-connector-agent/entities"
)

type CustomProcessor struct {
	logger hclog.Logger
}

// func (g *CustomProcessor) Process(input entities.PipelineEvent, output entities.PipelineEvent) error {
// 	g.logger.Debug("message received in CustomProcessor.Process %+v", input)
// 	return nil
// }

func (g *CustomProcessor) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	// fmt.Printf("PLUGIN PROCESSING in CustomProcessor.Process %+v\n", input)
	g.logger.Debug("[[[PLUGIN]]] message received in CustomProcessor.Process", "input", input, "inputType", input.GetType())
	output := input.Clone()
	output.WithData([]byte(`{"data":"processed by CustomProcessor"}`))
	return output, nil
}

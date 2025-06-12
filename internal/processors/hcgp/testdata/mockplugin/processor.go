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
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/mia-platform/integration-connector-agent/entities"
)

type MockProcessor struct {
	logger hclog.Logger
}

func (g *MockProcessor) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	g.logger.Trace("MockProcessor running process for input", "input", input)

	output := input.Clone()
	output.WithData([]byte(`{"data":"processed by CustomProcessor"}`))

	return output, nil
}

func (g *MockProcessor) Init(config map[string]interface{}) error {
	// Here you can initialize your processor with the provided configuration
	// For example, you might want to set up connections, load resources, etc.
	g.logger.Trace("MockProcessor initialized with config", "config", config)

	if config["fail"] != nil && config["fail"].(bool) {
		return fmt.Errorf("MockProcessor initialization failed due to configuration")
	}
	return nil
}

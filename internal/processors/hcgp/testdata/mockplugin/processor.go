// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package main

import (
	"encoding/json"
	"fmt"

	rpcprocessor "github.com/mia-platform/integration-connector-agent/adapters/rpc-processor"
	"github.com/mia-platform/integration-connector-agent/entities"
)

type MockProcessor struct {
	logger rpcprocessor.Logger
}

func (g *MockProcessor) Process(input entities.PipelineEvent) (entities.PipelineEvent, error) {
	g.logger.Trace("MockProcessor running process for input", "input", input)

	output := input.Clone()
	output.WithData([]byte(`{"data":"processed by CustomProcessor"}`))

	return output, nil
}

type ProcessorConfig struct {
	Message string `json:"message"`
	Fail    bool   `json:"fail"`
}

func (g *MockProcessor) Init(raw []byte) error {
	// Here you can initialize your processor with the provided configuration
	// For example, you might want to set up connections, load resources, etc.
	g.logger.Trace("MockProcessor initialized with config")

	var config ProcessorConfig
	if err := json.Unmarshal(raw, &config); err != nil {
		g.logger.Error("Failed to unmarshal configuration", "error", err)
		return err
	}

	g.logger.Trace("MockProcessor message ", config.Message)

	if config.Fail {
		return fmt.Errorf("MockProcessor initialization failed due to configuration")
	}
	return nil
}

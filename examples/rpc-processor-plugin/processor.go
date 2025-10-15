// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

type Config struct {
	Message string `json:"message"`
}

func (g *CustomProcessor) Init(raw []byte) error {
	// Here you can initialize your processor with the provided configuration
	// For example, you might want to set up connections, load resources, etc.
	var config Config
	if err := json.Unmarshal(raw, &config); err != nil {
		g.logger.WithError(err).Error("Failed to unmarshal configuration")
		return err
	}

	g.logger.WithField("message", config.Message).Info("CustomProcessor initialized with config")
	return nil
}

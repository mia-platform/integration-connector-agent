// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package rpcprocessor

import (
	"github.com/mia-platform/integration-connector-agent/entities"
	"github.com/mia-platform/integration-connector-agent/internal/processors/hcgp"

	"github.com/hashicorp/go-plugin"
)

type Config struct {
	Processor entities.InitializableProcessor
	Logger    Logger
}

func Serve(config *Config) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: hcgp.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			hcgp.PluginProcessorKey: &hcgp.PluginAdapter{Impl: config.Processor},
		},
		Logger: hcgp.NewLogAdapter(config.Logger),
	})
}

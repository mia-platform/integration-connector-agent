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

package hcgp

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	hashicorpplugin "github.com/mia-platform/integration-connector-agent/adapters/hashicorp-plugin"
	"github.com/mia-platform/integration-connector-agent/entities"
)

var (
	ErrPluginDispense       = fmt.Errorf("plugin dispense error")
	ErrPluginInitialization = fmt.Errorf("plugin initialization error")
)

type Plugin struct {
	client    *plugin.Client
	rpcClient plugin.ClientProtocol
}

// handshakeConfigs are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion: 1,
	MagicCookieKey:  "integration-connector-agent-plugin",
	// TODO: make this configurable
	MagicCookieValue: "go-plugin",
}

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"processor": &hashicorpplugin.ProcessorAdapter{},
}

func New(cfg Config) (entities.Processor, error) {
	// TODO: use standard JSON Logger format and set the right log level!
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "hc-go-plugin",
		Output:     os.Stdout,
		JSONFormat: true,
		Level:      hclog.Debug,
	})

	client := plugin.NewClient(&plugin.ClientConfig{
		Logger:          logger,
		Plugins:         pluginMap,
		HandshakeConfig: handshakeConfig,

		// #nosec:G204: this is path is configuration based and only used at service bootstrap
		Cmd: exec.Command(cfg.ModulePath),
		// UnixSocketConfig: &plugin.UnixSocketConfig{
		// 	TempDir: cfg.TempDir,
		// },
	})

	// TODO: We don't have a "stop" processor interface
	// defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPluginInitialization, err)
	}
	// defer rpcClient.Close()
	return &Plugin{
		client:    client,
		rpcClient: rpcClient,
	}, nil
}

func (p *Plugin) Process(event entities.PipelineEvent) (entities.PipelineEvent, error) {
	raw, err := p.rpcClient.Dispense("processor")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPluginDispense, err)
	}

	processorAdapter, ok := raw.(entities.Processor)
	if !ok {
		return nil, fmt.Errorf("%w: invalid interface type", ErrPluginDispense)
	}
	fmt.Printf("PROCESSOR WRAPPER STARTING PROCESS\n")
	return processorAdapter.Process(event)
}

func (p *Plugin) Close() error {
	if p.rpcClient != nil {
		if err := p.rpcClient.Close(); err != nil {
			return err
		}
	}

	if p.client != nil {
		p.client.Kill()
	}
	return nil
}

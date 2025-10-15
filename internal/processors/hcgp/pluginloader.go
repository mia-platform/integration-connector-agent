// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package hcgp

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"
)

const PluginProcessorKey = "processor"

var (
	ErrPluginDispense       = errors.New("plugin dispense error")
	ErrPluginInitialization = errors.New("plugin initialization error")
)

type Plugin struct {
	client    *plugin.Client
	rpcClient plugin.ClientProtocol
}

// HandshakeConfig are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "integration-connector-agent-plugin",
	MagicCookieValue: "rpc-plugin",
}

func New(log *logrus.Logger, cfg Config) (entities.Processor, error) {
	var pluginMap = map[string]plugin.Plugin{
		PluginProcessorKey: &PluginAdapter{},
	}

	log.WithFields(logrus.Fields{
		"modulePath": cfg.ModulePath,
	}).Trace("initializing plugin")

	client := plugin.NewClient(&plugin.ClientConfig{
		// #nosec:G204: this path is configuration based and only used at service bootstrap
		Cmd:             exec.Command(cfg.ModulePath),
		Logger:          NewLogAdapter(log),
		Plugins:         pluginMap,
		HandshakeConfig: HandshakeConfig,
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPluginInitialization, err)
	}

	p := &Plugin{
		client:    client,
		rpcClient: rpcClient,
	}

	return p.Init(cfg.InitOptions)
}

func (p *Plugin) Init(initOptions []byte) (entities.Processor, error) {
	if len(initOptions) == 0 {
		return p, nil
	}

	raw, err := p.rpcClient.Dispense(PluginProcessorKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPluginDispense, err)
	}
	processorAdapter, ok := raw.(entities.Initializable)
	if !ok {
		return nil, errors.New("invalid interface type, expected entities.Initializable")
	}

	if err := processorAdapter.Init(initOptions); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPluginInitialization, err)
	}
	return p, nil
}

func (p *Plugin) Process(event entities.PipelineEvent) (entities.PipelineEvent, error) {
	raw, err := p.rpcClient.Dispense(PluginProcessorKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPluginDispense, err)
	}

	processorAdapter, ok := raw.(entities.Processor)
	if !ok {
		return nil, errors.New("invalid interface type, expected entities.Processor")
	}

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

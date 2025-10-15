// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package hcgp

import (
	"fmt"
	"net/rpc"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/hashicorp/go-plugin"
)

const (
	ProcessRPCMethodName = "Plugin.Process"
	InitRPCMethodName    = "Plugin.Init"
)

type PluginAdapter struct {
	Impl entities.InitializableProcessor
}

func (p *PluginAdapter) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

func (PluginAdapter) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}

// Client used by the integration connecter agent to call the plugin.
type RPCClient struct{ client *rpc.Client }

func (g *RPCClient) Process(event entities.PipelineEvent) (entities.PipelineEvent, error) {
	var resp entities.Event

	if err := g.client.Call(ProcessRPCMethodName, event, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (g *RPCClient) Init(options []byte) error {
	resp := struct{}{}
	if err := g.client.Call(InitRPCMethodName, options, &resp); err != nil {
		return err
	}
	return nil
}

// Server used by the plugin to invoke the plugin processor when integration connector agent calls the plugin.
type RPCServer struct {
	Impl entities.InitializableProcessor
}

func (g *RPCServer) Process(input entities.Event, output *entities.Event) error {
	result, err := g.Impl.Process(&input)
	if err != nil {
		return err
	}
	concreteResult, ok := result.(*entities.Event)
	if !ok {
		return fmt.Errorf("expected *entities.Event, got %T", result)
	}
	*output = *concreteResult
	return nil
}

func (g *RPCServer) Init(options []byte, _ *struct{}) error {
	if err := g.Impl.Init(options); err != nil {
		return fmt.Errorf("plugin initialization error: %w", err)
	}
	return nil
}

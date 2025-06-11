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

package hashicorpplugin

import (
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/mia-platform/integration-connector-agent/entities"
)

type ProcessorAdapter struct {
	Impl entities.Processor
}

func (p *ProcessorAdapter) Server(*plugin.MuxBroker) (interface{}, error) {
	fmt.Printf("[[[[ADAPTER]]]] ProcessorAdapter.Server\n")
	return &RPCServerProcerssor{Impl: p.Impl}, nil
}

func (ProcessorAdapter) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	fmt.Printf("[[[[ADAPTER]]]] ProcessorAdapter.Client\n")
	return &RPCProcessorClient{client: c}, nil
}

// Client used by the integration connecter agent to call the plugin.
type RPCProcessorClient struct{ client *rpc.Client }

func (g *RPCProcessorClient) Process(event entities.PipelineEvent) (entities.PipelineEvent, error) {
	fmt.Printf("[[[[ADAPTER]]]] starting processor in RPCProcessorClient.Process %+v %p\n", event, g.client)

	var resp entities.Event
	err := g.client.Call("Plugin.Process", event, &resp)
	if err != nil {
		return nil, err
	}
	fmt.Printf("[[[[ADAPTER]]]] call done RPCProcessorClient.Process %+v %v\n", event, err)

	return &resp, err
}

// Server used by the plugin to invoke the plugin processor when integration connector agent calls the plugin.
type RPCServerProcerssor struct {
	Impl entities.Processor
}

func (g *RPCServerProcerssor) Process(input entities.Event, output *entities.Event) error {
	fmt.Printf("[[[[ADAPTER]]]] starting processor in RPCServerProcerssor.Process %+v %p\n", input, g.Impl)
	// *output = input

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

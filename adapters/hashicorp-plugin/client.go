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
	"net/rpc"

	"github.com/mia-platform/integration-connector-agent/entities"
)

const (
	ProcessRPCMethodName = "Plugin.Process"
	InitRPCMethodName    = "Plugin.Init"
)

// Client used by the integration connecter agent to call the plugin.
type RPCClient struct{ client *rpc.Client }

func (g *RPCClient) Process(event entities.PipelineEvent) (entities.PipelineEvent, error) {
	var resp entities.Event

	if err := g.client.Call(ProcessRPCMethodName, event, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (g *RPCClient) Init(options map[string]interface{}) error {
	resp := struct{}{}
	if err := g.client.Call(InitRPCMethodName, options, &resp); err != nil {
		return err
	}
	return nil
}

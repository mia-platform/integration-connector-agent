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

	"github.com/mia-platform/integration-connector-agent/entities"
)

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

func (g *RPCServer) Init(options map[string]interface{}, _ *struct{}) error {
	fmt.Printf("PLUGIN INIT!!!!!!")
	if err := g.Impl.Init(options); err != nil {
		return fmt.Errorf("plugin initialization error: %w", err)
	}
	return nil
}

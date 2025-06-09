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

package customprocessor

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"plugin"

	"github.com/mia-platform/integration-connector-agent/entities"
)

var (
	ErrPluginLoadFailed      = errors.New("can't load plugin")
	ErrInvalidPluginSignture = errors.New("plugin function has wrong signature")
)

type Plugin struct {
	module *plugin.Plugin
	proc   entities.Processor
}

func New(cfg Config) (*Plugin, error) {
	fmt.Printf("Loading plugin from %s with options: %v\n", cfg.ModulePath, cfg.InitOptions)
	fileexists, err := os.Stat(cfg.ModulePath)
	fmt.Printf("File exists: %v, error: %v\n", fileexists != nil, err)
	fmt.Printf("FILE INfo %+v\n", fileexists)

	module, err := plugin.Open(cfg.ModulePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPluginLoadFailed, err)
	}

	symbol, err := module.Lookup("Initialize")
	if err != nil {
		return nil, err
	}
	initFunc, ok := symbol.(func([]byte) (entities.Processor, error))
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPluginSignture, "Initialize function has wrong signature")
	}

	options, err := json.Marshal(cfg.InitOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plugin options: %w", err)
	}

	proc, err := initFunc(options)
	if err != nil {
		return nil, err
	}

	return &Plugin{module, proc}, nil
}

func (p *Plugin) Process(event entities.PipelineEvent) (entities.PipelineEvent, error) {
	return p.proc.Process(event)
}

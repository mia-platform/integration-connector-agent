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

package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	gp "plugin"

	"github.com/mia-platform/integration-connector-agent/entities"
)

var (
	ErrPluginLoadFailed      = errors.New("can't load plugin")
	ErrInvalidPluginSignture = errors.New("plugin function has wrong signature")
)

type Plugin struct {
	module *gp.Plugin
}

func New(cfg Config) (*Plugin, error) {
	module, err := gp.Open(cfg.ModulePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrPluginLoadFailed, err)
	}

	symbol, err := module.Lookup("Initialize")
	if err != nil {
		return nil, err
	}
	initFunc, ok := symbol.(func([]byte) error)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPluginSignture, "Initialize function has wrong signature")
	}

	options, err := json.Marshal(cfg.InitOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plugin options: %w", err)
	}
	if err := initFunc(options); err != nil {
		return nil, err
	}

	return &Plugin{module}, nil
}

func (p *Plugin) Process(event entities.PipelineEvent) (entities.PipelineEvent, error) {
	symbol, err := p.module.Lookup("Process")
	if err != nil {
		return nil, fmt.Errorf("failed to lookup Process function: %w", err)
	}

	processFunc, ok := symbol.(func(entities.PipelineEvent) (entities.PipelineEvent, error))
	if !ok {
		return nil, fmt.Errorf("Process function has wrong signature")
	}

	return processFunc(event)
}

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

package mongo

import (
	"fmt"

	"github.com/mia-platform/data-connector-agent/internal/config"
)

// Config contains the configuration needed to connect to a remote MongoDB instance
type Config struct {
	URI         config.SecretSource
	Database    string
	Collection  string
	OutputEvent map[string]any
	IDField     string
}

func (c *Config) Validate() error {
	if c.URI == "" {
		return fmt.Errorf("%w: URI is empty", config.ErrConfigNotValid)
	}
	if c.Collection == "" {
		return fmt.Errorf("%w: collection is empty", config.ErrConfigNotValid)
	}
	if c.OutputEvent == nil {
		return fmt.Errorf("%w: output event not set", config.ErrConfigNotValid)
	}
	if c.IDField == "" {
		c.IDField = "_id"
	}
	if _, ok := c.OutputEvent[c.IDField]; !ok {
		return fmt.Errorf("%w: ID field \"%s\" not found in output event", config.ErrConfigNotValid, c.IDField)
	}

	return nil
}

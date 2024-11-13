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

	"github.com/mia-platform/integration-connector-agent/internal/config"
)

// Config contains the configuration needed to connect to a remote MongoDB instance
type Config struct {
	URL         config.SecretSource `json:"url"`
	Database    string              `json:"-"`
	Collection  string              `json:"collection"`
	OutputEvent map[string]any      `json:"outputEvent"`
	IDField     string              `json:"idField"`
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("url is required")
	}
	if c.Collection == "" {
		return fmt.Errorf("collection is required")
	}
	if c.OutputEvent == nil {
		return fmt.Errorf("outputEvent is required")
	}
	if c.IDField == "" {
		return fmt.Errorf("idField is required")
	}
	if c.IDField == "_id" {
		return fmt.Errorf("idField cannot be \"_id\"")
	}
	if _, ok := c.OutputEvent[c.IDField]; !ok {
		return fmt.Errorf("idField \"%s\" not found in outputEvent", c.IDField)
	}

	return nil
}

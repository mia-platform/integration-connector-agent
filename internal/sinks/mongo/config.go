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
	"errors"

	"github.com/mia-platform/integration-connector-agent/internal/config"
)

// Config contains the configuration needed to connect to a remote MongoDB instance
type Config struct {
	URL        config.SecretSource `json:"url"`
	Collection string              `json:"collection"`
	InsertOnly bool                `json:"insertOnly"`

	Database string `json:"-"`
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return errors.New("url is required")
	}
	if c.Collection == "" {
		return errors.New("collection is required")
	}

	return nil
}

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

package jira

import (
	"github.com/mia-platform/data-connector-agent/internal/utils"
)

// Configuration is the rappresentation of the configuration for a Jira Cloud webhook
type Configuration struct {
	// Secret the webhook secret configuration for validating the data received
	Secret utils.SecretSource `json:"secret"`
}

// ReadConfiguration return the configuration data contained in the file at path or an error if it cannot be read or
// parsed correctly
func ReadConfiguration(path string) (*Configuration, error) {
	config := new(Configuration)
	if err := utils.ReadJSONFile(path, config); err != nil {
		return nil, err
	}

	return config, nil
}

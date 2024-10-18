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

type Configuration struct {
	Secret WebhookSecret `json:"secret"`
}

type WebhookSecret struct {
	FromEnv  string `json:"fromEnv"`
	FromFile string `json:"fromFile"`
}

func ReadConfiguration(path string) (*Configuration, error) {
	config := new(Configuration)
	if err := utils.ReadJSONFile(path, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Configuration) WebhookSecret() string {
	switch {
	case c.Secret.FromEnv != "":
	case c.Secret.FromFile != "":
	}

	return ""
}

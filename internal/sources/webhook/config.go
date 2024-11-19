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

package webhook

import (
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/config"
)

type Authentication struct {
	Secret     config.SecretSource `json:"secret"`
	HeaderName string              `json:"headerName"`
}

// Configuration is the representation of the configuration for a Jira Cloud webhook
type Configuration struct {
	// Secret the webhook secret configuration for validating the data received
	Authentication Authentication `json:"authentication"`
	WebhookPath    string         `json:"webhookPath"`

	Events *Events `json:"events,omitempty"`
}

func (c *Configuration) Validate() error {
	if c.WebhookPath == "" {
		return fmt.Errorf("webhook path is empty")
	}

	if c.Events == nil {
		return fmt.Errorf("events are empty")
	}

	return nil
}

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
	"errors"
	"fmt"
)

var (
	ErrWebhookPathRequired = errors.New("webhook path is required")
)

type ValidatingRequest interface {
	GetReqHeaders() map[string][]string
	Body() []byte
}

type Authentication interface {
	CheckSignature(req ValidatingRequest) error
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
		return ErrWebhookPathRequired
	}

	if c.Events == nil {
		return fmt.Errorf("events are empty")
	}

	return nil
}

func (c *Configuration) CheckSignature(req ValidatingRequest) error {
	if c == nil || c.Authentication == nil {
		return nil
	}
	return c.Authentication.CheckSignature(req)
}

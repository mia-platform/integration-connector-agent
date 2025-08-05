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

	"github.com/mia-platform/integration-connector-agent/internal/utils"
)

var (
	ErrWebhookPathRequired                = errors.New("webhook path is required")
	ErrSupportedEventsRequired            = errors.New("supported events are required")
	ErrInvalidWebhookAuthenticationConfig = errors.New("invalid webhook authentication configuration")
	ErrMissingRequest                     = errors.New("missing request for webhook authentication check")
)

type ValidatingRequest interface {
	GetReqHeaders() map[string][]string
	Body() []byte
}

type Authentication interface {
	CheckSignature(req ValidatingRequest) error
	Validate() error
}

// ContentTypeConfig allows configuring how to extract the payload field for a given content-type
type ContentTypeConfig map[string]string

type Configuration[T Authentication] struct {
	// Secret the webhook secret configuration for validating the data received
	Authentication T      `json:"authentication"`
	WebhookPath    string `json:"webhookPath"`

	Events *Events `json:"-"`
}

func (c *Configuration[T]) Validate() error {
	if c.WebhookPath == "" {
		return ErrWebhookPathRequired
	}

	if c.Events == nil {
		return ErrSupportedEventsRequired
	}

	if !utils.IsNil(c.Authentication) {
		return c.Authentication.Validate()
	}

	return nil
}

func (c *Configuration[T]) CheckSignature(req ValidatingRequest) error {
	if c == nil || utils.IsNil(c.Authentication) {
		return nil
	}

	if req == nil {
		return ErrMissingRequest
	}

	return c.Authentication.CheckSignature(req)
}

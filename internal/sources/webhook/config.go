// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

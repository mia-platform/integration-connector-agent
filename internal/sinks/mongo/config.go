// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

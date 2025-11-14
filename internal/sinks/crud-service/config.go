// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package crudservice

import (
	"errors"
	"fmt"
	"net/url"
)

var (
	ErrURLNotSet  = errors.New("URL not set in CRUD service sink configuration")
	ErrInvalidURL = errors.New("invalid URL in CRUD service sink configuration")
)

const DefaultPrimaryKey = "_eventId"

type Config struct {
	URL        string `json:"url"`
	InsertOnly bool   `json:"insertOnly,omitempty"`
	PrimaryKey string `json:"primaryKeyFieldName,omitempty"` //nolint: tagliatelle
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return ErrURLNotSet
	}

	if _, err := url.Parse(c.URL); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	if c.PrimaryKey == "" {
		c.PrimaryKey = DefaultPrimaryKey
	}

	return nil
}

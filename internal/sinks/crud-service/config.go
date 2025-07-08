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
	PrimaryKey string `json:"primaryKey,omitempty"`
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return ErrURLNotSet
	}

	if _, err := url.Parse(c.URL); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidURL, err)
	}

	if c.PrimaryKey == "" {
		c.PrimaryKey = DefaultPrimaryKey
	}

	return nil
}

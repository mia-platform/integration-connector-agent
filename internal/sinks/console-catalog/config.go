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

package consolecatalog

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/mia-platform/integration-connector-agent/config"
	"github.com/mia-platform/integration-connector-agent/internal/sinks/console-catalog/consoleclient"
)

var (
	ErrURLNotSet              = errors.New("URL not set in Console Catalog sink configuration")
	ErrInvalidURL             = errors.New("invalid URL in Console Catalog sink configuration")
	ErrInvalidLifecycleStatus = errors.New("invalid itemLifecycleStatus in Console Catalog sink configuration")
	ErrMissingField           = errors.New("missing required field in Console Catalog sink configuration")
)

type Config struct {
	URL                 string                        `json:"url"`
	TenantID            string                        `json:"tenantId"`
	ClientID            string                        `json:"clientId"`
	ClientSecret        config.SecretSource           `json:"clientSecret"`
	ItemType            string                        `json:"itemType"`
	ItemLifecycleStatus consoleclient.LifecycleStatus `json:"itemLifecycleStatus"`
	ItemIDTemplate      string                        `json:"itemIdTemplate"`
	ItemNameTemplate    string                        `json:"itemNameTemplate"`
}

func (c *Config) Validate() error {
	if c.URL == "" {
		return ErrURLNotSet
	}

	if _, err := url.Parse(c.URL); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidURL, err)
	}

	if c.TenantID == "" {
		return fmt.Errorf("%w: tenantId", ErrMissingField)
	}

	if c.ItemType == "" {
		return fmt.Errorf("%w: itemType", ErrMissingField)
	}

	if c.ClientID == "" {
		return fmt.Errorf("%w: clientId", ErrMissingField)
	}

	if c.ClientSecret.String() == "" {
		return fmt.Errorf("%w: clientSecret", ErrMissingField)
	}

	if c.ItemIDTemplate == "" {
		return fmt.Errorf("%w: itemIdTemplate", ErrMissingField)
	}

	if c.ItemNameTemplate == "" {
		return fmt.Errorf("%w: itemNameTemplate", ErrMissingField)
	}

	if c.ItemLifecycleStatus == "" {
		c.ItemLifecycleStatus = consoleclient.Published
	}

	if !consoleclient.IsValidLifecycleStatus(string(c.ItemLifecycleStatus)) {
		return fmt.Errorf("%w: %s", ErrInvalidLifecycleStatus, c.ItemLifecycleStatus)
	}

	return nil
}

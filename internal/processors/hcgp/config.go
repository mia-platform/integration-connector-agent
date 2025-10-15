// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package hcgp

import "encoding/json"

type Config struct {
	ModulePath  string          `json:"modulePath"`
	InitOptions json.RawMessage `json:"initOptions,omitempty"`
}

func (c Config) Validate() error {
	return nil
}

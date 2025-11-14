// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package hcgp

import "encoding/json"

type Config struct {
	ModulePath  string          `json:"modulePath"`
	InitOptions json.RawMessage `json:"initOptions,omitempty"`
}

func (c Config) Validate() error {
	return nil
}

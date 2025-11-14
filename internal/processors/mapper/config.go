// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package mapper

import "encoding/json"

type Config struct {
	OutputEvent json.RawMessage `json:"outputEvent"`
}

func (c Config) Validate() error {
	return nil
}

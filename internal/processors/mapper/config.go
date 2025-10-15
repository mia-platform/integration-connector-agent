// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package mapper

import "encoding/json"

type Config struct {
	OutputEvent json.RawMessage `json:"outputEvent"`
}

func (c Config) Validate() error {
	return nil
}

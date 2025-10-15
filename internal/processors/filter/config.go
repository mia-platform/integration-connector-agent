// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package filter

type Config struct {
	CELExpression string `json:"celExpression"`
}

func (c Config) Validate() error {
	return nil
}

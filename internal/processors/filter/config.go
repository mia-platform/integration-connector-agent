// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package filter

type Config struct {
	CELExpression string `json:"celExpression"`
}

func (c Config) Validate() error {
	return nil
}

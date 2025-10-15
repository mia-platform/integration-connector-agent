// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package config

import (
	"encoding/json"
	"fmt"
)

type GenericConfig struct {
	Type string `json:"type"`

	Raw []byte `json:"-"`
}

func (w *GenericConfig) UnmarshalJSON(data []byte) error {
	var writerConfig struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(data, &writerConfig); err != nil {
		return err
	}

	w.Type = writerConfig.Type
	w.Raw = data

	return nil
}

type Validator interface {
	Validate() error
}

func GetConfig[T Validator](config GenericConfig) (T, error) {
	var cfg T
	if err := json.Unmarshal(config.Raw, &cfg); err != nil {
		return cfg, err
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("%w: %w", ErrConfigNotValid, err)
	}

	return cfg, nil
}

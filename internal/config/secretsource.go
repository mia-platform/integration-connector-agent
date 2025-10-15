// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package config

import (
	"encoding/json"
	"os"
	"strings"
)

type SecretSource string

func (s SecretSource) String() string {
	return string(s)
}

type secretConfig struct {
	FromEnv  string `json:"fromEnv"`
	FromFile string `json:"fromFile"`
}

func readSecret(s *secretConfig) string {
	secret := ""
	switch {
	case s.FromEnv != "":
		secret = secretFromEnv(s.FromEnv)
	case s.FromFile != "":
		secret = secretFromFile(s.FromFile)
	}

	return strings.TrimSpace(secret)
}

func (s *SecretSource) UnmarshalJSON(data []byte) error {
	aux := &secretConfig{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	secret := readSecret(aux)
	*s = SecretSource(secret)

	return nil
}

// secretFromEnv return the value contained in envName environment variable or the empty string if is not found
func secretFromEnv(envName string) string {
	secret, _ := os.LookupEnv(envName)
	return secret
}

// secretFromFile return the value contained in the file at filePath or the empty string if an error is encountered
// during the read operation
func secretFromFile(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	return string(data)
}

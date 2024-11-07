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

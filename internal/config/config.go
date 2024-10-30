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
	"path"
	"strings"

	"github.com/mia-platform/configlib"
)

type Writer struct {
	Type        string         `json:"type"`
	URL         SecretSource   `json:"url"`
	OutputEvent map[string]any `json:"outputEvent"`
}

type Authentication struct {
	Secret SecretSource `json:"secret"`
}

type Integrations struct {
	Type           string         `json:"type"`
	Authentication Authentication `json:"authentication"`
	Writers        []Writer       `json:"writers"`
}

type Configuration struct {
	Integrations []Integrations `json:"integrations"`
}

func LoadServiceConfiguration(filePath string) (*Configuration, error) {
	jsonSchema, err := configlib.ReadFile("./config.schema.json")
	if err != nil {
		return nil, err
	}

	fileName := path.Base(filePath)
	fileNameWithoutExt := strings.TrimSuffix(fileName, path.Ext(fileName))
	dir := path.Dir(filePath)

	var config *Configuration
	if err := configlib.GetConfigFromFile(fileNameWithoutExt, dir, jsonSchema, &config); err != nil {
		return nil, err
	}

	return config, nil
}

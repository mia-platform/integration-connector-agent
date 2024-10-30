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
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

type Writer struct {
	Type        string         `json:"type"`
	URL         SecretSource   `json:"url"`
	OutputEvent map[string]any `json:"outputEvent"`
	Collection  string         `json:"collection"`
}

type Authentication struct {
	Secret SecretSource `json:"secret"`
}

type Integration struct {
	Type           string         `json:"type"`
	Authentication Authentication `json:"authentication"`
	Writers        []Writer       `json:"writers"`
}

type Configuration struct {
	Integrations []Integration `json:"integrations"`
}

var (
	ErrConfigNotValid = fmt.Errorf("configuration not valid")
)

func LoadServiceConfiguration(filePath string) (*Configuration, error) {
	jsonSchema, err := readFile("./config.schema.json")
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConfigNotValid, err)
	}

	jsonConfig, err := readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConfigNotValid, err)
	}

	if err = validateJSONConfig(jsonSchema, jsonConfig); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConfigNotValid, err)
	}

	var config *Configuration
	if err := json.Unmarshal(jsonConfig, &config); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConfigNotValid, err)
	}

	return config, nil
}

func validateJSONConfig(schema, jsonConfig []byte) error {
	schemaLoader := gojsonschema.NewBytesLoader(schema)
	documentLoader := gojsonschema.NewBytesLoader(jsonConfig)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("error validating: %s", err.Error())
	}
	if !result.Valid() {
		return fmt.Errorf("json schema validation errors: %s", result.Errors())
	}
	return nil
}

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
	"embed"
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

var (
	ErrConfigNotValid = fmt.Errorf("configuration not valid")
)

//go:embed config.schema.json
var jsonSchemaFile embed.FS

type Writer struct {
	Type string `json:"type"`

	Raw []byte `json:"-"`
}

func (w *Writer) UnmarshalJSON(data []byte) error {
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

type WriterConfigValidator interface {
	Validate() error
}

func WriterConfig[T WriterConfigValidator](config Writer) (T, error) {
	var cfg T
	if err := json.Unmarshal(config.Raw, &cfg); err != nil {
		return cfg, err
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("%w: %s", ErrConfigNotValid, err)
	}

	return cfg, nil
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

func LoadServiceConfiguration(filePath string) (*Configuration, error) {
	jsonSchema, err := jsonSchemaFile.ReadFile("config.schema.json")
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

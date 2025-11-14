// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package config

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

var (
	ErrConfigNotValid = errors.New("configuration not valid")
)

//go:embed config.schema.json
var jsonSchemaFile embed.FS

type Authentication struct {
	Secret SecretSource `json:"secret"`
}

type Processors []GenericConfig
type Sinks []GenericConfig
type Source GenericConfig

type Pipeline struct {
	Processors Processors `json:"processors"`
	Sinks      Sinks      `json:"sinks"`
}

type Integration struct {
	Source    GenericConfig `json:"source"`
	Pipelines []Pipeline    `json:"pipelines"`
}

type Configuration struct {
	Integrations []Integration `json:"integrations"`
}

func LoadServiceConfiguration(filePath string) (*Configuration, error) {
	jsonSchema, err := jsonSchemaFile.ReadFile("config.schema.json")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigNotValid, err)
	}

	jsonConfig, err := readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigNotValid, err)
	}

	if err = validateJSONConfig(jsonSchema, jsonConfig); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigNotValid, err)
	}

	var config *Configuration
	if err := json.Unmarshal(jsonConfig, &config); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConfigNotValid, err)
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

// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	errTransform = errors.New("error transforming data")
	errOperation = errors.New("error creating operation")
	errCast      = errors.New("error casting value")
)

var dynamicOperatorRegexp = regexp.MustCompile(`{{\s*(.*?)\s*}}`)

type operationType int

const (
	set operationType = iota
)

type castType string

const (
	castToString castType = "string"
	castToNumber castType = "number"
	castToNone   castType = ""
)

type operation struct {
	keyToUpdate   string
	valueKeys     []string
	operationType operationType
	templateValue any
	castTo        castType
}

func (o operation) getValueToSet(input []byte) any {
	if len(o.valueKeys) == 0 {
		return o.templateValue
	}
	key := o.valueKeys[0]
	value := gjson.GetBytes(input, key)
	return value.Value()
}

func (o operation) castValue(value any) (any, error) {
	if o.castTo == castToNone {
		return value, nil
	}

	switch o.castTo {
	case castToString:
		return fmt.Sprintf("%v", value), nil
	case castToNumber:
		switch v := value.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string:
			num, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("%w: cannot cast string '%s' to number: %w", errCast, v, err)
			}
			return num, nil
		case bool:
			if v {
				return float64(1), nil
			}
			return float64(0), nil
		default:
			return nil, fmt.Errorf("%w: cannot cast type %T to number", errCast, value)
		}
	default:
		return nil, fmt.Errorf("%w: unsupported cast type: %s", errCast, o.castTo)
	}
}

func (o operation) apply(input, output []byte) ([]byte, error) {
	switch o.operationType {
	case set:
		output, err := o.setData(input, output)
		if err != nil {
			return nil, err
		}
		return output, nil
	default:
		return nil, fmt.Errorf("operation not supported: %v", o.operationType)
	}
}

func (o operation) setData(input, output []byte) ([]byte, error) {
	if !gjson.ValidBytes(input) {
		err := json.Unmarshal(input, &map[string]any{})
		return nil, fmt.Errorf("%w: %w", errTransform, err)
	}
	if o.keyToUpdate == "" {
		return nil, fmt.Errorf("%w: output key is empty", errTransform)
	}

	valueToSet := o.getValueToSet(input)
	
	// Apply casting if configured
	if o.castTo != castToNone {
		castedValue, err := o.castValue(valueToSet)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to cast value for key '%s': %w", errTransform, o.keyToUpdate, err)
		}
		valueToSet = castedValue
	}

	out, err := sjson.SetBytes(output, o.keyToUpdate, valueToSet)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errTransform, err)
	}

	return out, nil
}

// newOperation creates a new operation from the given input key and output value.
// keyToUpdate are the keys where to retrieve the value from the input JSON.
// outputValue is the key where to set the value in the output JSON.
func newOperation(keyToUpdate string, template gjson.Result) (operation, error) {
	var templateValue gjson.Result
	var castTo castType

	// Check if template is an object with "value" and "castTo" fields
	if template.IsObject() {
		valueField := template.Get("value")
		castToField := template.Get("castTo")
		
		if valueField.Exists() && castToField.Exists() {
			// Extract the castTo value
			castToStr := castToField.String()
			if castToStr != "" {
				// Validate castTo value
				if castToStr != string(castToString) && castToStr != string(castToNumber) {
					return operation{}, fmt.Errorf("%w: invalid castTo value '%s', must be 'string' or 'number'", errOperation, castToStr)
				}
				castTo = castType(castToStr)
			}
			templateValue = valueField
		} else {
			// If it's an object but doesn't have the value/castTo structure, treat it as a regular template
			templateValue = template
		}
	} else {
		templateValue = template
	}

	matches := dynamicOperatorRegexp.FindAllStringSubmatch(templateValue.String(), -1)
	if len(matches) > 1 {
		return operation{}, fmt.Errorf("%w: unsupported combine template: %s", errOperation, templateValue)
	}

	var valueKeys []string
	for _, match := range matches {
		valueKeys = append(valueKeys, match[1:]...)
	}

	return operation{
		valueKeys:     valueKeys,
		keyToUpdate:   keyToUpdate,
		operationType: set,
		templateValue: templateValue.Value(),
		castTo:        castTo,
	}, nil
}

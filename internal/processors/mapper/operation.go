// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package mapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	errTransform = errors.New("error transforming data")
	errOperation = errors.New("error creating operation")
)

var dynamicOperatorRegexp = regexp.MustCompile(`{{\s*(.*?)\s*}}`)

type operationType int

const (
	set operationType = iota
)

type operation struct {
	keyToUpdate   string
	valueKeys     []string
	operationType operationType
	templateValue any
}

func (o operation) getValueToSet(input []byte) any {
	if len(o.valueKeys) == 0 {
		return o.templateValue
	}
	key := o.valueKeys[0]
	value := gjson.GetBytes(input, key)
	return value.Value()
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

	out, err := sjson.SetBytes(output, o.keyToUpdate, o.getValueToSet(input))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errTransform, err)
	}

	return out, nil
}

// newOperation creates a new operation from the given input key and output value.
// keyToUpdate are the keys where to retrieve the value from the input JSON.
// outputValue is the key where to set the value in the output JSON.
func newOperation(keyToUpdate string, template gjson.Result) (operation, error) {
	matches := dynamicOperatorRegexp.FindAllStringSubmatch(template.String(), -1)
	if len(matches) > 1 {
		return operation{}, fmt.Errorf("%w: unsupported combine template: %s", errOperation, template)
	}

	var valueKeys []string
	for _, match := range matches {
		valueKeys = append(valueKeys, match[1:]...)
	}

	return operation{
		valueKeys:     valueKeys,
		keyToUpdate:   keyToUpdate,
		operationType: set,
		templateValue: template.Value(),
	}, nil
}

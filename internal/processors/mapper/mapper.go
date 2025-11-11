// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package mapper

import (
	"encoding/json"
	"strings"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Mapper struct {
	operations []operation
}

func (m Mapper) Process(event entities.PipelineEvent) (entities.PipelineEvent, error) {
	logrus.WithFields(logrus.Fields{
		"eventType":      event.GetType(),
		"primaryKeys":    event.GetPrimaryKeys().Map(),
		"operationCount": len(m.operations),
	}).Debug("starting mapper processing")

	output := []byte("{}")
	var err error
	for i, operation := range m.operations {
		logrus.WithFields(logrus.Fields{
			"eventType":      event.GetType(),
			"primaryKeys":    event.GetPrimaryKeys().Map(),
			"operationIndex": i,
		}).Trace("applying mapper operation")

		output, err = operation.apply(event.Data(), output)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"eventType":      event.GetType(),
				"primaryKeys":    event.GetPrimaryKeys().Map(),
				"operationIndex": i,
			}).WithError(err).Error("mapper operation failed")
			return nil, err
		}
	}
	event.WithData(output)

	logrus.WithFields(logrus.Fields{
		"eventType":   event.GetType(),
		"primaryKeys": event.GetPrimaryKeys().Map(),
		"outputSize":  len(output),
	}).Debug("mapper processing completed successfully")

	return event, nil
}

func New(cfg Config) (*Mapper, error) {
	model, err := json.Marshal(cfg.OutputEvent)
	if err != nil {
		return nil, err
	}

	ops, err := generateOperations(gjson.ParseBytes(model))
	if err != nil {
		return nil, err
	}

	return &Mapper{
		operations: ops,
	}, nil
}

func generateOperations(jsonData gjson.Result) ([]operation, error) {
	result := []operation{}
	var resError error

	var walk func(data gjson.Result, keyPrefix string)
	walk = func(data gjson.Result, keyPrefix string) {
		data.ForEach(func(key, value gjson.Result) bool {
			keyToUpdate := key.String()
			if keyPrefix != "" {
				keyToUpdate = strings.Join([]string{keyPrefix, key.String()}, ".")
			}

			// Check if this is a casting configuration object
			isCastConfig := value.IsObject() && value.Get("value").Exists() && value.Get("castTo").Exists()

			if !isCastConfig && (value.IsObject() || value.IsArray()) {
				walk(value, keyToUpdate)
				return true
			}

			if key.Exists() && key.String() != "" {
				operation, err := newOperation(keyToUpdate, value)
				if err != nil {
					resError = err
					return false
				}
				result = append(result, operation)
			}

			return true
		})
	}

	walk(jsonData, "")

	return result, resError
}

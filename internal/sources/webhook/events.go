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

package webhook

import (
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/sirupsen/logrus"

	"github.com/tidwall/gjson"
)

var (
	ErrMissingFieldID = fmt.Errorf("missing id field in event")
)

type Events struct {
	Supported          map[string]Event
	EventTypeFieldPath string
}

type Event struct {
	Operation  entities.Operation
	GetFieldID func(parsedData gjson.Result) entities.PkFields
}

func GetPrimaryKeyByPath(path string) func(parsedData gjson.Result) entities.PkFields {
	return func(parsedData gjson.Result) entities.PkFields {
		value := parsedData.Get(path).String()
		if value == "" {
			return nil
		}

		return entities.PkFields{{Key: path, Value: value}}
	}
}

func (e *Events) getPipelineEvent(logger *logrus.Entry, rawData []byte) (entities.PipelineEvent, error) {
	parsed := gjson.ParseBytes(rawData)
	webhookEvent := parsed.Get(e.EventTypeFieldPath).String()

	event, ok := e.Supported[webhookEvent]
	if !ok {
		logger.WithFields(logrus.Fields{
			"webhookEvent": webhookEvent,
			"event":        string(rawData),
		}).Trace("unsupported webhook event")
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedWebhookEvent, webhookEvent)
	}

	if event.GetFieldID == nil {
		logger.WithFields(logrus.Fields{
			"webhookEvent": webhookEvent,
			"event":        string(rawData),
		}).Trace("missing GetFieldID function")
		return nil, fmt.Errorf("%w: %s missing GetFieldID function", ErrUnsupportedWebhookEvent, webhookEvent)
	}

	pk := event.GetFieldID(parsed)
	if pk.IsEmpty() {
		logger.WithFields(logrus.Fields{
			"webhookEvent": webhookEvent,
			"event":        string(rawData),
		}).Trace("unsupported webhook event")
		return nil, fmt.Errorf("%w: %s", ErrMissingFieldID, webhookEvent)
	}

	return &entities.Event{
		PrimaryKeys:   pk,
		OperationType: event.Operation,
		Type:          webhookEvent,

		OriginalRaw: rawData,
	}, nil
}

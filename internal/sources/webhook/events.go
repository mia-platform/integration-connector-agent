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
	"net/http"

	"github.com/mia-platform/integration-connector-agent/entities"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var (
	ErrMissingFieldID = fmt.Errorf("missing id field in event")
)

type EventTypeParam struct {
	Data    gjson.Result
	Headers http.Header
}

type Events struct {
	Supported    map[string]Event
	GetEventType func(data EventTypeParam) string
	PayloadKey   ContentTypeConfig
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

func GetEventTypeByPath(path string) func(data EventTypeParam) string {
	return func(data EventTypeParam) string {
		if data.Data.Exists() {
			return data.Data.Get(path).String()
		}
		return ""
	}
}

type RequestInfo struct {
	data    []byte
	headers http.Header
}

func (e *Events) getPipelineEvent(logger *logrus.Entry, requestInfo RequestInfo) (entities.PipelineEvent, error) {
	rawData := requestInfo.data

	parsed := gjson.ParseBytes(rawData)
	webhookEvent := e.GetEventType(EventTypeParam{
		Data:    parsed,
		Headers: requestInfo.headers,
	})

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

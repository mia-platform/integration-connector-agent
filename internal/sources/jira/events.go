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

package jira

import (
	"fmt"

	"github.com/mia-platform/integration-connector-agent/internal/entities"

	"github.com/tidwall/gjson"
)

type Events map[string]Event

type Event struct {
	Operation entities.Operation
	FieldID   string
}

const (
	issueCreated     = "jira:issue_created"
	issueUpdated     = "jira:issue_updated"
	issueDeleted     = "jira:issue_deleted"
	issueEventIDPath = "issue.id"

	webhookEventPath = "webhookEvent"
)

var DefaultSupportedEvents = Events{
	issueCreated: {
		Operation: entities.Write,
		FieldID:   issueEventIDPath,
	},
	issueUpdated: {
		Operation: entities.Write,
		FieldID:   issueEventIDPath,
	},
	issueDeleted: {
		Operation: entities.Delete,
		FieldID:   issueEventIDPath,
	},
}

func (e Events) getPipelineEvent(rawData []byte) (entities.PipelineEvent, error) {
	parsed := gjson.ParseBytes(rawData)
	webhookEvent := parsed.Get(webhookEventPath).String()

	event, ok := e[webhookEvent]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedWebhookEvent, webhookEvent)
	}

	return &entities.Event{
		ID:            parsed.Get(event.FieldID).String(),
		OperationType: event.Operation,

		OriginalRaw: rawData,
	}, nil

}

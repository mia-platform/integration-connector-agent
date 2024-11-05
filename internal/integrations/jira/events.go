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

	"github.com/mia-platform/data-connector-agent/internal/entities"
	"github.com/tidwall/gjson"
)

const (
	issueCreated = "jira:issue_created"
	issueUpdated = "jira:issue_updated"
	issueDeleted = "jira:issue_deleted"
)

func getPipelineEvent(rawData []byte) (entities.PipelineEvent, error) {
	parsed := gjson.ParseBytes(rawData)
	id := parsed.Get("issue.id").String()
	webhookEvent := parsed.Get("webhookEvent").String()

	operationType, err := getOperationType(webhookEvent)
	if err != nil {
		return nil, err
	}

	return &entities.Event{
		ID:            id,
		OperationType: operationType,

		OriginalRaw:    rawData,
		OriginalParsed: parsed,
	}, nil
}

func getOperationType(event string) (entities.Operation, error) {
	switch event {
	case issueCreated:
		fallthrough
	case issueUpdated:
		return entities.Write, nil
	case issueDeleted:
		return entities.Delete, nil
	}
	return 0, fmt.Errorf("%w: %s", ErrUnsupportedWebhookEvent, event)
}

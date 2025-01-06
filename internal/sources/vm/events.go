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

package vm

import (
	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
)

const (
	SentinelMetrics = "sentinel:metrics"
	ProcessSignal   = "process:signal"   // TO BE IMPLEMENTED
	ProcessWatch    = "process:watch"    // TO BE IMPLEMENTED
	SystemException = "system:exception" // TO BE IMPLEMENTED
	SentinelStatus  = "sentinel:status"

	processPath  = "payload.ID"
	systemIDPath = "payload.hostID"
	sentinelPath = "sentinelID"
)

var DefaultSupportedEvents = webhook.Events{
	Supported: map[string]webhook.Event{
		SentinelMetrics: {
			Operation: entities.Write,
			FieldID:   systemIDPath,
		},
		ProcessSignal: {
			Operation: entities.Write,
			FieldID:   processPath,
		},
		ProcessWatch: {
			Operation: entities.Write,
			FieldID:   processPath,
		},
		SystemException: {
			Operation: entities.Write,
			FieldID:   systemIDPath,
		},
		SentinelStatus: {
			Operation: entities.Write,
			FieldID:   sentinelPath,
		},
	},
	EventTypeFieldPath: webhookEventPath,
}

// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
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

package azure

import (
	"strings"

	"github.com/mia-platform/integration-connector-agent/entities"
)

type ActivityLogEventData struct {
	Records []*ActivityLogEventRecord `json:"records,omitempty"`
}

type ActivityLogEventRecord struct {
	RoleLocation    string         `json:"RoleLocation,omitempty"`   //nolint:tagliatelle
	Stamp           string         `json:"Stamp,omitempty"`          //nolint:tagliatelle
	ReleaseVersion  string         `json:"ReleaseVersion,omitempty"` //nolint:tagliatelle
	Time            string         `json:"time,omitempty"`
	ResourceID      string         `json:"resourceId,omitempty"`
	OperationName   string         `json:"operationName,omitempty"`
	Category        string         `json:"category,omitempty"`
	ResultType      string         `json:"resultType,omitempty"`
	ResultSignature string         `json:"resultSignature,omitempty"`
	DurationMs      string         `json:"durationMs,omitempty"`
	CallerIPAddress string         `json:"callerIpAddress,omitempty"`
	CorrelationID   string         `json:"correlationId,omitempty"`
	Level           string         `json:"level,omitempty"`
	Properties      map[string]any `json:"properties,omitempty"`
}

func (r *ActivityLogEventRecord) entityOperationType() entities.Operation {
	if strings.HasSuffix(strings.ToLower(r.OperationName), "delete") ||
		strings.HasSuffix(strings.ToLower(r.OperationName), "delete/action") {
		return entities.Delete
	}

	return entities.Write
}

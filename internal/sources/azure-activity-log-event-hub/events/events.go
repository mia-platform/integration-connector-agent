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

package azureactivitylogeventhubevents

type ActivityLogEventData struct {
	Records []*ActivityLogEventRecord `json:"records"`
}

type ActivityLogEventRecord struct {
	RoleLocation    string                    `json:"RoleLocation"`   //nolint:tagliatelle
	Stamp           string                    `json:"Stamp"`          //nolint:tagliatelle
	ReleaseVersion  string                    `json:"ReleaseVersion"` //nolint:tagliatelle
	Time            string                    `json:"time"`
	ResourceID      string                    `json:"resourceId"`
	OperationName   string                    `json:"operationName"`
	Category        string                    `json:"category"`
	ResultType      string                    `json:"resultType"`
	ResultSignature string                    `json:"resultSignature"`
	DurationMs      string                    `json:"durationMs"`
	CallerIPAddress string                    `json:"callerIpAddress"`
	CorrelationID   string                    `json:"correlationId"`
	Identity        *ActivityLogEventIdentity `json:"identity"`
	Level           string                    `json:"level"`
	Properties      map[string]any            `json:"properties"`
}

type ActivityLogEventIdentity struct {
	Authorization *ActivityLogAuthorization `json:"authorization"`
	Claims        map[string]string         `json:"claims"`
}

type ActivityLogAuthorization struct {
	Scope    string                `json:"scope"`
	Action   string                `json:"action"`
	Evidence AuthorizationEvidence `json:"evidence"`
}

type AuthorizationEvidence struct {
	Role                string `json:"role"`
	RoleAssignmentScope string `json:"roleAssignmentScope"`
	RoleAssignmentID    string `json:"roleAssignmentId"`
	RoleDefinitionID    string `json:"roleDefinitionId"`
	PrincipalID         string `json:"principalId"`
	PrincipalType       string `json:"principalType"`
}

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

type jiraIssue struct {
	ID     string         `json:"id"`
	Self   string         `json:"self"`
	Key    string         `json:"key"`
	Fields map[string]any `json:"fields"`
}

type jiraUser struct {
	Name        string `json:"name"`
	Email       string `json:"emailAddress"`
	DisplayName string `json:"displayName"`
	Active      string `json:"active"`
}

type issueChange struct {
	ToString   string         `json:"toString"`
	To         map[string]any `json:"to"`
	FromString string         `json:"fromString"`
	From       map[string]any `json:"from"`
	FieldType  string         `json:"fieldtype"`
	Field      string         `json:"field"`
}

type changelog struct {
	ID    int64                    `json:"id"`
	Items []map[string]issueChange `json:"items"`
}

type comment struct{}

type jiraIssueEvent struct {
	ID           int64     `json:"id"`
	Timestamp    int64     `json:"timestamp"`
	Issue        jiraIssue `json:"issue"`
	User         jiraUser  `json:"user,omitempty"`
	WebhookEvent string    `json:"webhookEvent"`
	Changelog    changelog `json:"changelog,omitempty"`
	Comment      comment   `json:"comment,omitempty"`
}

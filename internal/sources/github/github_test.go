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

package github

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/entities"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/processors"
	fakewriter "github.com/mia-platform/integration-connector-agent/internal/sinks/fake"
	"github.com/mia-platform/integration-connector-agent/internal/sources/webhook"
	"github.com/mia-platform/integration-connector-agent/internal/testutils"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	t.Run("unmarshal config", func(t *testing.T) {
		t.Setenv("GITHUB_WEBHOOK_SECRET", "GITHUB_WEBHOOK_SECRET")
		t.Cleanup(func() { os.Unsetenv("GITHUB_WEBHOOK_SECRET") })

		rawConfig, err := os.ReadFile("testdata/config.json")
		require.NoError(t, err)

		actual := &Config{}
		require.NoError(t, json.Unmarshal(rawConfig, actual))
		require.NoError(t, actual.Validate())

		require.Equal(t, &Config{
			WebhookPath: "/github/webhook",
			Authentication: webhook.HMAC{
				HeaderName: "X-Hub-Signature-256",
				Secret:     "GITHUB_WEBHOOK_SECRET",
			},
		}, actual)
	})
}

func TestAddSourceToRouter_PullRequest(t *testing.T) {
	t.Setenv("GITHUB_WEBHOOK_SECRET", "GITHUB_WEBHOOK_SECRET")
	t.Cleanup(func() { os.Unsetenv("GITHUB_WEBHOOK_SECRET") })

	logger, _ := test.NewNullLogger()
	ctx := context.Background()

	rawConfig, err := os.ReadFile("testdata/config.json")
	require.NoError(t, err)

	app, router := testutils.GetTestRouter(t)
	proc := &processors.Processors{}
	s := fakewriter.New(nil)
	p1, err := pipeline.New(logger, proc, s)
	require.NoError(t, err)
	pg := pipeline.NewGroup(logger, p1)

	err = AddSourceToRouter(ctx, rawConfig, pg, router)
	require.NoError(t, err)

	t.Run("pull_request opened event", func(t *testing.T) {
		body, err := os.ReadFile("testdata/pull_request_opened.json")
		require.NoError(t, err)
		req := getWebhookRequest(bytes.NewBuffer(body))
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Eventually(t, func() bool {
			return len(s.Calls()) == 1
		}, 1*time.Second, 10*time.Millisecond)
		call := s.Calls().LastCall()
		require.Equal(t, entities.Write, call.Operation)
		actualEvent, ok := call.Data.(*entities.Event)
		require.True(t, ok)
		require.Equal(t, entities.PkFields{{Key: "pull_request.id", Value: "123456"}}, actualEvent.PrimaryKeys)
		require.Equal(t, "pull_request", actualEvent.Type)
		require.Equal(t, entities.Write, actualEvent.OperationType)
		// Check that the injected event type is present in the raw payload
		require.Contains(t, string(actualEvent.OriginalRaw), "\"_github_event_type\":\"pull_request\"")
	})
}

func getWebhookRequest(body *bytes.Buffer) *http.Request {
	secret := "GITHUB_WEBHOOK_SECRET"
	hmac := getHMACValidationHeader(secret, body.Bytes())
	req := httptest.NewRequest(http.MethodPost, "/github/webhook", body)
	req.Header.Add("X-Hub-Signature-256", fmt.Sprintf("sha256=%s", hmac))
	req.Header.Add("X-GitHub-Event", "pull_request")
	return req
}

func getHMACValidationHeader(secret string, body []byte) string {
	hasher := hmac.New(sha256.New, []byte(secret))
	hasher.Write(body)
	return hex.EncodeToString(hasher.Sum(nil))
}

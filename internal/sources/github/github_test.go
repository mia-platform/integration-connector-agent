// Copyright Mia srl
// SPDX-License-Identifier: Apache-2.0

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
		require.Equal(t, fakewriter.Call{
			Operation: entities.Write,
			Data: &entities.Event{
				PrimaryKeys:   entities.PkFields{{Key: "pull_request.id", Value: "123456"}},
				Type:          "pull_request",
				OperationType: entities.Write,
				OriginalRaw:   body,
			},
		}, call)
	})
}

func getWebhookRequest(body *bytes.Buffer) *http.Request {
	secret := "GITHUB_WEBHOOK_SECRET"
	hmac := getHMACValidationHeader(secret, body.Bytes())
	req := httptest.NewRequest(http.MethodPost, "/github/webhook", body)
	req.Header.Add("X-Hub-Signature-256", fmt.Sprintf("sha256=%s", hmac))
	return req
}

func getHMACValidationHeader(secret string, body []byte) string {
	hasher := hmac.New(sha256.New, []byte(secret))
	hasher.Write(body)
	return hex.EncodeToString(hasher.Sum(nil))
}

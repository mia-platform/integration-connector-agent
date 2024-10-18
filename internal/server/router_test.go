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

package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/data-connector-agent/internal/config"
	integration "github.com/mia-platform/data-connector-agent/internal/integrations"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestSetupRouter(t *testing.T) {
	log, _ := test.NewNullLogger()
	env := config.EnvironmentVariables{
		HTTPPort:      "3000",
		ServicePrefix: "/my-prefix",
		ServiceType:   integration.Jira,
	}

	app, err := NewRouter(env, log)
	require.NoError(t, err, "unexpected error")

	t.Run("API documentation is correctly exposed without prefix - json", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/documentations/json", nil)
		response, err := app.Test(request)
		require.NoError(t, err)
		defer response.Body.Close()

		require.Equal(t, fiber.StatusOK, response.StatusCode, "The response statusCode should be 200")
		body, readBodyError := io.ReadAll(response.Body)
		require.NoError(t, readBodyError)
		require.True(t, string(body) != "", "The response body should not be an empty string")
	})

	t.Run("API documentation is correctly exposed without prefix - yaml", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/documentations/yaml", nil)
		response, err := app.Test(request)
		require.NoError(t, err)
		defer response.Body.Close()

		require.Equal(t, fiber.StatusOK, response.StatusCode, "The response statusCode should be 200")
		body, readBodyError := io.ReadAll(response.Body)
		require.NoError(t, readBodyError)
		require.True(t, string(body) != "", "The response body should not be an empty string")
	})
}

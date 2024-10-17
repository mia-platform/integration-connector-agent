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

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestStatusRoutes(t *testing.T) {
	app := fiber.New()
	serviceName := "my-service-name"
	serviceVersion := "0.0.0"
	StatusRoutes(app, serviceName, serviceVersion)

	t.Run("/-/healthz - ok", func(t *testing.T) {
		expectedResponse := fmt.Sprintf("{\"status\":\"OK\",\"name\":\"%s\",\"version\":\"%s\"}", serviceName, serviceVersion)
		request := httptest.NewRequest(http.MethodGet, "/-/healthz", nil)
		response, err := app.Test(request)
		require.NoError(t, err)

		defer response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode, "The response statusCode should be 200")
		body, readBodyError := io.ReadAll(response.Body)
		require.NoError(t, readBodyError)
		require.Equal(t, expectedResponse, string(body), "The response body should be the expected one")
	})

	t.Run("/-/ready - ok", func(t *testing.T) {
		expectedResponse := fmt.Sprintf("{\"status\":\"OK\",\"name\":\"%s\",\"version\":\"%s\"}", serviceName, serviceVersion)
		request := httptest.NewRequest(http.MethodGet, "/-/ready", nil)
		response, err := app.Test(request)
		require.NoError(t, err)

		defer response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode, "The response statusCode should be 200")
		body, readBodyError := io.ReadAll(response.Body)
		require.NoError(t, readBodyError)
		require.Equal(t, expectedResponse, string(body), "The response body should be the expected one")
	})

	t.Run("/-/check-up - ok", func(t *testing.T) {
		expectedResponse := fmt.Sprintf("{\"status\":\"OK\",\"name\":\"%s\",\"version\":\"%s\"}", serviceName, serviceVersion)
		request := httptest.NewRequest(http.MethodGet, "/-/check-up", nil)
		response, err := app.Test(request)
		require.NoError(t, err)

		defer response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode, "The response statusCode should be 200")
		body, readBodyError := io.ReadAll(response.Body)
		require.NoError(t, readBodyError)
		require.Equal(t, expectedResponse, string(body), "The response body should be the expected one")
	})
}

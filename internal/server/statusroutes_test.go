// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mia-platform/integration-connector-agent/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestStatusRoutes(t *testing.T) {
	app := fiber.New()
	serviceName := "my-service-name"
	serviceVersion := utils.Version
	statusRoutes(app, serviceName, serviceVersion)

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

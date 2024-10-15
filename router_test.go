package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"data-connector-agent/config"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
)

func TestSetupRouter(t *testing.T) {
	log, _ := test.NewNullLogger()
	env := config.EnvironmentVariables{
		ServiceVersion: "my-version",
		HTTPPort:       "3000",
		ServicePrefix:  "my-prefix",
	}

	app, err := setupRouter(env, log)
	require.NoError(t, err, "unexpected error")

	t.Run("API documentation is correctly exposed without prefix - json", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/documentations/json", nil)
		response, err := app.Test(request)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, response.StatusCode, "The response statusCode should be 200")

		body, readBodyError := io.ReadAll(response.Body)
		require.NoError(t, readBodyError)
		require.True(t, string(body) != "", "The response body should not be an empty string")
	})

	t.Run("API documentation is correctly exposed without prefix - yaml", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/documentations/yaml", nil)
		response, err := app.Test(request)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, response.StatusCode, "The response statusCode should be 200")

		body, readBodyError := io.ReadAll(response.Body)
		require.NoError(t, readBodyError)
		require.True(t, string(body) != "", "The response body should not be an empty string")
	})
}

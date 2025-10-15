// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

package testutils

import (
	"testing"

	swagger "github.com/davidebianchi/gswagger"
	oasfiber "github.com/davidebianchi/gswagger/support/fiber"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func GetTestRouter(t *testing.T) (*fiber.App, *swagger.Router[fiber.Handler, fiber.Router]) {
	t.Helper()

	app := fiber.New()
	router, err := swagger.NewRouter(oasfiber.NewRouter(app), swagger.Options{
		Openapi: &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test",
				Version: "test-version",
			},
		},
	})
	require.NoError(t, err)

	return app, router
}

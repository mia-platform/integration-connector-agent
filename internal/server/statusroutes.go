// Copyright Mia srl
// SPDX-License-Identifier: AGPL-3.0-only or Commercial

package server

import (
	"github.com/gofiber/fiber/v2"
)

// statusResponse type.
type statusResponse struct {
	Status  string `json:"status"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// statusRoutes add status routes to router.
func statusRoutes(app *fiber.App, serviceName, serviceVersion string) {
	app.Get("/-/healthz", func(c *fiber.Ctx) error {
		status := statusResponse{
			Status:  "OK",
			Name:    serviceName,
			Version: serviceVersion,
		}
		return c.JSON(status)
	})

	app.Get("/-/ready", func(c *fiber.Ctx) error {
		status := statusResponse{
			Status:  "OK",
			Name:    serviceName,
			Version: serviceVersion,
		}
		return c.JSON(status)
	})

	app.Get("/-/check-up", func(c *fiber.Ctx) error {
		status := statusResponse{
			Status:  "OK",
			Name:    serviceName,
			Version: serviceVersion,
		}
		return c.JSON(status)
	})
}

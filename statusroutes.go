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
	"github.com/gofiber/fiber/v2"
)

// StatusResponse type.
type StatusResponse struct {
	Status  string `json:"status"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// StatusRoutes add status routes to router.
func StatusRoutes(app *fiber.App, serviceName, serviceVersion string) {
	app.Get("/-/healthz", func(c *fiber.Ctx) error {
		status := StatusResponse{
			Status:  "OK",
			Name:    serviceName,
			Version: serviceVersion,
		}
		return c.JSON(status)
	})

	app.Get("/-/ready", func(c *fiber.Ctx) error {
		status := StatusResponse{
			Status:  "OK",
			Name:    serviceName,
			Version: serviceVersion,
		}
		return c.JSON(status)
	})

	app.Get("/-/check-up", func(c *fiber.Ctx) error {
		status := StatusResponse{
			Status:  "OK",
			Name:    serviceName,
			Version: serviceVersion,
		}
		return c.JSON(status)
	})
}

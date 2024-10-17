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
	"context"
	"fmt"
	"path"

	"data-connector-agent/config"

	swagger "github.com/davidebianchi/gswagger"
	oasfiber "github.com/davidebianchi/gswagger/support/fiber"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
	middleware "github.com/mia-platform/glogger/v4/middleware/fiber"
	"github.com/sirupsen/logrus"
)

func setupRouter(env config.EnvironmentVariables, log *logrus.Logger) (*fiber.App, error) {
	app := fiber.New()

	middlewareLog := glogrus.GetLogger(logrus.NewEntry(log))
	app.Use(middleware.RequestMiddlewareLogger[*logrus.Entry](middlewareLog, []string{"/-/"}))
	StatusRoutes(app, "data-connector-agent", env.ServiceVersion)
	if env.ServicePrefix != "" && env.ServicePrefix != "/" {
		log.WithField("servicePrefix", env.ServicePrefix).Trace("applying service prefix")
		app.Use(pprof.New(pprof.Config{Prefix: fmt.Sprintf("%s/", path.Clean(env.ServicePrefix))}))
	}

	oasRouter, err := swagger.NewRouter(oasfiber.NewRouter(app), swagger.Options{
		Context: context.Background(),
		Openapi: &openapi3.T{
			Info: &openapi3.Info{
				Title:   "data-connector-agent",
				Version: env.ServiceVersion,
			},
		},
		JSONDocumentationPath: "/documentations/json",
		YAMLDocumentationPath: "/documentations/yaml",
		PathPrefix:            env.ServicePrefix,
	})

	if err != nil {
		return nil, err
	}

	// TODO: add here your routes

	if err = oasRouter.GenerateAndExposeOpenapi(); err != nil {
		return nil, err
	}

	return app, nil
}

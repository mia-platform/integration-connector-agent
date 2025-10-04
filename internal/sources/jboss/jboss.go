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

package jboss

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mia-platform/integration-connector-agent/internal/config"
	"github.com/mia-platform/integration-connector-agent/internal/pipeline"
	"github.com/mia-platform/integration-connector-agent/internal/sources"

	swagger "github.com/davidebianchi/gswagger"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

const (
	defaultWildFlyURL      = "http://localhost:9990/management"
	defaultPollingInterval = time.Second
	defaultUsername        = "admin"
)

// Duration is a custom type that can unmarshal from JSON strings
type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

type Config struct {
	WildFlyURL      string              `json:"wildFlyUrl,omitempty"`
	Username        string              `json:"username,omitempty"`
	Password        config.SecretSource `json:"password"`
	PollingInterval Duration            `json:"pollingInterval,omitempty"`
}

func (c *Config) Validate() error {
	if c.Password.String() == "" {
		return fmt.Errorf("password must be provided")
	}
	return nil
}

func (c *Config) withDefaults() {
	if c.WildFlyURL == "" {
		c.WildFlyURL = defaultWildFlyURL
	}
	if c.Username == "" {
		c.Username = defaultUsername
	}
	if c.PollingInterval == 0 {
		c.PollingInterval = Duration(defaultPollingInterval)
	}
}

type JBossSource struct {
	ctx      context.Context
	log      *logrus.Logger
	config   *Config
	pipeline pipeline.IPipelineGroup
	client   *JBossClient
	done     chan struct{}
}

func NewJBossSource(
	ctx context.Context,
	log *logrus.Logger,
	cfg config.GenericConfig,
	pipeline pipeline.IPipelineGroup,
	oasRouter *swagger.Router[fiber.Handler, fiber.Router],
) (sources.CloseableSource, error) {
	config, err := config.GetConfig[*Config](cfg)
	if err != nil {
		return nil, err
	}

	config.withDefaults()
	if err := config.Validate(); err != nil {
		return nil, err
	}

	client, err := NewJBossClient(config.WildFlyURL, config.Username, config.Password.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create JBoss client: %w", err)
	}

	source := &JBossSource{
		ctx:      ctx,
		log:      log,
		config:   config,
		pipeline: pipeline,
		client:   client,
		done:     make(chan struct{}),
	}

	source.start()

	return source, nil
}

func (s *JBossSource) start() {
	s.pipeline.Start(s.ctx)

	// Start the polling goroutine
	go s.pollWildFly()
}

func (s *JBossSource) pollWildFly() {
	ticker := time.NewTicker(s.config.PollingInterval.Duration())
	defer ticker.Stop()

	s.log.WithFields(logrus.Fields{
		"wildflyUrl":      s.config.WildFlyURL,
		"pollingInterval": s.config.PollingInterval.Duration().String(),
	}).Info("Starting JBoss/WildFly polling")

	// Poll immediately on start
	s.performPoll()

	for {
		select {
		case <-ticker.C:
			s.performPoll()
		case <-s.ctx.Done():
			s.log.Info("Context cancelled, stopping JBoss polling")
			return
		case <-s.done:
			s.log.Info("JBoss source closed, stopping polling")
			return
		}
	}
}

func (s *JBossSource) performPoll() {
	timestamp := time.Now()
	s.log.Debug("Starting JBoss/WildFly poll cycle")

	// Log the request being made
	s.log.WithFields(logrus.Fields{
		"wildflyUrl": s.config.WildFlyURL,
		"username":   s.config.Username,
		"timestamp":  timestamp,
	}).Debug("Making request to JBoss/WildFly management API")

	deployments, err := s.client.GetDeployments(s)
	if err != nil {
		s.log.WithError(err).Error("Failed to get deployments from WildFly")
		return
	}

	// Log successful response
	s.log.WithFields(logrus.Fields{
		"deploymentsCount": len(deployments),
	}).Debug("Successfully received deployments from JBoss/WildFly")

	if len(deployments) == 0 {
		s.log.Debug("No deployments found in JBoss/WildFly response")
		return
	}

	// Process each deployment
	for i, deployment := range deployments {
		s.log.WithFields(logrus.Fields{
			"deploymentIndex": i,
			"deployment":      deployment,
		}).Debug("Processing deployment from JBoss/WildFly response")

		event, err := CreateDeploymentEvent(deployment, timestamp)
		if err != nil {
			s.log.WithError(err).WithField("deploymentName", deployment.Name).
				Warn("Failed to create event for deployment")
			continue
		}

		s.log.WithFields(logrus.Fields{
			"deploymentName":   deployment.Name,
			"deploymentStatus": deployment.Status,
			"eventPrimaryKeys": event.GetPrimaryKeys(),
			"eventType":        event.GetType(),
			"operation":        event.Operation().String(),
		}).Debug("Created deployment event, sending to pipeline/sink")

		s.pipeline.AddMessage(event)
		s.log.WithField("deploymentName", deployment.Name).Debug("Deployment event sent to pipeline/sink successfully")
	}

	s.log.WithFields(logrus.Fields{
		"processedDeployments": len(deployments),
	}).Debug("Completed JBoss/WildFly poll cycle")
}

func (s *JBossSource) Close() error {
	close(s.done)
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

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
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	glogrus "github.com/mia-platform/glogger/v4/loggers/logrus"
)

// JBossClient handles HTTP digest authentication with JBoss/WildFly management interface
type JBossClient struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewJBossClient creates a new JBoss client with digest authentication
func NewJBossClient(baseURL, username, password string) (*JBossClient, error) {
	if baseURL == "" || username == "" || password == "" {
		return nil, fmt.Errorf("baseURL, username, and password are required")
	}

	return &JBossClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// WildFlyPayload represents the management API request payload
type WildFlyPayload struct {
	Operation      string      `json:"operation"`
	Address        interface{} `json:"address"`
	Recursive      bool        `json:"recursive,omitempty"`
	IncludeRuntime bool        `json:"include-runtime,omitempty"`
	JSONPretty     int         `json:"json.pretty"`
}

// WildFlyResponse represents the management API response
type WildFlyResponse struct {
	Outcome            string      `json:"outcome"`
	Result             interface{} `json:"result,omitempty"`
	FailureDescription string      `json:"failure-description,omitempty"`
}

// Deployment represents a JBoss/WildFly deployment with all available information
type Deployment struct {
	Name          string                 `json:"name"`
	RuntimeName   string                 `json:"runtimeName"`
	Status        string                 `json:"status"`
	Enabled       bool                   `json:"enabled"`
	Persistent    bool                   `json:"persistent"`
	Content       []DeploymentContent    `json:"content,omitempty"`
	Subdeployment interface{}            `json:"subdeployment,omitempty"`
	Subsystem     map[string]interface{} `json:"subsystem,omitempty"`
}

// DeploymentContent represents the content hash information
type DeploymentContent struct {
	Hash map[string]string `json:"hash,omitempty"`
}

// UndertowSubsystem represents Undertow web server subsystem information
type UndertowSubsystem struct {
	ActiveSessions  int                    `json:"active-sessions"`
	ContextRoot     string                 `json:"context-root"`
	Server          string                 `json:"server"`
	SessionsCreated int                    `json:"sessions-created"`
	VirtualHost     string                 `json:"virtual-host"`
	Servlet         map[string]ServletInfo `json:"servlet,omitempty"`
}

// ServletInfo represents servlet information
type ServletInfo struct {
	MaxRequestTime   int    `json:"max-request-time"`
	MinRequestTime   int    `json:"min-request-time"`
	RequestCount     int    `json:"request-count"`
	ServletClass     string `json:"servlet-class"`
	ServletName      string `json:"servlet-name"`
	TotalRequestTime int    `json:"total-request-time"`
}

// GetDeployments retrieves all deployments from WildFly
func (c *JBossClient) GetDeployments(s *JBossSource) ([]Deployment, error) {

	payload := WildFlyPayload{
		Operation:      "read-resource",
		Address:        []string{"deployment", "*"}, // use [] for all or you can specify ["subsystem", "datasources", "data-source", "*"]
		Recursive:      true,
		IncludeRuntime: true,
		JSONPretty:     1,
	}

	s.log.WithFields(map[string]interface{}{
		"baseURL":   c.baseURL,
		"username":  c.username,
		"operation": payload.Operation,
		"address":   payload.Address,
	}).Debug("JBoss client: preparing to make management API request")

	response, err := c.makeRequest(s.ctx, payload)
	if err != nil {
		s.log.WithError(err).Debug("JBoss client: management API request failed")
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	s.log.WithFields(map[string]interface{}{
		"response": response.Result,
	}).Debug("JBoss client: received response from management API")

	if response.Outcome != "success" {
		s.log.WithField("failureDescription", response.FailureDescription).Debug("JBoss client: management API returned failure")
		return nil, fmt.Errorf("WildFly request failed: %s", response.FailureDescription)
	}

	var deployments []Deployment

	// Handle array of deployment results
	switch result := response.Result.(type) {
	case []interface{}:
		s.log.WithField("deploymentCount", len(result)).Debug("JBoss client: processing deployment array")

		for i, item := range result {
			itemData, ok := item.(map[string]interface{})
			if !ok {
				s.log.WithField("itemIndex", i).Debug("JBoss client: skipping invalid item data")
				continue
			}

			// Extract address array
			addressData, ok := itemData["address"].([]interface{})
			if !ok || len(addressData) == 0 {
				s.log.WithField("itemIndex", i).Debug("JBoss client: no address data found")
				continue
			}

			// Extract deployment name from address
			var deploymentName string
			for _, addr := range addressData {
				if addrMap, ok := addr.(map[string]interface{}); ok {
					if deployment, exists := addrMap["deployment"]; exists {
						if name, ok := deployment.(string); ok {
							deploymentName = name
							break
						}
					}
				}
			}

			if deploymentName == "" {
				s.log.WithField("itemIndex", i).Debug("JBoss client: no deployment name found in address")
				continue
			}

			// Extract deployment details from result
			resultData, ok := itemData["result"].(map[string]interface{})
			if !ok {
				s.log.WithField("deploymentName", deploymentName).Debug("JBoss client: no result data found")
				continue
			}

			deployment := Deployment{
				Name:        deploymentName,
				RuntimeName: getStringValue(resultData, "runtime-name"),
				Status:      getStringValue(resultData, "status"),
				Enabled:     getBoolValue(resultData, "enabled"),
				Persistent:  getBoolValue(resultData, "persistent"),
			}

			// Parse content array if available
			if contentData, ok := resultData["content"].([]interface{}); ok {
				deployment.Content = parseContentArray(contentData)
			}

			// Parse subdeployment if available
			if subdeployment, ok := resultData["subdeployment"]; ok {
				deployment.Subdeployment = subdeployment
			}

			// Parse subsystem information if available
			if subsystemData, ok := resultData["subsystem"].(map[string]interface{}); ok {
				deployment.Subsystem = subsystemData

				// Log specific subsystem information for debugging
				if undertowData, ok := subsystemData["undertow"].(map[string]interface{}); ok {
					s.log.WithFields(map[string]interface{}{
						"deploymentName":  deploymentName,
						"contextRoot":     getStringValue(undertowData, "context-root"),
						"activeSessions":  getIntValue(undertowData, "active-sessions"),
						"sessionsCreated": getIntValue(undertowData, "sessions-created"),
						"server":          getStringValue(undertowData, "server"),
						"virtualHost":     getStringValue(undertowData, "virtual-host"),
					}).Debug("JBoss client: parsed Undertow subsystem information")

					// Log servlet information if available
					if servletData, ok := undertowData["servlet"].(map[string]interface{}); ok {
						for servletName, servletInfo := range servletData {
							if servletMap, ok := servletInfo.(map[string]interface{}); ok {
								s.log.WithFields(map[string]interface{}{
									"deploymentName":   deploymentName,
									"servletName":      servletName,
									"servletClass":     getStringValue(servletMap, "servlet-class"),
									"requestCount":     getIntValue(servletMap, "request-count"),
									"maxRequestTime":   getIntValue(servletMap, "max-request-time"),
									"minRequestTime":   getIntValue(servletMap, "min-request-time"),
									"totalRequestTime": getIntValue(servletMap, "total-request-time"),
								}).Debug("JBoss client: parsed servlet information")
							}
						}
					}
				}
			}

			s.log.WithFields(map[string]interface{}{
				"deploymentName":       deployment.Name,
				"deploymentStatus":     deployment.Status,
				"deploymentEnabled":    deployment.Enabled,
				"deploymentPersistent": deployment.Persistent,
				"deploymentSubsystem":  deployment.Subsystem != nil,
				"deploymentContent":    len(deployment.Content),
			}).Debug("JBoss client: parsed deployment with enhanced mapping")

			deployments = append(deployments, deployment)
		}
	default:
		s.log.WithField("resultType", fmt.Sprintf("%T", response.Result)).Debug("JBoss client: unexpected result type")
		return nil, fmt.Errorf("unexpected result type: %T", response.Result)
	}

	s.log.WithField("totalDeployments", len(deployments)).Debug("JBoss client: successfully retrieved deployments")
	return deployments, nil
}

// makeRequest performs an HTTP request with digest authentication
func (c *JBossClient) makeRequest(ctx context.Context, payload WildFlyPayload) (*WildFlyResponse, error) {
	log := glogrus.FromContext(ctx)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.WithError(err).Debug("JBoss client: failed to marshal request payload")
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	log.WithFields(map[string]interface{}{
		"url":         c.baseURL,
		"payloadSize": len(payloadBytes),
		"payload":     string(payloadBytes),
	}).Debug("JBoss client: making initial HTTP request")

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(payloadBytes))
	if err != nil {
		log.WithError(err).Debug("JBoss client: failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Log the raw request details
	log.WithFields(map[string]interface{}{
		"method":      req.Method,
		"url":         req.URL.String(),
		"headers":     req.Header,
		"bodySize":    len(payloadBytes),
		"requestBody": string(payloadBytes),
	}).Debug("JBoss client: raw initial HTTP request details")

	// First request to get the WWW-Authenticate header
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.WithError(err).Debug("JBoss client: initial HTTP request failed")
		return nil, fmt.Errorf("failed to make initial request: %w", err)
	}
	resp.Body.Close()

	// Log the raw response details
	log.WithFields(map[string]interface{}{
		"statusCode":      resp.StatusCode,
		"status":          resp.Status,
		"responseHeaders": resp.Header,
		"proto":           resp.Proto,
	}).Debug("JBoss client: raw initial HTTP response details")

	if resp.StatusCode != http.StatusUnauthorized {
		log.WithField("expectedStatus", http.StatusUnauthorized).Debug("JBoss client: expected 401 Unauthorized for digest auth")
		return nil, fmt.Errorf("expected 401 Unauthorized, got %d", resp.StatusCode)
	}

	// Parse WWW-Authenticate header for digest challenge
	authHeader := resp.Header.Get("WWW-Authenticate")
	if authHeader == "" {
		log.Debug("JBoss client: no WWW-Authenticate header found")
		return nil, fmt.Errorf("no WWW-Authenticate header found")
	}

	log.WithField("authHeader", authHeader).Debug("JBoss client: parsing digest authentication challenge")

	digestAuth, err := c.parseDigestAuth(authHeader)
	if err != nil {
		log.WithError(err).Debug("JBoss client: failed to parse digest auth challenge")
		return nil, fmt.Errorf("failed to parse digest auth: %w", err)
	}

	log.Debug("JBoss client: creating authenticated HTTP request")

	// Create new request with digest authentication
	req, err = http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(payloadBytes))
	if err != nil {
		log.WithError(err).Debug("JBoss client: failed to create authenticated request")
		return nil, fmt.Errorf("failed to create authenticated request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	authHeaderValue, err := c.createDigestAuthHeader(digestAuth, "POST", "/management")
	if err != nil {
		log.WithError(err).Debug("JBoss client: failed to create digest auth header")
		return nil, fmt.Errorf("failed to create auth header: %w", err)
	}
	req.Header.Set("Authorization", authHeaderValue)

	// Log the raw authenticated request details
	log.WithFields(map[string]interface{}{
		"method":           req.Method,
		"url":              req.URL.String(),
		"headers":          req.Header,
		"bodySize":         len(payloadBytes),
		"requestBody":      string(payloadBytes),
		"authHeaderLength": len(authHeaderValue),
		"digestAuthRealm":  digestAuth.realm,
		"digestAuthNonce":  digestAuth.nonce,
	}).Debug("JBoss client: raw authenticated HTTP request details")

	log.Debug("JBoss client: making authenticated HTTP request")

	// Make authenticated request
	resp, err = c.httpClient.Do(req)
	if err != nil {
		log.WithError(err).Debug("JBoss client: authenticated HTTP request failed")
		return nil, fmt.Errorf("failed to make authenticated request: %w", err)
	}
	defer resp.Body.Close()

	// Log raw response details
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		responseHeaders[key] = strings.Join(values, ", ")
	}

	log.WithFields(map[string]interface{}{
		"statusCode":      resp.StatusCode,
		"status":          resp.Status,
		"contentLength":   resp.ContentLength,
		"contentType":     resp.Header.Get("Content-Type"),
		"responseHeaders": responseHeaders,
		"protocol":        resp.Proto,
	}).Debug("JBoss client: received authenticated HTTP response")

	if resp.StatusCode != http.StatusOK {
		log.WithField("expectedStatus", http.StatusOK).Debug("JBoss client: authenticated request failed")
		return nil, fmt.Errorf("request failed with status: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Debug("JBoss client: failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	log.WithFields(map[string]interface{}{
		"responseSize": len(body),
		"response":     string(body),
	}).Debug("JBoss client: received response body")

	var wildflyResp WildFlyResponse
	if err := json.Unmarshal(body, &wildflyResp); err != nil {
		log.WithError(err).Debug("JBoss client: failed to unmarshal response")
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	log.WithField("outcome", wildflyResp.Outcome).Debug("JBoss client: successfully parsed response")
	return &wildflyResp, nil
}

// digestAuth holds digest authentication parameters
type digestAuth struct {
	realm  string
	nonce  string
	qop    string
	opaque string
}

// parseDigestAuth parses WWW-Authenticate header
func (c *JBossClient) parseDigestAuth(authHeader string) (*digestAuth, error) {
	if !strings.HasPrefix(authHeader, "Digest ") {
		return nil, fmt.Errorf("not a digest auth header")
	}

	auth := &digestAuth{}

	// Simple regex patterns to extract digest auth parameters
	realmRegex := regexp.MustCompile(`realm="([^"]*)"`)
	nonceRegex := regexp.MustCompile(`nonce="([^"]*)"`)
	qopRegex := regexp.MustCompile(`qop="([^"]*)"`)
	opaqueRegex := regexp.MustCompile(`opaque="([^"]*)"`)

	if matches := realmRegex.FindStringSubmatch(authHeader); len(matches) > 1 {
		auth.realm = matches[1]
	}
	if matches := nonceRegex.FindStringSubmatch(authHeader); len(matches) > 1 {
		auth.nonce = matches[1]
	}
	if matches := qopRegex.FindStringSubmatch(authHeader); len(matches) > 1 {
		auth.qop = matches[1]
	}
	if matches := opaqueRegex.FindStringSubmatch(authHeader); len(matches) > 1 {
		auth.opaque = matches[1]
	}

	if auth.realm == "" || auth.nonce == "" {
		return nil, fmt.Errorf("missing required digest auth parameters")
	}

	return auth, nil
}

// createDigestAuthHeader creates the Authorization header value for digest auth
func (c *JBossClient) createDigestAuthHeader(auth *digestAuth, method, uri string) (string, error) {
	// Generate cnonce
	cnonceBytes := make([]byte, 16)
	_, err := rand.Read(cnonceBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate cnonce: %w", err)
	}
	cnonce := hex.EncodeToString(cnonceBytes)

	nc := "00000001"

	// Calculate HA1 = MD5(username:realm:password)
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", c.username, auth.realm, c.password)))
	ha1Hex := hex.EncodeToString(ha1[:])

	// Calculate HA2 = MD5(method:uri)
	ha2 := md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri)))
	ha2Hex := hex.EncodeToString(ha2[:])

	// Calculate response
	var response string
	if auth.qop == "auth" || auth.qop == "auth-int" {
		// response = MD5(HA1:nonce:nc:cnonce:qop:HA2)
		responseHash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1Hex, auth.nonce, nc, cnonce, auth.qop, ha2Hex)))
		response = hex.EncodeToString(responseHash[:])
	} else {
		// response = MD5(HA1:nonce:HA2)
		responseHash := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", ha1Hex, auth.nonce, ha2Hex)))
		response = hex.EncodeToString(responseHash[:])
	}

	// Build authorization header
	authParts := []string{
		fmt.Sprintf(`username="%s"`, c.username),
		fmt.Sprintf(`realm="%s"`, auth.realm),
		fmt.Sprintf(`nonce="%s"`, auth.nonce),
		fmt.Sprintf(`uri="%s"`, uri),
		fmt.Sprintf(`response="%s"`, response),
	}

	if auth.opaque != "" {
		authParts = append(authParts, fmt.Sprintf(`opaque="%s"`, auth.opaque))
	}

	if auth.qop != "" {
		authParts = append(authParts, fmt.Sprintf(`qop=%s`, auth.qop))
		authParts = append(authParts, fmt.Sprintf(`nc=%s`, nc))
		authParts = append(authParts, fmt.Sprintf(`cnonce="%s"`, cnonce))
	}

	authHeaderValue := "Digest " + strings.Join(authParts, ", ")
	return authHeaderValue, nil
}

// Close closes the HTTP client (if needed)
func (c *JBossClient) Close() error {
	// HTTP client doesn't need explicit closing in Go
	return nil
}

// Helper functions to safely extract values from map[string]interface{}
func getStringValue(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolValue(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getIntValue(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		if i, ok := val.(float64); ok {
			return int(i)
		}
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

// parseContentArray parses the content array from deployment result
func parseContentArray(contentData []interface{}) []DeploymentContent {
	var content []DeploymentContent
	for _, item := range contentData {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if hashData, ok := itemMap["hash"].(map[string]interface{}); ok {
				deploymentContent := DeploymentContent{
					Hash: make(map[string]string),
				}
				for key, val := range hashData {
					if strVal, ok := val.(string); ok {
						deploymentContent.Hash[key] = strVal
					}
				}
				content = append(content, deploymentContent)
			}
		}
	}
	return content
}

// parseUndertowSubsystem parses Undertow subsystem information
func parseUndertowSubsystem(data map[string]interface{}) *UndertowSubsystem {
	if data == nil {
		return nil
	}

	undertow := &UndertowSubsystem{
		ActiveSessions:  getIntValue(data, "active-sessions"),
		ContextRoot:     getStringValue(data, "context-root"),
		Server:          getStringValue(data, "server"),
		SessionsCreated: getIntValue(data, "sessions-created"),
		VirtualHost:     getStringValue(data, "virtual-host"),
		Servlet:         make(map[string]ServletInfo),
	}

	// Parse servlet information
	if servletData, ok := data["servlet"].(map[string]interface{}); ok {
		for servletName, servletInfo := range servletData {
			if servletMap, ok := servletInfo.(map[string]interface{}); ok {
				servlet := ServletInfo{
					MaxRequestTime:   getIntValue(servletMap, "max-request-time"),
					MinRequestTime:   getIntValue(servletMap, "min-request-time"),
					RequestCount:     getIntValue(servletMap, "request-count"),
					ServletClass:     getStringValue(servletMap, "servlet-class"),
					ServletName:      getStringValue(servletMap, "servlet-name"),
					TotalRequestTime: getIntValue(servletMap, "total-request-time"),
				}
				undertow.Servlet[servletName] = servlet
			}
		}
	}

	return undertow
}

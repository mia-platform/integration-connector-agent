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

package consoleclient

import (
	"context"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	m2mTokenPath = "/api/m2m/oauth/token" //nolint:gosec // false positive for G101
)

type TokenManager interface {
	SetAuthHeader(req *http.Request) error
}

type ClientSecretBasic struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	endpoint     string

	cachedTkn *oauth2.Token
	lock      sync.Mutex
}

func NewClientCredentialsTokenManager(baseURL, clientID, clientSecret string) (*ClientSecretBasic, error) {
	endpointURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	endpointURL.Path = m2mTokenPath
	endpoint := endpointURL.String()
	return &ClientSecretBasic{
		BaseURL:      baseURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		endpoint:     endpoint,
	}, nil
}

func (t *ClientSecretBasic) SetAuthHeader(req *http.Request) error {
	if err := t.ensureCachedToken(req.Context()); err != nil {
		return err
	}

	t.cachedTkn.SetAuthHeader(req)
	return nil
}

func (t *ClientSecretBasic) ensureCachedToken(ctx context.Context) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.cachedTkn.Valid() {
		return nil
	}

	config := clientcredentials.Config{
		ClientID:     t.ClientID,
		ClientSecret: t.ClientSecret,
		TokenURL:     t.endpoint,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	tkn, err := config.Token(ctx)
	if err != nil {
		return err
	}

	t.cachedTkn = tkn
	return nil
}

// Copyright (C) 2025 Mia srl
// SPDX-License-Identifier: AGPL-3.0-or-later OR Commercial
// See LICENSE.md for more details

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

package consolecatalog

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	m2mTokenPath = "api/m2m/oauth/token"
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

func newClientCredentialsTokenManager(baseURL, clientID, clientSecret string) *ClientSecretBasic {
	return &ClientSecretBasic{
		BaseURL:      baseURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		endpoint:     fmt.Sprintf("%s%s", baseURL, m2mTokenPath),
	}
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

	if t.cachedTkn != nil && !isExpired(t.cachedTkn.ExpiresIn) {
		return nil
	}

	config := clientcredentials.Config{
		ClientID:     t.ClientID,
		ClientSecret: t.ClientSecret,
		TokenURL:     t.endpoint,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	tkn, err := config.TokenSource(ctx).Token()
	if err != nil {
		return err
	}

	t.cachedTkn = tkn
	return nil
}

func isExpired(expiresIn int64) bool {
	return time.Now().After(time.Now().Add(time.Duration(expiresIn) * time.Second))
}

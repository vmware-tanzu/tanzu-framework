package auth

/*
Inspired from https://gitlab.eng.vmware.com/olympus/api/blob/master/pkg/common/auth/oidc.go
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/pkg/errors"
)

const (
	// AuthTokenDir is a directory where cluster access token and refresh tokens are stored.
	AuthTokenDir = "tokens"

	// ExtraIDToken is the key in the Extra fields map that contains id_token.
	ExtraIDToken = "id_token"

	// StgIssuer is the CSP staging issuer.
	StgIssuer = "https://console-stg.cloud.vmware.com/csp/gateway/am/api"

	// ProdIssuer is the CSP issuer.
	ProdIssuer = "https://console.cloud.vmware.com/csp/gateway/am/api"

	// APITokenKey is the env var for an API token override.
	APITokenKey = "TMC_API_TOKEN"

	defaultLoginTimeout = 5 * time.Minute
)

var (
	// LoginSuccessPage is the html page displayed by the browser on successful login
	LoginSuccessPage = strings.TrimSpace(`
<p><strong>tmc login flow complete</strong></p>
<p>You are authenticated and can close this page.</p>
`)

	// KnownIssuers are known OAuth2 endpoints in each CSP environment.
	KnownIssuers = map[string]oauth2.Endpoint{
		StgIssuer: {
			AuthURL:   "https://console-stg.cloud.vmware.com/csp/gateway/discovery",
			TokenURL:  "https://console-stg.cloud.vmware.com/csp/gateway/am/api/auth/authorize",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
		ProdIssuer: {
			AuthURL:   "https://console.cloud.vmware.com/csp/gateway/discovery",
			TokenURL:  "https://console.cloud.vmware.com/csp/gateway/am/api/auth/authorize",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
)

// IDTokenFromTokenSource parses out the id token from extra info in tokensource if available, or returns empty string.
func IDTokenFromTokenSource(token oauth2.Token) (idTok string) {
	extraTok := token.Extra("id_token")
	if extraTok != nil {
		idTok = extraTok.(string)
	}
	return
}

// Token is a CSP token.
type Token struct {
	// IDToken for OIDC.
	IDToken string `json:"id_token"`

	// TokenType is the type of token.
	TokenType string `json:"token_type"`

	// ExpiresIn is experation in seconds.
	ExpiresIn int64 `json:"expires_in"`

	// Scope of the token.
	Scope string `json:"scope"`

	// AccessToken from CSP.
	AccessToken string `json:"access_token"`

	// RefreshToken for use with Refresh Token grant.
	RefreshToken string `json:"refresh_token"`
}

// GetAccessTokenFromAPIToken fetches CSP access token using the API-token.
func GetAccessTokenFromAPIToken(apiToken string, issuer string) (*Token, error) {
	api := fmt.Sprintf("%s/auth/api-tokens/authorize", issuer)
	data := url.Values{}
	data.Set("refresh_token", apiToken)
	req, _ := http.NewRequest("POST", api, bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to obtain access token. Please provide valid VMware Cloud Services API-token")
	}
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("Http status: %s", resp.Status)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	token := Token{}

	if err = json.Unmarshal(body, &token); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal auth token")
	}

	return &token, nil
}

// GetIssuer returns the appropriate CSP issuer based on the environment.
func GetIssuer(staging bool) string {
	if staging {
		return StgIssuer
	}
	return ProdIssuer
}

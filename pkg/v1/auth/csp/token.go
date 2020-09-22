package csp

/*
Inspired from https://gitlab.eng.vmware.com/olympus/api/blob/master/pkg/common/auth/oidc.go
*/

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	APITokenKey = "CSP_API_TOKEN"

	defaultLoginTimeout = 5 * time.Minute
)

var (
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

// Claims are the jwt claims.
// TODO (pbarker): make this proper.
type Claims struct {
	Username    string
	Permissions []string
	OrgID       string
	Raw         map[string]interface{}
}

// ParseToken the JWT payload and return the decoded information.
// TODO (pbarker): need to get this from another place
func ParseToken(tkn *oauth2.Token) (*Claims, error) {
	accessToken := strings.Split(tkn.AccessToken, ".")
	if len(accessToken) < 3 {
		panic("invalid accessToken")
	}
	jwtPayload := accessToken[1] // Get just the payload part of the JWT

	if l := len(jwtPayload) % 4; l > 0 {
		jwtPayload += strings.Repeat("=", 4-l)
	}

	payload, err := base64.URLEncoding.DecodeString(jwtPayload)
	if err != nil {
		return nil, err
	}

	c := make(map[string]interface{})
	err = json.Unmarshal(payload, &c)
	if err != nil {
		return nil, err
	}

	perm := []string{}
	p, ok := c["perms"].([]interface{})
	if !ok {
		log.Warning("could not cast permissions")
	}
	for _, i := range p {
		perm = append(perm, i.(string))
	}
	uname, ok := c["username"].(string)
	if !ok {
		return nil, fmt.Errorf("could not parse username from token")
	}
	orgID, ok := c["context_name"].(string)
	if !ok {
		return nil, fmt.Errorf("could not parse orgID from token")
	}
	claims := &Claims{
		Username:    uname,
		Permissions: perm,
		OrgID:       orgID,
		Raw:         c,
	}

	return claims, nil
}

// GetToken fetches a token for the current auth context.
func GetToken(g *clientv1alpha1.GlobalServerAuth) (*oauth2.Token, error) {
	if !IsExpired(g.Expiration.Time) {
		tok := &oauth2.Token{
			AccessToken: g.AccessToken,
			Expiry:      g.Expiration.Time,
		}
		return tok.WithExtra(map[string]interface{}{
			"id_token": g.IDToken,
		}), nil
	}

	// TODO (pbarker): support more issuers.
	token, err := GetAccessTokenFromAPIToken(g.RefreshToken, ProdIssuer)
	if err != nil {
		return nil, err
	}

	g.Type = "api-token"
	expiration := time.Now().Local().Add(time.Second * time.Duration(token.ExpiresIn))
	g.Expiration = metav1.NewTime(expiration)
	g.RefreshToken = token.RefreshToken
	g.AccessToken = token.AccessToken
	g.IDToken = token.IDToken

	tok := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       expiration,
	}
	return tok.WithExtra(map[string]interface{}{
		"id_token": token.IDToken,
	}), nil
}

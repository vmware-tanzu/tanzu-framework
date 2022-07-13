// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package csp

import (
	"context"
	"fmt"
	"time"

	"github.com/aunum/log"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	grpc_oauth "google.golang.org/grpc/credentials/oauth"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
)

const (
	mdKeyAuthToken   = "Authorization"
	authTokenPrefix  = "Bearer "
	mdKeyAuthIDToken = "X-User-Id"
	apiToken         = "api-token"
)

// WithCredentialDiscovery returns a grpc.CallOption that adds credentials into gRPC calls.
// The credentials are loaded from the auth context found on the machine.
func WithCredentialDiscovery() (grpc.CallOption, error) {
	cfg, err := config.GetClientConfig()
	if err != nil {
		return nil, err
	}
	// Wrap our TokenSource to supply id tokens
	return grpc.PerRPCCredentials(&TokenSource{
		TokenSource: &configSource{cfg},
	}), nil
}

// WithStaticCreds will wrap a static access token into a grpc.CallOption
func WithStaticCreds(accessToken string) grpc.CallOption {
	return grpc.PerRPCCredentials(&grpc_oauth.TokenSource{
		TokenSource: oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: accessToken},
		),
	})
}

type configSource struct {
	*configv1alpha1.ClientConfig
}

// Token fetches the token.
func (c *configSource) Token() (*oauth2.Token, error) {
	g, err := c.GetCurrentServer()
	if err != nil {
		return nil, err
	}
	if !g.IsGlobal() {
		return nil, fmt.Errorf("trying to fetch token for non global server")
	}
	if !IsExpired(g.GlobalOpts.Auth.Expiration.Time) {
		tok := &oauth2.Token{
			AccessToken: g.GlobalOpts.Auth.AccessToken,
			Expiry:      g.GlobalOpts.Auth.Expiration.Time,
		}
		return tok.WithExtra(map[string]interface{}{
			ExtraIDToken: g.GlobalOpts.Auth.IDToken,
		}), nil
	}
	token, err := GetAccessTokenFromAPIToken(g.GlobalOpts.Auth.RefreshToken, ProdIssuer)
	if err != nil {
		return nil, err
	}

	g.GlobalOpts.Auth.Type = apiToken
	expiration := time.Now().Local().Add(time.Second * time.Duration(token.ExpiresIn))
	g.GlobalOpts.Auth.Expiration = metav1.NewTime(expiration)
	g.GlobalOpts.Auth.RefreshToken = token.RefreshToken
	g.GlobalOpts.Auth.AccessToken = token.AccessToken
	g.GlobalOpts.Auth.IDToken = token.IDToken

	// Acquire tanzu config lock
	config.AcquireTanzuConfigLock()
	defer config.ReleaseTanzuConfigLock()

	// TODO: Add Read/Write locking mechanism before updating the configuration
	// Currently we are only acquiring the lock while updating the configuration
	if err := config.StoreClientConfig(c.ClientConfig); err != nil {
		return nil, err
	}

	tok := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       expiration,
	}
	return tok.WithExtra(map[string]interface{}{
		ExtraIDToken: token.IDToken,
	}), nil
}

// TokenSource supplies PerRPCCredentials from an oauth2.TokenSource using CSP as the IDP.
// It will supply access token through authorization header and id_token through user-Id header
type TokenSource struct {
	oauth2.TokenSource
}

// GetRequestMetadata gets the request metadata as a map from a TokenSource.
func (ts TokenSource) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := ts.Token()
	if err != nil {
		return nil, err
	}

	headers := map[string]string{mdKeyAuthToken: authTokenPrefix + " " + token.AccessToken}
	idTok := IDTokenFromTokenSource(token)
	if idTok != "" {
		headers[mdKeyAuthIDToken] = idTok
	}

	return headers, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security.
func (ts TokenSource) RequireTransportSecurity() bool {
	return true
}

// GetAuthOptsOrExit returns the grpc auth options. If accessToken is not empty it uses it, else it fetches the token
// from the current auth context. If it encounters and error, it exits.
func GetAuthOptsOrExit() grpc.CallOption {
	var authOpts grpc.CallOption
	var err error
	authOpts, err = WithCredentialDiscovery()
	if err != nil {
		log.Fatal("Not logged in. Please retry after logging in")
	}

	return authOpts
}

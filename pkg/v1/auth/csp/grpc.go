// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package csp

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/aunum/log"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpc_oauth "google.golang.org/grpc/credentials/oauth"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
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
	cfg, err := client.GetConfig()
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
	*clientv1alpha1.Config
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

	if err := client.StoreConfig(c.Config); err != nil {
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

// ConnectToEndpointOrExit returns a client connection to the provided endpoint. If it encounters an error, it exits.
func ConnectToEndpointOrExit(ctxopts ...ContextOpts) *grpc.ClientConn {
	conn, err := ConnectToEndpoint(ctxopts...)
	if err != nil {
		os.Exit(1)
	}
	return conn
}

// ContextOpts for the context.
type ContextOpts func(context.Context) context.Context

// ConnectToEndpoint attempts to connect to the provided endpoint. If endpoint is empty, it picks up the endpoint
// from the current auth ctx.
func ConnectToEndpoint(ctxopts ...ContextOpts) (*grpc.ClientConn, error) {
	cfg, err := client.GetConfig()
	if err != nil {
		log.Errorf("Could not get current auth context with error: %v", err)
		return nil, err
	}
	s, err := cfg.GetCurrentServer()
	if err != nil {
		return nil, err
	}
	endpoint := s.GlobalOpts.Endpoint
	unaryInterceptors := []grpc.UnaryClientInterceptor{
		unaryClientInterceptor(ctxopts...),
	}

	streamInterceptors := []grpc.StreamClientInterceptor{
		streamClientInterceptor(ctxopts...),
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(unaryInterceptors...)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(streamInterceptors...)),
		grpc.WithBlock(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(DefaultTimeout)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, endpoint, dialOpts...)
	if err != nil {
		log.Errorf("Could not reach backend: %+v with error: %+v\n", endpoint, err)
		return nil, err
	}

	return conn, nil
}

// unaryClientInterceptor adds the client information metadata to the outgoing unary gRPC request context
func unaryClientInterceptor(ctxopts ...ContextOpts) grpc.UnaryClientInterceptor {
	return func(reqCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		timeoutCtx, cancel := context.WithTimeout(reqCtx, time.Duration(DefaultTimeout)*time.Second)
		defer cancel()

		for _, opt := range ctxopts {
			timeoutCtx = opt(timeoutCtx)
		}
		return invoker(timeoutCtx, method, req, reply, cc, opts...)
	}
}

// streamClientInterceptor adds the client information metadata to the outgoing streaming gRPC request context
func streamClientInterceptor(ctxopts ...ContextOpts) grpc.StreamClientInterceptor {
	return func(reqCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		timeoutCtx, cancel := context.WithTimeout(reqCtx, time.Duration(DefaultTimeout)*time.Second)
		defer cancel()

		for _, opt := range ctxopts {
			timeoutCtx = opt(timeoutCtx)
		}
		return streamer(timeoutCtx, desc, cc, method, opts...)
	}
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

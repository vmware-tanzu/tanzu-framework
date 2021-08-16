// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"crypto/tls"
	"os"
	"time"

	"github.com/aunum/log"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
)

const (
	// DialTimeout is the default gRPC dial timeout in seconds.
	DialTimeout = 30

	// PingTime is the default gRPC keep-alive ping time internal in seconds. This corresponds to the minimum ping wait
	// time on the server side. Ref.: https://pkg.go.dev/google.golang.org/grpc/keepalive#EnforcementPolicy
	PingTime = 300

	// PingTimeout is the default gRPC keep-alive ping timeout in seconds.
	PingTimeout = 30

	// UnaryTimeout is the default unary RPC timeout in seconds.
	UnaryTimeout = 30
)

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
	cfg, err := config.GetClientConfig()
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
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
		})),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(unaryInterceptors...)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(streamInterceptors...)),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    PingTime * time.Second,
			Timeout: PingTimeout * time.Second,
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(DialTimeout)*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, endpoint, dialOpts...)
	if err != nil {
		log.Errorf("Could not reach backend: %+v with error: %+v\n", endpoint, err)
		return nil, err
	}

	return conn, nil
}

// unaryClientInterceptor adds a default timeout of 30 seconds and the provided context options to the outgoing unary gRPC request context
func unaryClientInterceptor(ctxopts ...ContextOpts) grpc.UnaryClientInterceptor {
	return func(reqCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		timeoutCtx, cancel := context.WithTimeout(reqCtx, time.Duration(UnaryTimeout)*time.Second)
		defer cancel()

		for _, opt := range ctxopts {
			timeoutCtx = opt(timeoutCtx)
		}
		return invoker(timeoutCtx, method, req, reply, cc, opts...)
	}
}

// streamClientInterceptor adds the provided context options to the outgoing streaming gRPC request context
func streamClientInterceptor(ctxopts ...ContextOpts) grpc.StreamClientInterceptor {
	return func(reqCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		for _, opt := range ctxopts {
			reqCtx = opt(reqCtx)
		}
		return streamer(reqCtx, desc, cc, method, opts...)
	}
}

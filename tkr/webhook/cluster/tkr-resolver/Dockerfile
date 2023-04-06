# syntax=docker/dockerfile:1.4
# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
# Enable Buildkit

# Build from publicly reachable source by default, but allow people to re-build images on
# top of their own trusted images.
ARG BUILDER_BASE_IMAGE=golang:1.18
ARG DISTROLESS_BASE_IMAGE=gcr.io/distroless/static:nonroot

# Build the manager binary
FROM $BUILDER_BASE_IMAGE as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY apis/run/ apis/run/
COPY util/ util/

COPY tkr/go.mod tkr/go.mod
COPY tkr/go.sum tkr/go.sum

WORKDIR /workspace/tkr
# Setting default GOPROXY to https://proxy.golang.org,direct and GOSUMDB to sum.golang.org which can be override by Makefile
ARG GOSUMDB
ARG GOPROXY
ENV GOSUMDB=$GOSUMDB
ENV GOPROXY=$GOPROXY
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/.local/share/golang --mount=type=cache,target=/go/pkg/mod go mod download

# Copy the go source
COPY tkr/ ./

# Build
ARG LD_FLAGS
ENV LD_FLAGS="$LD_FLAGS "'-extldflags "-static"'
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/.local/share/golang --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -ldflags "$LD_FLAGS" -o manager webhook/cluster/tkr-resolver/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM $DISTROLESS_BASE_IMAGE
WORKDIR /
COPY --from=builder /workspace/tkr/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]

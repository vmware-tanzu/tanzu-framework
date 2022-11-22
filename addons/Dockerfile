# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Enable Buildkit
# syntax=docker/dockerfile:1.4

# Build from publicly reachable source by default, but allow people to re-build images on
# top of their own trusted images.
ARG BUILDER_BASE_IMAGE=golang:1.18
ARG CACHE_GOBUILD=type=cache,target=/root/.cache/go-build
ARG CACHE_GOLANG=cache,target=/root/.local/share/golang
ARG CACHE_GOMOD=type=cache,target=/go/pkg/mod

# Build the manager binary
FROM $BUILDER_BASE_IMAGE as builder
WORKDIR /workspace
# Copy the go source
COPY ./ ./


# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN  --mount=$CACHE_GOLANG --mount=$CACHE_GOMOD go mod download

# Build
ARG LD_FLAGS
ENV LD_FLAGS="$LD_FLAGS "'-extldflags "-static"'
RUN $CACHE cd addons && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -ldflags "$LD_FLAGS" -o bin/manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/addons/bin/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]

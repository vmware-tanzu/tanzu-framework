# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Enable Buildkit
# syntax=docker/dockerfile:1.4

ARG CACHE_GOBUILD=type=cache,target=/root/.cache/go-build
ARG CACHE_GOLANG=cache,target=/root/.local/share/golang
ARG CACHE_GOMOD=type=cache,target=/go/pkg/mod
ARG CACHE="--mount=${CACHE_GOLANG} --mount=${CACHE_GOBUILD} --mount=${CACHE_GOMOD}"

# Build from publicly reachable source by default, but allow people to re-build images on
# top of their own trusted images.
ARG BUILDER_BASE_IMAGE=golang:1.18

# Build the manager binary
FROM $BUILDER_BASE_IMAGE as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY apis/run/ apis/run/
COPY util/ util/

COPY object-propagation/go.mod object-propagation/go.mod
COPY object-propagation/go.sum object-propagation/go.sum

WORKDIR /workspace/object-propagation
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/.local/share/golang --mount=type=cache,target=/go/pkg/mod go mod download
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer

# Copy the go source
COPY object-propagation/ ./

# Build
ARG LD_FLAGS
ENV LD_FLAGS="$LD_FLAGS "'-extldflags "-static"'
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/.local/share/golang --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -ldflags "$LD_FLAGS" -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/object-propagation/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]
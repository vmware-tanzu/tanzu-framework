# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Build from publicly reachable source by default, but allow people to re-build images on
# top of their own trusted images.
ARG BUILDER_BASE_IMAGE=mirror.gcr.io/library/golang:1.16

# Build the post-deploy binary
FROM $BUILDER_BASE_IMAGE as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the source
COPY main.go main.go
COPY pkg/ pkg/
COPY Makefile Makefile
#COPY .git/ .git/

RUN make native

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/tkg-pinniped-post-deploy .
CMD ["/tkg-pinniped-post-deploy"]

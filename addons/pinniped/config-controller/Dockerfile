# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Build from publicly reachable source by default, but allow people to re-build images on
# top of their own trusted images.
ARG BUILDER_BASE_IMAGE=golang:1.17

# Build the config-controller-manager binary
FROM $BUILDER_BASE_IMAGE as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the source
COPY main.go main.go
COPY controllers/controllers.go controllers/controllers.go
COPY controllers/handlers.go controllers/handlers.go
COPY utils/utils.go utils/utils.go
COPY constants/constants.go constants/constants.go

# Build
ARG LD_FLAGS
ENV LD_FLAGS="$LD_FLAGS "'-extldflags "-static"'
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -ldflags "$LD_FLAGS" -o pinniped-config-controller-manager .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/pinniped-config-controller-manager .
USER nonroot:nonroot
ENTRYPOINT ["/pinniped-config-controller-manager"]

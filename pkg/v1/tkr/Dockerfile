# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Build from publicly reachable source by default, but allow people to re-build images on
# top of their own trusted images.
ARG BUILDER_BASE_IMAGE=mirror.gcr.io/library/golang:1.16

# Build the manager binary
FROM $BUILDER_BASE_IMAGE as builder

WORKDIR /workspace


ENV GOPRIVATE=github.com/vmware-tanzu/tanzu-framework

# Create the image's SSH directory.
RUN mkdir -p /root/.ssh && chmod 0700 /root/.ssh

COPY $SSH_PRIVATE_KEY  /root/.ssh/id_rsa
RUN chmod 0400 /root/.ssh/id_rsa


# Ensure git knows how to access github.
RUN git config --global "url.git@github.com:.insteadof" "https://github.com/"

# Seed Githab's host keys to prevent SSH from hanging.
RUN ssh-keyscan github.com >>/root/.ssh/known_hosts

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN  go mod download
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer


# Copy the go source
COPY apis/ apis/
COPY pkg/v1/tkr/controllers/ pkg/v1/tkr/controllers/
COPY pkg/v1/tkr/main.go pkg/v1/tkr/main.go
COPY pkg/v1/tkr/pkg/ pkg/v1/tkr/pkg/


# Build
ARG LD_FLAGS
ENV LD_FLAGS="$LD_FLAGS "'-extldflags "-static"'
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -ldflags "$LD_FLAGS" -o manager pkg/v1/tkr/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]

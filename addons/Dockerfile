# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Build the manager binary
FROM harbor-repo.vmware.com/dockerhub-proxy-cache/library/golang:1.16 as builder

WORKDIR /workspace

ENV GOPRIVATE=github.com/vmware-tanzu/*
ENV GONOSUMDB=github.com/vmware-tanzu/*

ARG SSH_PRIVATE_KEY

# Create the image's SSH directory.
RUN mkdir -p /root/.ssh && chmod 0700 /root/.ssh
COPY $SSH_PRIVATE_KEY  /root/.ssh/id_rsa
RUN chmod 0400 /root/.ssh/id_rsa

# Ensure git knows how to access github.
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

# Seed Github's host keys to prevent SSH from hanging.
RUN ssh-keyscan github.com >>/root/.ssh/known_hosts

# Copy the go source
COPY ./ ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Build
RUN cd addons && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -ldflags '-extldflags "-static"' -o bin/manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/addons/bin/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]

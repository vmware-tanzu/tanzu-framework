# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

FROM harbor-repo.vmware.com/dockerhub-proxy-cache/library/golang:1.14.4 as builder

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

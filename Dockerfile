# Copyright 2023 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# This Dockerfile is currently consumed by build tooling https://github.com/vmware-tanzu/build-tooling-for-integrations
# to build components in tanzu-framework, check out build-tooling.mk to understand how this is being consumed.

ARG BUILDER_BASE_IMAGE=golang:1.18

FROM --platform=${BUILDPLATFORM} $BUILDER_BASE_IMAGE as base
ARG COMPONENT
ARG GOPROXY_ARG
ENV GOPROXY=${GOPROXY_ARG}
WORKDIR /workspace
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    cd $COMPONENT && go mod download

# Linting
FROM harbor-repo.vmware.com/dockerhub-proxy-cache/golangci/golangci-lint:v1.50 AS lint-base
FROM base AS lint
RUN --mount=target=. \
    --mount=from=lint-base,src=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/golangci-lint \
    cd $COMPONENT && golangci-lint run --config /workspace/.golangci.yaml --timeout 10m0s ./...

FROM base AS fmt
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd $COMPONENT && go fmt ./...

FROM base AS vet
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd $COMPONENT && go vet ./...

# Testing
FROM base AS test
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd $COMPONENT && mkdir /out && go test -v -coverprofile=/out/cover.out ./...

# Build the manager binary
FROM base as builder
ARG TARGETOS
ARG TARGETARCH
ARG LD_FLAGS
ENV LD_FLAGS="$LD_FLAGS "'-extldflags "-static"'
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    cd $COMPONENT && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GO111MODULE=on go build -o /out/manager ./main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot as image
WORKDIR /
COPY --from=builder /out/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]

FROM scratch AS unit-test-coverage
COPY --from=test /out/cover.out /cover.out

FROM scratch AS bin-unix
COPY --from=builder /out/manager /

FROM bin-unix AS bin-linux
FROM bin-unix AS bin-darwin

FROM scratch AS bin-windows
COPY --from=builder /out/manager /manager.exe

FROM bin-${TARGETOS} as bin

# syntax=docker/dockerfile:1.4
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
COPY featuregates/controller/go.mod featuregates/controller/go.mod
COPY featuregates/controller/go.sum featuregates/controller/go.sum
COPY featuregates/client featuregates/client
COPY apis/ apis/
COPY cli/runtime cli/runtime
COPY capabilities/client capabilities/client
COPY util/go.mod util/go.mod

WORKDIR featuregates/controller
# Setting default GOPROXY to https://proxy.golang.org,direct and GOSUMDB to sum.golang.org which can be override by Makefile
ARG GOSUMDB
ARG GOPROXY
ENV GOSUMDB=$GOSUMDB
ENV GOPROXY=$GOPROXY
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/.local/share/golang --mount=type=cache,target=/go/pkg/mod go mod download

# Copy the go source
COPY ./featuregates/controller/pkg ./pkg
COPY ./featuregates/controller/main.go ./main.go

# Build
ARG LD_FLAGS
ENV LD_FLAGS="$LD_FLAGS "'-extldflags "-static"'

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/root/.local/share/golang --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "$LD_FLAGS" -o manager ./main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM $DISTROLESS_BASE_IMAGE
WORKDIR /
COPY --from=builder /workspace/featuregates/controller/manager .
USER nonroot:nonroot

ENTRYPOINT ["/manager"]

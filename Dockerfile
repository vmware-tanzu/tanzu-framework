# Copyright 2023 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

ARG BUILDER_BASE_IMAGE=golang:1.19

FROM --platform=${BUILDPLATFORM} $BUILDER_BASE_IMAGE as base
ARG COMPONENT
ARG GOPROXY_ARG
ENV GOPROXY=${GOPROXY_ARG}
WORKDIR /workspace
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    cd $COMPONENT && go mod download

# Linting
FROM golangci/golangci-lint:latest AS lint-base
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

# Download and install Carvel's imgpkg program.
FROM --platform=${BUILDPLATFORM} $BUILDER_BASE_IMAGE as carvel-base
ARG IMGPKG_VERSION
ARG BUILDARCH
RUN wget -O /bin/imgpkg https://github.com/vmware-tanzu/carvel-imgpkg/releases/download/${IMGPKG_VERSION}/imgpkg-linux-${BUILDARCH} && \
    chmod +x /bin/imgpkg

# Download, extract, and install Tanzu CLI Plugin Builder
# Note: Until the first version of Tanzu CLI is released, we will be cloning
# the vmware-tanzu/tanzu-cli repo and building the plugin builder here.
# When the first release is available, we'll get the plugin builder tool
# from their central repository.
FROM base AS cli-plugin-builder-install
RUN cd /tmp && \
    git clone https://github.com/vmware-tanzu/tanzu-cli.git && \
    cd tanzu-cli/cmd/plugin/builder && \
    go build -o /bin/tanzu-builder . && \
    cd /workspace

# Run Tanzu plugin builder and compile all plugins in the project's cmd/cli/plugin directory.
FROM cli-plugin-builder-install AS cli-plugin-build-prep
ARG CLI_PLUGIN_VERSION
ARG CLI_PLUGIN
ARG OCI_REGISTRY
RUN --mount=type=bind,readwrite \
    --mount=from=carvel-base,src=/bin/imgpkg,target=/bin/imgpkg \
    tanzu-builder plugin build --match "${CLI_PLUGIN}" \
        --version "${CLI_PLUGIN_VERSION}" \
        --binary-artifacts "./artifacts/plugins" && \
    tanzu-builder plugin build-package \
        --oci-registry "${OCI_REGISTRY}" && \
    mkdir -p /out/plugin-artifacts && \
    cp -r artifacts /out/plugin-artifacts

# Run Tanzu plugin builder and publish all plugins in the project's cmd/plugin directory.
FROM cli-plugin-builder-install AS cli-plugin-publish 
ARG REPOSITORY
ARG PUBLISHER
ARG VENDOR
ARG IMGPKG_USERNAME
ARG IMGPKG_PASSWORD
ENV IMGPKG_USERNAME=${IMGPKG_USERNAME} IMGPKG_PASSWORD=${IMGPKG_PASSWORD}
RUN --mount=type=bind,readwrite \
    --mount=from=carvel-base,src=/bin/imgpkg,target=/bin/imgpkg \
    tanzu-builder plugin publish-package \
        --repository "${REPOSITORY}" \
        --publisher "${PUBLISHER}" \
        --vendor "${VENDOR}" \
        --package-artifacts "./build/artifacts/packages"

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

FROM scratch as cli-plugin-build
COPY --from=cli-plugin-build-prep /out/plugin-artifacts/ .

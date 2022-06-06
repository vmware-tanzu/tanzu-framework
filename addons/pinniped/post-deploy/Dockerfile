# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Build from publicly reachable source by default, but allow people to re-build images on
# top of their own trusted images.
ARG BUILDER_BASE_IMAGE=golang:1.17

# Build the post-deploy binary
FROM $BUILDER_BASE_IMAGE as builder

WORKDIR /workspace
# Copy the Go Modules manifests
# We depend on tanzu-auth-controller-manager, so copy it in
COPY tanzu-auth-controller-manager tanzu-auth-controller-manager
COPY post-deploy/go.mod post-deploy/go.mod
COPY post-deploy/go.sum post-deploy/go.sum
RUN cd post-deploy && go mod download

# Copy the source
COPY post-deploy/cmd/ post-deploy/cmd/
COPY post-deploy/pkg/ post-deploy/pkg/
COPY post-deploy/Makefile post-deploy/Makefile
#COPY .git/ .git/

RUN make native -C post-deploy

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/post-deploy/tkg-pinniped-post-deploy-job .
COPY --from=builder /workspace/post-deploy/tkg-pinniped-post-deploy-controller .
CMD ["/tkg-pinniped-post-deploy-job"]

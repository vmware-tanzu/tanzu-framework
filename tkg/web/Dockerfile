# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Build from publicly reachable source by default, but allow people to re-build images on
# top of their own trusted images.
ARG BUILDER_BASE_IMAGE=node:16

# Build the manager binary
FROM $BUILDER_BASE_IMAGE as builder

WORKDIR /workspace

COPY api/spec.yaml /workspace/api/spec.yaml
COPY angular.json package.json package-lock.json tsconfig.app.json tsconfig.json tsconfig.spec.json tslint.json /workspace/
COPY src /workspace/src/

RUN npm ci --legacy-peer-deps
RUN npm run build:prod

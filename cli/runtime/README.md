# Tanzu CLI Integration Library

[![Go Reference](https://pkg.go.dev/badge/github.com/vmware-tanzu/tanzu-framework.svg)](https://pkg.go.dev/github.com/vmware-tanzu/tanzu-framework/cli/runtime)

## Overview

The Tanzu CLI is based on a plugin architecture. This architecture enables teams to build, own, and release their own piece of functionality as well as enable external partners to integrate with the system. The Tanzu CLI Integration Library provides functionality and helper methods to develop Tanzu CLI plugins.

Developers can use the `Builder` admin plugin to bootstrap a new plugin which can then use tooling and functionality available within the integration library to implement its own features.

## Documentation

The [documentation](docs) provides a getting-started guide and details on how to consume the integration library for plugin development.

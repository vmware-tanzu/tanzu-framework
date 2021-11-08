# Context-aware API-driven Plugin Discovery

## Abstract

The Tanzu CLI is an amalgamation of all the Tanzu infrastructure elements under one unified core CLI experience. The core CLI supports a plugin model where the developers of different Tanzu services (bundled or SaaS) can distribute plugins that target functionalities of the services they own. When users switch between different services via the CLI context, we want to surface only the relevant plugins for the given context for a crisp user experience.

## Key Concepts

- CLI - The Tanzu command line interface, built to be pluggable.
- Service - Any tanzu service, user-managed or SaaS. E.g., TKG, TCE, TMC, etc
- Server - An instance of service. E.g., A single TKG management-cluster, a specific TMC endpoint, etc.
- Context - an isolated scope of relevant client-side configurations for a combination of user identity and server identity. There can be multiple contexts for the same combination of {user, server}. This is currently referred to as `Server` in the Tanzu CLI, which can also mean an instance of a service. Hence, we shall use `Context` to avoid confusion.
- Plugin - A scoped piece of functionality that can be used to extend the CLI. Usually refers to a single binary that is invoked by the root Tanzu CLI.
- Scope - the context association level of a plugin
- Stand-alone - independent of the CLI context
- Context-scoped - scoped to one or more contexts
- Discovery - the interface to fetch the list of available plugins and their supported versions
- Distribution - the interface to deliver a plugin for user download
- Scheme - the specific mechanism to discover or download a plugin
- Discovery Scheme - e.g., REST API, CLIPlugin kubernetes API, manifest YAML
- Distribution Scheme - e.g., OCI image, Google Cloud Storage, S3
- Discovery Source - the source server of a plugin metadata for discovery, e.g., a REST API server, a management cluster, a local manifest file, OCI compliant image containing manifest file
- Distribution Repository - the repository of plugin binary for distribution, e.g., an OCI compliant image registry, Google Cloud Storage, an S3 compatible object storage server
- Plugin Descriptor - the metadata about a single plugin version that is installed locally and is used by the core to construct a sub-command under `tanzu`

## Background

## Goals

## Non Goals

## High-Level Design

## Detailed Design

## Alternatives Considered

## Security Considerations

## Compatibility

## Implementation

## Open Issues

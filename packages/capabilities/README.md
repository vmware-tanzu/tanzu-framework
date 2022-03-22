# Capabilities Package

Capabilities package provides Capabilities API that offers the ability to query a cluster's capabilities.
A "capability" is defined as anything a Kubernetes cluster can do or have, such as objects, and the API surface area.
Capability discovery can be used to answer questions such as `Is this a TKG cluster?`, `Does this cluster have a
resource Foo?` etc.

## Components

* capabilities-controller

## Usage Example

To learn more about the Capabilities API and controller and how to use it, refer to this
[doc](../../../docs/api-machinery/capability-discovery.md)

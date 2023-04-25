# Use cases

This document provides some use cases for Framework to understand the project's
context.

## Use Case #1 - Features and FeatureGates

### Context

Allow developers to deliver new functionality to users rapidly but safely.
If there is no way to do that, it would be difficult for individual teams to
develop features in parallel without coordinating with the larger team or
resorting to an expensive branching model. This reduces the velocity, inhibits
experimentation, and also makes the development difficult.

### Solution

Framework provides powerful Features and FeatureGates APIs that allows the
teams to have a system to control rollout and availability of new Features in
controller and plugin logic. With these APIs, it would be easy to modify the
behavior of the plugin or controller without changing the code for more
controlled experimentation, these can be incredibly useful for agile management
style environments.

Check out this [document](runtime-core/features-and-featuregates.md) to
learn more about the Features and FeatureGates APIs and how to use them.

## Use Case #2 - Capabilities

### Context

Determining how to interact with clusters running versioned pieces of software
on various infrastructure providers is complex. The lack of a standard means to
discover details about cluster resource composition and API surface area could
result in manual poking and prodding by teams and this approach could
eventually manifest as undesirable patterns.

### Solution

Framework provides Capability discovery [Go package](https://github.com/vmware-tanzu/tanzu-framework/tree/main/capabilities/client/pkg/discovery)
and [Capability API](https://github.com/vmware-tanzu/tanzu-framework/blob/main/apis/run/v1alpha1/capability_types.go)
to query a cluster's capabilities. It can be used to understand the API surface
area and query for objects in the cluster.

For more detailed information on Capability functionality offered by Framework
check out this [document](runtime-core/capability-discovery.md).

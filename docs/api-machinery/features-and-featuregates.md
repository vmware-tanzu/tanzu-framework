# Framework Features

Framework offers Features and Featuregates APIs to allow developers to have a
system to control rollout and availability of new Features in controller logic.

## Key Concepts

* **Feature**: Functionality that is being introduced to a codebase which will
  be gated by FeatureGate. Features can be activated/deactivated based on the
  stability level policy. Features are created and managed by the API developers
  creating the functionality.
* **FeatureGate**: The mechanism that provides a means to control the state for
  a set of Features. FeatureGates are a resource that is managed by Platform
  Operators.
* **Feature Activation**: The reconciled state of a Feature. If a FeatureGate
  exists that references a specific Feature, it is reconciled by the Feature controller
  to set the correct state of the Feature in its status. If a FeatureGate doesn't exist
  for a Feature, Feature will always default to their defined activation state in the
  stability policy.
* **TKG SDK**: A collection of packages and tools provided by Framework intended
  to provide developers with additional integration capabilities.

## Features API

The Features API provides controller developers a means to roll out new code in a more
predictable manner, using a standard durable API.

Features are a Cluster-level resources and are set to default activation state when
they are created.

The [Features API Spec](apis/core/v1alpha2/feature_types.go) provides following fields:

* **description**: To provide description about the feature.
* **stability**: To indicate the stability level of the feature. Valid stability levels are
  Work In Progress, Experimental, Technical Preview, Stable and Deprecated. Each stability
  level has a policy associated with it and the Feature should adhere to that policy. Learn
  more about the stability level policies [here](##stability-level-policies).

The status of the Feature resource has the observed state of the feature.

### Example

In this example, we will define a `Technical Preview` feature, which is deactivated by default
and is discoverable and mutable. Also, it doesn't void the warranty support if activated for
the environment.

```yaml
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: Feature
metadata:
  name: big-cache
spec:
  description: "A sample big cache Feature"
  stability: "Technical Preview"
```

## FeatureGates API

The [FeatureGates API](apis/core/v1alpha2/featuregate_types.go) provides a
means for operators to set Feature state in an eventually consistent,
level-triggered manner. FeatureGates are a Cluster-level resources.

FeatureGates can be set at cluster creation time or anytime afterword.

The Spec offers the following fields:

* **Features**: A list of Features to set activated/deactivated.

There are two possible outcomes for the features listed in the spec:

* Applied - indicates that the feature intent has been successfully applied.
* Invalid - indicates that the feature intent specified in the spec is invalid.

### Example

This example FeatureGate will be used to toggle our big-cache Feature to activated
in our cluster, overriding its default.

```yaml
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: FeatureGate
metadata:
  name: featuregate-sample
spec:
  features:
    - name: big-cache
      activate: true
```

## Stability Level Policies

Every Feature has a stability level and that Feature should adhere to the policy
defined for that stability level. Below are the various stability levels that
are currently supported:

| Stability Level   | Default Activation State | Discoverable | State Immutable ? | Voids Warranty If Activated? | Stability Level Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
|-------------------|--------------------------|--------------|-------------------|------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Work In Progress  | false                    | false        | false             | true                         | Feature is still under development. It is not ready to be used, except by the team working on it. Activating this feature is not recommended under any circumstances.                                                                                                                                                                                                                                                                                                                           |
| Experimental      | false                    | true         | false             | true                         | Feature is not ready, but may be used in pre-production environments. However, if an experimental feature has ever been used in an environment, that environment will not be supported. Activating an experimental feature requires you to permanently, irrevocably void all support guarantees for this environment by setting permanentlyVoidAllSupportGuarantees in feature reference in featuregate spec to true. You will need to recreate the environment to return to a supported state. |
| Technical Preview | false                    | true         | false             | false                        | Feature is not ready, but is not believed to be dangerous. The feature itself is unsupported, but activating a technical preview feature does not affect the support status of the environment.                                                                                                                                                                                                                                                                                                 |
| Stable            | true                     | true         | true              | false                        | Feature is ready and fully supported                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| Deprecated        | true                     | true         | false             | false                        | Feature is destined for removal, usage is discouraged. Deactivate this feature prior to upgrading to a release which has removed it to validate that you are not still using it and to prevent users from introducing new usage of it.                                                                                                                                                                                                                                                          |

# Framework Features

Framework offers Features and Featuregates APIs to allow developers to have a
system to control rollout and availability of new Features in controller and
plugin logic.

With these primitives, Framework provides a path to progress your code from
early development through alpha, beta and GA in a deliberate manner.

## Key Concepts

- **Feature**: Functionality that is being introduced to a codebase which will
be gated based on Maturity level. Features can be activated/deactivated and follow the standard
 GVK versioning maturity. Features are created and managed by the API developers
creating the functionality.
- **FeatureGate**: The mechanism that provides a means to control the activation
state for a set of Features. FeatureGates are a resource that is managed by
Platform Operators.
- **Feature Activation**: The reconciled state of all Features. If a FeatureGate
exists that references a specific Feature, it is reconciled by the FeatureGate
to display the
correct Activations and Deactivations. If a FeatureGate doesn't exist, Features
will always default to their defined activation.
- **TKG SDK**: A collection of packages and tools provided by Framework intended
to provide developers with additional integration capabilities.

## Features API

The Features API provides controller and plugin developers a means to roll out
new code in a more predictable manner, using a standard durable API.

The [Features API Spec](apis/config/v1alpha1/feature_types.go) provides the
following concepts:

- Activation: Determine whether the Feature is enabled by default
- Maturity: The lifecycle path of the Feature: `dev`, `alpha`, `beta`, `ga`.
Most GA featues are destined to be removed as they become mainline code, but
some stay durable.
- Discoverability: Whether or not the FeatureGates will interact with the Feature
or not, or ignore it. Early development Features should always start off as not
discoverable, until they are at least alpha stability.

###  Example
In this example, we will define a discoverable development-level feature, which
will default to activated. It is mutable, so once set on a FeatureGate, it can
be toggled.
```
apiVersion: config.tanzu.vmware.com/v1alpha1
kind: Feature
metadata:
  name: big-cache
  namespace: tkg-system
spec:
  description: "A sample big cache Feature"
  immutable: false
  discoverable: true
  maturity: "dev"
  activated: false
```

## FeatureGates API

The [FeatureGates API](apis/config/v1alpha1/featuregate_types.go) provides a
means for operators to set Feature state in an eventually consistent,
level-triggered manner.

FeatureGates can be set at cluster creation time or anytime afterword.

FeatureGates are a Cluster-level resource. Its NamespaceSelector field
implements a [LabelSelector](https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#LabelSelector)
to determine which Namespaces their Features will affect.
By default, this is all namespaces.

Though the stock TKG FeatureGate defaults to controlling all namespaces,
operators can manually adjust this as needed to add additional FeatureGates in
whatever configuration makes sense.
Despite being highly configurable, it is anticipated that this default
FeatureGate will likely suffice for the foreseeable future.
The Spec offers the following fields:

- **NamespaceSelector**: Determine which NS range this gate controls. Defaults
to all namespaces.
- **Features**: A list of Features to set activated/deactivated.

### Example
This example FeatureGate will be used to toggle our big-cache Feature to on in
our cluster, overriding its default. Features must be marked discoverable to be
managed by the FeatureGate.
That said, all non-dev features should be marked discoverable.
```
apiVersion: config.tanzu.vmware.com/v1alpha1
kind: FeatureGate
metadata:
  name: featuregate-sample
spec:
  namespaceSelector:
    matchExpressions:
      - key: kubernetes.io/metadata.name
        operator: In
        values:
          - tkg-system
          - default
  features:
    - name: big-cache
      activate: true
```

## Feature Promotion Best Practices

The tools provided here were meant to be used to allow developers to release
Features deliberately.

The following table shows recommendations on how developers can use the API for
their own projects, to ensure their Features have a predictable path to GA.
That said,
the API was meant to provide flexibility for each Features individual needs, so
these are only suggestions.

| Maturity Level | Version Convention | Discoverable | Released |  Default Activation | Maturity Level Description |
|--- | --- | --- | --- | --- | --- |
| dev | <version>-*+build | false | false | Deactivated | Local and +build tags denote dev |
| alpha | <version>-alpha | False | True | Deactivated | Alpha releases have to be manually toggled to be discoverable to keep users safe|
| beta | <version>-beta | True | True | Deactivated | Beta Features can default to activated too, where it makes sense|
| ga | <version> | True | True | True | Activated | Offramp to removal of conditionals in code paths|
| deprecated| * | True | True | Conditional | Deprecated allows devs to express future intent|

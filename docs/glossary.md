# Tanzu Framework Glossary

**Condition:** A Kubernetes concept that communicates the status of a resource during its lifecycle. Conditions can include timestamps for these state changes and so provide a timeline of what happened to a resource. A condition may include a ‘reason’ string which can provide detail especially useful when troubleshooting.

**Discovery API:** How an API consumer discovers the APIs and API versions/revisions supported by a specific instance of the local control plane.

**FeatureGate:** The mechanism that provides a means to control the activation state for a set of Features. FeatureGates are a resource that is managed by Platform Operators.

**Feature Activation:** The reconciled state of all Features. If a FeatureGate exists that references a specific Feature, it is reconciled by the FeatureGate to display the correct Activations and Deactivations. If a FeatureGate doesn't exist, Features will always default to their defined activation.

**Finalizer:** A code function in a K8s controller resource used to help cleanup. Best practice is to provide a timestamp for deletion of objects to avoid further reconciliation overhead the controller needs to do for that object.

**Package** (also known as Addon or Extension): A unit of extensibility for a Kubernetes cluster that leverages the [Carvel](https://carvel.dev/) tools.  A single package is a combination of configuration metadata and OCI images that ultimately inform the package manager what software it holds and how to install it into a Kubernetes cluster.

**Predicate:** A defined condition, such a change in version, used to help filter events that controllers watch when trying to reconcile.

# Tanzu Framework Glossary

**Addon:** see definition of “Package.” Framework will be replacing this word with “package” over time.

**BOM - Bill of materials:** A manifest that identifies components and their versions in a release.

**Builder:** This helps plugin authors by scaffolding CLI plugin repositories for new plugins, and by building CLI plugins from code for specified architectures.

**Catalog:** Holds the information of all currently installed plugins on a host OS.

**Condition:** A Kubernetes concept that communicates the status of a resource during its lifecycle. Conditions can include timestamps for these state changes and so provide a timeline of what happened to a resource. A condition may include a ‘reason’ string which can provide detail especially useful when troubleshooting.

**Core, Auto-managed, and User-managed:** refers to the way in which the lifecycle of a package is managed. Note that these terms are defined for convenience and should not show up in the API or CLI.

* **Core:** Packages that are installed in clusters by default, typically because they’re required for basic cluster functionality. The versions for these packages are managed as part of the Tanzu Kubernetes release/Bill of Materials and managed in lock step with the version of Tanzu Kubernetes Grid (TKG) and/or Kubernetes. Thus, the version of these packages is automatically selected, and upgrades happen automatically when the cluster itself is upgraded.

* **Auto-managed:** Packages that a user opts to install, but for which the version of the package is automatically selected. Like core packages, upgrades happen automatically when the cluster itself is upgraded. Details of this will be determined by the linked patching effort.

* **User-managed:** Packages whose lifecycle is explicitly managed by a user. This would include packages delivered by TKG which are not automatically managed as part of TKG upgrades. Instead, users will manage the lifecycle for these packages: when to install, what version to install, when to upgrade, and what version to upgrade to.

**Discovery API:** How an API consumer discovers the APIs and API versions/revisions supported by a specific instance of the local control plane.

**Extension:** See definition of “Package”. Formerly a term used to mean the same thing.

**Extensibility:** The Tanzu CLI was built to be an extensible framework, allowing teams to add functionality via a plugin rather than a separate CLI, to enable a more cohesive user experience.

**FeatureGate:** The mechanism that provides a means to control the activation state for a set of Features. FeatureGates are a resource that is managed by Platform Operators.

**Feature Activation:** The reconciled state of all Features. If a FeatureGate exists that references a specific Feature, it is reconciled by the FeatureGate to display the correct Activations and Deactivations. If a FeatureGate doesn't exist, Features will always default to their defined activation.

**Finalizer:** A code function in a K8s controller resource used to help cleanup. Best practice is to provide a timestamp for deletion of objects to avoid further reconciliation overhead the controller needs to do for that object.

**Groups:** How plugin output is displayed. Currently Admin, Extra, Run, System, Version are displayed within groups.

**Hub and Spoke:** Employed by controller-runtime to cut down on the number of permutations of conversions between different versions. One version is designated as the “hub” (by implementing the Hub interface) and all the other versions define conversions to and from the “hub”.

**Management Package:** A package that’s intended to run in a management cluster; the unit of extensibility for the management cluster.

**Package** (also known as Addon or Extension): A unit of extensibility for a Kubernetes cluster that leverages the [Carvel](https://carvel.dev/) tools.  A single package is a combination of configuration metadata and OCI images that ultimately inform the package manager what software it holds and how to install it into a Kubernetes cluster.

**Plugin:** Unit of extensibility for CLI, delivered as a package and potentially referenced from a management package. The CLI consists of plugins, each of which shows up as either a resource or command (such as ‘login’ or ‘apps’. Plugins are developed in Go using the Cobra framework and conform to the Tanzu styleguide to ensure consistency between plugins.

**Predicate:** A defined condition, such a change in version, used to help filter events that controllers watch when trying to reconcile.

**Repository:** Represents a group of plugin artifacts that are installable by the Tanzu CLI. A repository is defined as an interface.

**TKG SDK:** A collection of packages and tools provided by Framework intended to provide developers with additional integration capabilities.

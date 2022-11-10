# Capability Discovery

## Table of Contents

* [Capability Discovery](#capability-discovery)
  * [Discovery Go Package](#discovery-go-package)
    * [Building a ClusterQueryClient](#building-a-clusterqueryclient)
    * [Building and Executing Queries](#building-and-executing-queries)
  * [Executing Pre-defined TKG queries](#executing-pre-defined-tkg-queries)
  * [Capability CRD](#capability-crd)
    * [Example Capability Custom Resource](#example-capability-custom-resource)

------------------------

The capability discovery Go package in `github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery` go module, along
with the `Capability` CRD offer the ability to query a cluster's capabilities. A "capability" is defined as anything a
Kubernetes cluster can do or have, such as objects and the API surface area. Capability discovery can be used to answer
questions such as `Is this a TKG cluster?`, `Does this cluster have a resource Foo?` etc.

## Discovery Go Package

The [`capabilities/client/pkg/discovery`](https://github.com/vmware-tanzu/tanzu-framework/tree/main/capabilities/client/pkg/discovery)
provides methods to query a Kubernetes cluster for the state of its API surface.

`ClusterQueryClient` allows clients to build queries to inspect a cluster and evaluate results.

The sections below illustrate how to build a client and query for APIs and objects.

### Building a ClusterQueryClient

Use the constructor(s) from `discovery` package to get a query client.

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/client/config"
    "github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
)

cfg := config.GetConfig()

clusterQueryClient, err := discovery.NewClusterQueryClientForConfig(cfg)
if err != nil {
    log.Error(err)
}
```

### Building and Executing Queries

Use `Group`, `Object` and `Schema` functions in the `discovery` package to build queries and execute them.

```go
import "github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"

// Define objects to query.
var pod = corev1.ObjectReference{
    Kind:       "Pod",
    Name:       "testpod",
    Namespace:  "testns",
    APIVersion: "v1",
}

var testAnnotations = map[string]string{
    "cluster.x-k8s.io/provider": "infrastructure-fake",
}

// Define queries.
var testObject = Object("podObj", &pod).WithAnnotations(testAnnotations)

var testGVR = Group("podResource", testapigroup.SchemeGroupVersion.Group).WithVersions("v1").WithResource("pods")

// Build query client.
c := clusterQueryClient.Query(testObject, testGVR)

// Execute returns combined result of all queries.
found, err := c.Execute()
if err != nil {
    log.Error(err)
}

if found {
    log.Info("Queries successful")
}

// Inspect granular results of each query using the Results method (should be called after Execute).
if result := c.Results().ForQuery("podResource"); result != nil {
    if result.Found {
        log.Info("Pod resource found")
    } else {
        log.Infof("Pod resource not found. Reason: %s", result.NotFoundReason)
    }
}
```

## Executing Pre-defined TKG queries

The `capabilities/client/pkg/discovery/tkg` package builds on top of the generic discovery package and exposes
pre-defined queries to determine a TKG cluster's capabilities.

Some examples are shown below.

```go
import tkgdiscovery "github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery/tkg"

c, err := tkgdiscovery.NewDiscoveryClientForConfig(cfg)
if err != nil {
    log.Fatal(err)
}

if c.IsTKGm() {
    log.Info("This is a TKGm cluster")
}

if c.IsManagementCluster() {
    log.Info("Management cluster")
}

if c.IsWorkloadCluster() {
    log.Info("Workload cluster")
}

if c.HasCloudProvider(ctx, tkgdiscovery.CloudProviderVsphere) {
    log.Info("Cluster has vSphere cloud provider")
}
```

## Capability CRD

Every TKG cluster starting from v1.4.0 includes a `Capability` CRD and an associated controller. Like the Go package
described above, a `Capability` CR can be used to craft queries to inspect a cluster's state and store the results the
CR's `status` field. `Capability` CRD's specification allows for different types of queries to inspect a cluster.

The full API can be found in [apis/core/v1alpha2/capability_types.go](../../apis/core/v1alpha2/capability_types.go)

### Example Capability Custom Resource

The following custom resource checks if the cluster is a TKG cluster which supports feature gating
abilities, and if it has NSX networking capabilities.

```yaml
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: Capability
metadata:
  name: tkg-capabilities
spec:
  serviceAccountName: my-service-account
  queries:
    - name: "tanzu-cluster-with-feature-gating"
      groupVersionResources:
        - name: "tanzu-resource"
          group: "run.tanzu.vmware.com"
          versions:
            - v1alpha1
          resource: "tanzukubernetesreleases"
        - name: "featuregate-resource"
          group: "config.tanzu.vmware.com"
          versions:
            - v1alpha1
          resource: "featuregates"
    - name: "nsx-support"
      objects:
        - name: "nsx-namespace"
          objectReference:
            kind: "Namespace"
            name: "vmware-system-nsx"
            apiVersion: "v1"
```

To execute the queries with the above CR, a serviceAccountName needs to be specified to give capabilities controller
enough privileges to query for resources. Refer to [Security model](#security-model) to understand how a ServiceAccount
is used to query for resources.

The capabilities controller:

1. Watches `Capability` resources that are created or updated.
1. Executes queries specified in the spec.
1. Writes the results to the status field of the resource.

After reconciliation, results can be inspected by looking at the status field. Results are grouped by GVK, Object and
Partial Schema queries, and provide a predictable data structure for consumers to parse. They can be accessed by the
paths `status.results.groupVersionResources`, `status.results.objects` and `status.results.partialSchemas` respectively.

An example of query results is shown below.

```yaml
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: Capability
metadata:
  name: tkg-capabilities
spec:
  # Omitted
status:
  results:
  - groupVersionResources:
    - found: true
      name: tanzu-resource
    - found: true
      name: featuregate-resource
    name: tanzu-cluster-with-feature-gating
  - name: nsx-support
    objects:
    - found: false
      name: nsx-namespace
```

### Security Model

Capabilities controller container runs with a service account that has access to all service accounts and secrets in the
cluster. This service account is not used for querying resources. If a user is querying for objects, then each
Capabilities CR must specify a service account to allow the Capabilities CR owner to query for only resources they have
access to. This avoids the problem of privilege escalation by not relying on the shared service account to query for
resources. The additional benefit of users specifying the service account is they can query for more resources than
what the shared service account has access to. But, if the user is querying for the existence of a GVR, then
service account name is not needed as part of the spec as the Capabilities controller uses a default service account
(`tanzu-capabilities-manager-default-sa`) which doesn't have any permissions to access cluster resources to execute
those queries.

**Ex 1:**

If you as a user want to query for a particular resource, for example a pod named `nginx` in `foo` namespace, create a
Role, RoleBinding and add the ServiceAccount as a subject in the RoleBinding and specify the ServiceAccount in the
Capability CR.

Create RBAC rules:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-sa
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: foo
  name: pod-reader
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: my-role-binding
  namespace: foo
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: pod-reader
subjects:
  - kind: ServiceAccount
    name: my-sa
    namespace: default
```

Create a Capability CR in the same namespace as the ServiceAccount(in this case `default` namespace).

```yaml
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: Capability
metadata:
  name: nginx-capability
spec:
  serviceAccountName: my-sa
  queries:
    - name: "Query for nginx pod"
      objects:
        - name: "query nginx"
          objectReference:
            kind: "Pod"
            name: "nginx"
            apiVersion: "v1"
            namespace: "foo"
```

**Ex 2:**

If you are just querying to check the existence of a GVR, lets say FeatureGate API, then you need not specify the
service account name in the Capability CR.

```yaml
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: Capability
metadata:
  name: tkg-capabilities
spec:
  queries:
    - name: "tanzu-cluster-with-feature-gating"
      groupVersionResources:
        - name: "featuregate-resource"
          group: "config.tanzu.vmware.com"
          versions:
            - v1alpha1
          resource: "featuregates"
```

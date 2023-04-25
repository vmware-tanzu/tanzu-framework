# Tanzu Framework Controllers Using Kubebuilder

## Table of Contents

* [Summary](#summary)
* [Goals](#goals)
* [Non Goals](#non-goals)
* [Building Framework Controllers](#building-Framework-controllers)
  * [Create a new API and Controller](#create-a-new-api-and-controller)
  * [Change API and Controller](#change-api-and-controller)
  * [Generate Deepcopy Functions and Manifests](#generate-deepcopy-functions-and-manifests)
* [Advanced Use Cases With Builder](#advanced-use-cases-with-builder)
  * [Watching and Owning Multiple Resource Types](#watching-and-owning-multiple-resource-types)
  * [Event Filtering with Predicates](#event-filtering-with-predicates)
* [Writing Conversion Webhooks](#writing-conversion-webhooks)
  * [Choosing the Hub Version](#choosing-the-hub-version)
  * [Conversion Round Trip-ability](#conversion-round-trip-ability)
  * [Storage Version Migration](#storage-version-migration)
* [Best Practices](#best-practices)
  * [CRD Changes](#crd-changes)
  * [Validation and Defaulting](#validation-and-defaulting)
  * [Requeue and Rate Limiting](#requeue-and-rate-limiting)
  * [Logging](#logging)
  * [Metrics](#metrics)
  * [Events](#events)
  * [Status Conditions](#status-conditions)
  * [OwnerReference](#ownerreference)
  * [Finalizers for Deletion](#finalizers-for-deletion)
* [References](#references)

## Summary

This document provides guidance for controller development in `vmware-tanzu/tanzu-framework` repo using Kubebuilder and
outlines some best practices.

## Goals

* Standardize Framework controller development using Kubebuilder.
* Provide guidance on how to get started with writing new controllers.
* Provide best practices for controller development and webhook conversion.

## Non Goals

* Provide guidance on resource naming and versioning.

## Building Framework Controllers

The `tanzu-framework` repo is already set up for using Kubebuilder with the domain `tanzu.vmware.com`, which means all the resources
created in `tanzu-framework` will end with `tanzu.vmware.com` in its API group.

### Create a new API and Controller

To create a new API and controller with Kubebuilder, do the following:

```shell
$ kubebuilder create api --group foo --version v1alpha1 --kind Bar
Create Resource [y/n]
y
Create Controller [y/n]
y
Writing scaffold for you to edit...
apis/foo/v1alpha1/bar_types.go
controllers/foo/bar_controller.go
```

### Change API and Controller

API changes go in the file `apis/foo/v1alpha1/bar_types.go`, where the kind `Bar` is defined. Typically, each kind has
a `spec` and a `status`. The `spec` specifies the desired state of an object and `status` represents the observed state
of the object by the controller.

Controller changes go in the file `controllers/foo/bar_controller.go`. This is where the reconciliation of the desired
state of the object happens, so this is the place where most of the business logic is written.

### Generate Deepcopy Functions and Manifests

Every custom resource (and builtin) object in Kubernetes implements
the [`runtime.Object`](https://pkg.go.dev/k8s.io/apimachinery@v0.20.5/pkg/runtime#Object) interface. A requirement for
implementing the interface is to implement the `DeepCopyObject()` method, which defines how to make a deep copy of the
object. To avoid API authors hand-writing deep copy functions, Kubernetes and controller-runtime provide utilities for
automatically generating deep copy functions.

To generate deep copy functions for the API, do the following:

```shell
$ make generate
~/go/bin/controller-gen object:headerFile="hack/boilerplate.go.txt",year=2021 paths="./..."
```

This results in a deep copy file called `apis/foo/v1alpha1/zz_generated.deepcopy.go`.

Kubebuilder can also generate Kustomize manifests for deploying the controller(s) on a Kubernetes cluster.

To generate manifests for the controller, do:

```shell
$ make manifests
~/go/bin/controller-gen \
"crd" \
paths=./apis/... \
output:crd:artifacts:config=config/crd/bases
```

Note that deepcopy functions and manifests need to be generated every time there is a change to the API or markers.

## Advanced Use Cases With Builder

### Watching and Owning Multiple Resource Types

The [`Builder`](https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/builder#Builder) pattern allows the
controller to watch or own multiple resource types. This flexibility allows the controller to manage lifecycle of a
different resource or respond to events of arbitrary resources. For example, the `ReplicaSet` controller creates and
watches `Pod` objects. When the `Pod` objects change, it can trigger a reconcile to the parent `ReplicaSet` object.

`Builder` functions allow controller authors to setup watches for different resource types and respond to changes.

This example from cluster-api (CAPI) illustrates that the MachineSet controller responds to both events of
`clusterv1.MachineSet` resource and the `clusterv1.Machine` resources it owns. The map function allows you to map which
`clusterv1.MachineSet` custom resource is reconciled when a particular `clusterv1.Machine` resource is updated.

```go
c, err := ctrl.NewControllerManagedBy(mgr).
    For(&clusterv1.MachineSet{}).
    Owns(&clusterv1.Machine{}).
    Watches(
        &source.Kind{Type: &clusterv1.Machine{}},
        &handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(r.MachineToMachineSets)},
    ).
    WithOptions(options).
    WithEventFilter(predicates.ResourceNotPaused(r.Log)).
    Build(r)
```

### Event Filtering with Predicates

With predicates, you can filter events that result in a reconcile. For example, the following code illustrates filtering
out events that don’t result in resource version changes for the Bar custom resource.

```go
func (r *BarReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
            For(&foov1alpha1.Bar{}).
            WithEventFilter(predicate.ResourceVersionChangedPredicate{}).
            Complete(r)
}
```

## Writing Conversion Webhooks

To handle API version changes, Kubebuilder provides a way to
scaffold [conversion webhooks](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#webhook-conversion)
to convert custom resources between API versions. To cut down on the number of permutations of conversions between
different versions, controller-runtime employs a “hub and spoke” model. One version is designated as the “hub” (by
implementing the Hub interface) and all the other versions define conversions to and from the “hub”.

Note that regardless of multiple versions of an API, only one version is ever stored in API server/etcd and that is
called the “storage version” of the resource.

### Choosing the Hub Version

Kubebuilder docs provide an example of a multi-version API going from v1 -> v2 -> v3, where v1 is designated as the hub
version. This means your controller is likely operating on v1 objects i.e. doing a GET for the v1 version of resource
where the webhook may be converting to v1 if the storage version is different and using v1 imports in the code.

In practice, API authors might not want to designate an old version as the hub. This is because when introducing newer
versions, features are added via new fields to the custom resource and the controller is updated to handle new features.
Converting to and operating on an old version might mean losing or discarding new information.

Depending on what your controller does, care must be taken when designating a hub version. In most cases, it’s better to
designate the newest version you are introducing as the hub and update your controller imports to work on the newest
version of the custom resource. The downside is that this will result in more changes to the conversion functions.

### Conversion Round Trip-ability

In Kubernetes, all versions of a resource must be round trippable through each other with no information loss. In other
words, conversions should not result in losing any field’s value on a round trip.

This gets into unwieldy territory when there are new fields added in a new API version and the old versions do not/need
not support these fields. This means when you convert a v2 resource with new fields to v1 and convert back to v2, you
could lose information about the new fields since v1 would not know about them. The common solution for this scenario is
to adjust the v2 -> v1 conversion functions to carry those fields as annotations on the v1 resource and set those fields
in v2 objects using the annotations in the v1 -> v2 conversion functions.

To make writing conversions easier, Kubernetes provides a handy tool
called [`conversion-gen`](https://pkg.go.dev/k8s.io/code-generator/cmd/conversion-gen) for auto-generating conversion
functions for structs and fields that have the same name across two versions of the resource. Cluster API provides a
nice [example](https://github.com/kubernetes-sigs/cluster-api/blob/102c753ec14c3f6ebfbce7084e1176eb62e6d80d/Makefile#L261)
of using conversion-gen to generate conversion functions.

### Storage Version Migration

There is only one storage version for a resource in Kubernetes. In Kubebuilder, the storage version can be set using the
`//+kubebuilder:storageversion` marker. When the storage version changes for a resource, existing objects in etcd are
not automatically converted to the new storage version. But newly created or updated objects are written at the new
storage version.

When you are deprecating and dropping an old version, you will need to make sure there are no objects stored in that
version.
Upstream [docs](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#upgrade-existing-objects-to-a-new-stored-version)
offer a couple of options to handle that migration. But in most cases, when you upgrade a controller which works
exclusively on the newest API version (as mentioned in the “hub” section and assuming the new version is also set as the
storage version), the objects will be automatically converted when the new controller comes up, lists all existing
objects as part of “list and watch” and writes it back at the new storage version using the conversion functions you’ve
written. That is, the controller asks for objects in the new version, the webhook converts all objects to the new
version, and the controller writes it back to the new version which is also the storage version now.

Also note that when you no longer want to support an API version, you can use the `//+kubebuilder:unservedversion`
marker to not serve that version i.e. you will get an error if you try to GET or POST an object with that version.

## Best Practices

### CRD Changes

* Changing CRD properties (short names, categories, print columns etc) must be done via
  Kubebuilder [markers](https://book.kubebuilder.io/reference/markers/crd.html) and never by hand. Always consider
  backward compatibility and user experience when changing properties.

### Validation and Defaulting

* API authors should add [validation markers](https://book.kubebuilder.io/reference/markers/crd-validation.html) for
  fields wherever applicable, so it’s reflected in the generated OpenAPI schema. Kubebuilder markers provide basic
  validation such as required, max, min, enum etc.
* In addition to OpenAPI schema
  validation, [validation webhooks](https://kubebuilder.io/cronjob-tutorial/webhook-implementation.html) should be
  written for APIs for fields that cannot be easily validated by the markers. Example: Is this subnet CIDR field valid?
  Are these two fields mutually exclusive?
* For optional fields, sane defaults must be provided
  via [defaulting webhooks](https://kubebuilder.io/cronjob-tutorial/webhook-implementation.html) wherever applicable.
* Note that validation and defaulting webhooks should only be defined for the “hub” version. This also means that if a
  new version v2 is added, the webhooks should be moved to the v2 package and updated to work on the v2 type.

### Requeue and Rate Limiting

* Use `RequeueAfter` when
  returning [`ctrl.Result`](https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile#Result) from your
  reconcile function for requests you don’t need to be requeued immediately.
* Decide if you want to reconcile periodically or only when an object changes. For controllers that are managing
  external resources, it might make sense to reconcile periodically to account for out-of-band changes to external
  resources (example: VM is deleted and needs to be recreated). You can use `RequeueAfter` to manage this behavior.
* The default workqueue rate limiter used for controlling rate of reconcile for a custom resource may not be appropriate
  for all use cases. Depending on what the controller is doing, the default exponential backoff time may be too low (
  resulting in a lot of quick reconciles which might overload an external API that’s being called) or too high (large
  delays between reconciles that cause not to react to an external object’s changes).

In such cases, you can use a different workqueue rate limiter.
Client-go [provides](https://github.com/kubernetes/client-go/blob/master/util/workqueue/default_rate_limiters.go) some
rate limiters for use. But if you need a more specific rate limiting behavior, you can write your own by implementing
the [`Ratelimiter`](https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/ratelimiter#RateLimiter)
interface and passing it as an [option](https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/controller#Options)
to your controller. For example:

```go
func (r *BarReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
            For(&foov1alpha1.Bar{}).
            WithOptions(controller.Options{
                RateLimiter: workqueue.DefaultItemBasedRateLimiter(),
            }).
            Complete(r)
}
```

### Logging

* Use structured logging.
* Controller-runtime expects a [logr](https://github.com/go-logr/logr) logger implementation and provides helpers to set
  zap as a logr backend in [`pkg/log`](https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/log/zap).

### Metrics

* By default,
  controller-runtime [publishes](https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/metrics#pkg-constants)
  some performance metrics for each controller and exposes a prometheus metrics
  server.
* Controller authors can also publish relevant additional metrics using
  the [`pkg/metrics`](https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/metrics) package.

### Events

* Controller-runtime makes an `EventRecorder` available to emit events from within the controller.
* Events are useful for observability when users do a `kubectl describe` on a resource. Controller authors can create
  important events in the reconcile loop (such as VM created/deleted), including warnings that provide easier to digest
  information than users digging through logs.

### Status Conditions

* Use status [conditions](https://dev.to/maelvls/what-the-heck-are-kubernetes-conditions-for-4je7) to indicate readiness
  status of the resource that is being reconciled. Conditions provide a good user experience for knowing the status of a
  resource.

### OwnerReference

* If you are creating resources owned by the main resource from within the controller and if you need the owned resource
  to be deleted when the main resource is deleted, you will need to set an owner reference using
  the [`ctrl.SetControllerReference`](https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/controller/controllerutil#SetControllerReference)
  function.

### Finalizers for Deletion

* Finalizers let controllers run pre-delete hooks. Any controller which creates and manages external resources must add
  finalizer(s) and must clean up the external resources before the custom resource is deleted to avoid leakage.
* A custom resource with a finalizer set will not be immediately deleted. Instead, it’s DeletionTimestamp is set. The
  controller can check if the operation was a DELETE by checking if `bar.ObjectMeta.DeletionTimestamp.IsZero()`
  is `false`.
* Finalizers must have a meaningful name with the group version appended to avoid collisions since they are just a
  string slice attached to the custom resource object. Example: `virtualmachine.finalizers.foo.tanzu.vmware.com`.

## References

1. [controller-runtime FAQ](https://github.com/kubernetes-sigs/controller-runtime/blob/master/FAQ.md)
1. [Kubebuilder book](https://book.kubebuilder.io)
1. [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)

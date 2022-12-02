# Tanzu Framework API Machinery Guide

The purpose of this walkthrough is to provide a good starting place for those writing new features atop Framework.

The primary aim is to introduce the APIs and tools available in Framework to developers of new APIs.

We will use stock Kubebuilder and manually setup a Feature, and the Query for the controller via Capabilities API.

There are experimental Kubebuilder plugins for Framework under development which will remove the majority of the manual
toil below, but this guide will continue to walk through this code manually to ensure the concepts are clear.

These new APIs are documented here:

* [Features](../features-and-featuregates.md#features-api)
* [FeatureGate](../features-and-featuregates.md#featuregates-api)
* [Capability](../capability-discovery.md)

## Generate a new API

We recommend using [Kubebuilder](https://kubebuilder.io/) and [Controller-Runtime](https://github.com/kubernetes-sigs/controller-runtime)
based tools to create your new Resources, but the good news is that the Features API and Capabilities will probably work
with any method of creating and running controllers.

First, initialize a new project.

```sh
kubebuilder init --domain mydomain.com
```

Lets generate a new API using stock Kubebuilder plugins:

```sh
kubebuilder create api --group example --version v1alpha1 --kind MegaCache
```

## Add a Feature

You may reference [this example of the types file](examples/megacache_types.go.sample).

Open the new API type file in your editor of choice. One way to define a feature and manage its lifecycle is via
feature tags and our codegen plugin.

```sh
vim api/v1alpha1/megacache_types.go
```

Add the following line to your types file one line above the type itself.

```go
//+tanzu:feature:name=mega-cache,stability=Technical Preview
```

This tag states that this Type makes use of a feature `mega-cache` with stability level as `Technical Preview`, which is
mutable, discoverable, deactivated by default and does not void support warranty of the environment if activated.

More info on these levels is available [here](../features-and-featuregates.md##stability-level-policies).

## Modify the type

Also in the types file, add a field to your new MegacacheSpec type named CacheSize.
This will dictate the size of our experimental cache.

```go
type MegaCacheSpec struct {
        Size int `json:"bool,omitempty"`
}
```

## Register FeatureGate Scheme

To do so, open `main.go` in the editor of your choice.

Here, the FeatureGate Kind needs be registered in Scheme for type v1alpha2.FeatureGate after importing the appropriate package.

The init function should look like below after adding the appropriate Kinds to the scheme:

```go
import corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
...

func init() {
        utilruntime.Must(clientgoscheme.AddToScheme(scheme))
        utilruntime.Must(corev1alpha2.AddToScheme(scheme))
        utilruntime.Must(examplev1alpha1.AddToScheme(scheme))
        //+kubebuilder:scaffold:scheme
}
```

## Modify the Controller

You may reference [this example of the controller code](examples/controller.go.sample).

Next, the controller code needs some updating. We want it to provide a manager that listens for Feature updates,
and we want to provide a simple means of checking the feature in our reconciler function.

### Update Imports

First add the import for the Framework Core API package and the SDK FeatureGate package.

```go
corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
gate "github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/featuregateclient"
```

### Update Manager

We need to update the manager setup to listen for the Feature updates.

It should look like this afterword:

```go
// SetupWithManager sets up the controller with the Manager.
func (r *MegaCacheReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&mygroupv1alpha1.MegaCache{}).
        Watches(&source.Kind{Type: &corev1alpha2.Feature{}}, eventHandler(mgr.GetClient())).
        Complete(r)
}

func eventHandler(c client.Client) handler.EventHandler {
    return handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
        MegaCacheList := &mygroupv1alpha1.MegaCacheList{}
        if err := c.List(context.Background(), MegaCacheList); err != nil {
            log.Infof("list-failed err=%q", err)
        }
        var requests []reconcile.Request
        for _, item := range MegaCacheList.Items {
            requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
                Name:      item.Name,
                Namespace: item.Namespace,
            }})
        }
        return requests
    })
}
```

### Manage Feature

Now we can control our controller behavior based on Feature status.

To do so, in this case we will add the following functionality in the reconcile logic.

This code will essentially toggle the logic of the entire controller, but your code should be more granular.

```go
enabled, err := gate.IsFeatureActivated(ctx, r.Client, "megacache")
if err != nil {
        return ctrl.Result{}, err
}
if !enabled {
        log.Info("feature deactivated")
        return ctrl.Result{}, nil
}
```

## Generate and Install

Now you can generate your code and manifests as normal. After this, your directory will be populated with
generated code and manifests that reference them in `config/`.

```sh
make generate
make manifests
```

## Run the controller

From here, run `make run` to boot your dev controller for development purposes. This boots your controller in your local
environment and does not require building or pushing docker images.

Open a new console tab, and leave this controller running for now.

We are going to toggle the Feature and watch the output in this console.

```sh
make run
```

### Deploying to Real Clusters

To use the controller and API outside development, you will need to provide an authenticated docker repository in the
env variable `IMG`, and then the docker images can be built with:

```sh
export IMG=your.docker.image.tag

make docker-build && make docker-push
```

The deploy command will deploy the controller to the K8s cluster specified in `~/.kube/config`:

```sh
make deploy
```

These commands are not required for the purposes of this guide.

## Create an Example MegaCache

Lets create an example MegaCache resource. To do so, copy the follow YAML and apply it to your cluster:

```yaml
apiVersion: example.mydomain.com/v1alpha1
kind: MegaCache
metadata:
  name: mycache
  size: 10000
```

Once applied, we can query for this using Capabilities.

## Create a Capability resource

As described above, Capability Discovery provides an API to query a Kubernetes cluster about its current state. We will do so
now to determine if our new API exists.

We can create a Capability to define whether our new resource exists. Create the following in a text file:

```yaml
apiVersion: run.tanzu.vmware.com/v1alpha1
kind: Capability
metadata:
  name: my-megacache
spec:
  queries:
    - name: "megacache-v1alpha1"
      groupVersionResources:
        - name: "megacache"
          group: "example.mydomain.com"
          versions:
            - v1alpha1
```

The above resource defines a query that determines if this cluster provides the megacache v1alpha1 GVR.

Once applied, this resource will provide us with an example we can use Capability Discovery to query for.

### Install Capability Resource

Now kubectl apply this resource:

```sh
kubectl apply -f megacache-capability.yaml
```

And view the Status to see the state of the queries!

```sh
kubectl get capability my-megacache -o yaml
```

We can see that our new Capability clearly denotes that our capability exists - both the API and the my-cache object are present!

```yaml
status:
  results:
  - groupVersionResources:
    - found: true
      name: megacache
    name: megacache-v1alpha1
```

## Summary

Now we have created a new Feature, a new API and used a new Capability to assert their presence. Great job!

# Readiness Framework

The readiness framework provides functionality to define and evaluate k8s clusters for suitability of running enterprise workloads.

## Key Concepts

* **Readiness**: Readiness of a cluster is a measure of whether that cluster can make that set of capabilities available to application workloads running on that cluster. These capabilities can be related to areas like security, compliance, scalability and resiliency.
* **ReadinessChecks**: A Readiness definition constitutes one or more Readiness Checks. Each check relies on availablility of atleast one **active** ReadinessProviders for fulfilment.
* **ReadinessProviders**: The Readiness Provider defines the set of checks that it fulfills along with the conditions that should be evaluated to true.
* **ReadinessProviderConditions**:  The Readiness Condition is one of the basic building blocks of the Readiness definition. It can convert the state of the cluster to a boolean value.

## Readiness API

The Readiness API allows users to define cluster evaluation criterion as a set of readiness checks. For a readiness resource to be "ready", all the constituent checks must be fulfilled.

Readiness checks can be of 2 types:

1. Basic: These checks rely on the state of cluster for their fulfillment.
2. Composite: These checks rely on a group of other basic checks for their fulfilment.

### Example

The following manifest defines a Readiness resource with 2 checks.
These checks are required to be satisfied by atleast one **active** ReadinessProvider (See [ReadinessProvider Example](#example-1)), so that `my-org-baseline` can be evaluated to ready.

```yaml
---
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: Readiness
metadata:
  name: my-org-baseline
spec:
  checks:
    - category: Security
      name: com.vmware.tanzu.certificate-management
      type: basic
    - category: Packaging
      name: com.vmware.tanzu.package-management
      type: basic
```

## ReadinessProvider API

The ReadinessProvider API allows users to define a set of conditions. These conditions map the state of the cluster to a boolean value. A logical AND of all the ReadinessProviderConditions determines whether the ReadinessProvider is active.

### Example

The above manifest creates 2 ReadinessProvider resources which will evaluate if `cert-manager` and `kapp-controller` are available in the cluster. The providers also specify `checkRefs` which will aid in making the `my-org-baseline` Readiness resource **_ready_**.

```yaml
---
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: ReadinessProvider
metadata:
  name: cert-manager-provider
spec:
  checkRefs:
    - com.vmware.tanzu.certificate-management # This is the reference to check name defined in one of the readiness resources
  conditions:
    - name: certificate-crd
      resourceExistenceCondition:
        apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        name: certificates.cert-manager.io

---
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: ReadinessProvider
metadata:
  name: kapp-provider
spec:
  checkRefs:
    - com.vmware.tanzu.package-management
  conditions:
    - name: kapp-controller-deployment
      resourceExistenceCondition: # Namespaced resource
        apiVersion: apps/v1
        kind: Deployment
        name: kapp-controller
        namespace: kapp-controller
    - name: kapp-crd
      resourceExistenceCondition: # Clustered resource
        apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        name: apps.kappctrl.k14s.io
```

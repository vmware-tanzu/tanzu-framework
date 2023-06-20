# Usage Guide

## Abstractions

### Readiness Check

An organization can have a set of approved checks. Each check is a free form string. Some examples of checks are as follows:

1. com.org1.k8s.certificate-management
2. com.org1.k8s.fips

### Readiness Definition

Readiness definition is represented by the CRD `Readiness`. Readiness definition is a collection of checks. Any given Readiness definition is said to be ready if all the mentioned checks are true.

Given k8s cluster can have multiple Readiness definitions.

Any given check can be present in zero or more readiness definitions. This means, an approved check may not be present in any of the readiness definitions in the cluster. Also, any given check can be present in more than one readiness defintions.

### Readiness Provider

Readiness provider is represented by the CRD `ReadinessProvider`. The readiness provider has a collection of conditions that will be evaluated by the readiness controller and a set of readiness checks that will be set to true if all the conditions evaluate to true.

A readiness provider can provide zero or more readiness checks. For example, a cluster essentials provider can provide the checks `secretgen` and `kapp-controller`.

A readiness provider has zero or more conditions. The provider evaluates to `Success` if there are no conditions defined. The provider evaluates to `Success` if there is at least one condition defined and all the defined conditions evaluate to `true`.

## Example

### Checks

Let's assume we have the following checks approved by the organization org1

1. com.org1.k8s.package-management
2. com.org1.k8s.secret-management
3. com.org1.k8s.certificate-management

### Service Account

For the readiness providers to be able to query various reources, a service account which has required role bindings can be provided in the spec.
A sample yaml is defined below, which grants permissions to read CRDs. For creating these resources, run `kubectl apply -f <filename>`. We'll be referring to the details of the created service account in the following sections.

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: crd-read-sa
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: crd-read-role
  namespace: default
rules:
  - apiGroups:
    - "apiextensions.k8s.io"
    resources:
      - customresourcedefinitions
    verbs:
      - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: crd-read-rolebinding
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: crd-read-role
subjects:
  - kind: ServiceAccount
    name: crd-read-sa
    namespace: default

```

### Readiness Providers

Now, let's defined three readiness providers, one for each of the above checks.

#### Package Management Provider

The following readiness provider manifest has one `checkRef` for `com.org1.k8s.package-management`. This denotes that this provider provides the readiness check `com.org1.k8s.package-management`.

Also note that the following manifest contains four different conditions where each condition check for existence of a CRD with a certain name. In this example, we assume that the package management is ready if these four CRDs are present in the cluster.

```yaml
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: ReadinessProvider
metadata:
  name: kapp-controller
spec:
  checkRefs:
  - com.org1.k8s.package-management
  conditions:
  - name: internal-package-metadata
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: internalpackagemetadatas.internal.packaging.carvel.dev
  - name: internal-package
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: internalpackages.internal.packaging.carvel.dev
  - name: package-install
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: packageinstalls.packaging.carvel.dev
  - name: package-repository
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: packagerepositories.packaging.carvel.dev
  serviceAccount:
    name: crd-read-sa
    namespace: default
```

Save the above manifest in a file and run `kubectl apply -f <filename>` to deploy it on the Kubernetes cluster where the readiness framework is already installed.

#### Secret Management Provider

Similar to package management provider, install the secret management provider into the cluster by applying the manifest given below.

```yaml
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: ReadinessProvider
metadata:
  name: secretgen
spec:
  checkRefs:
  - com.org1.k8s.secret-management
  conditions:
  - name: certificate
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: certificates.secretgen.k14s.io
  - name: password
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: passwords.secretgen.k14s.io
  - name: rsakey
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: rsakeys.secretgen.k14s.io
  - name: secretexport
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: secretexports.secretgen.carvel.dev
  - name: secretimport
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: secretimports.secretgen.carvel.dev
  - name: sshkey
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: sshkeys.secretgen.k14s.io
  - name: secrettemplate
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: secrettemplates.secretgen.carvel.dev
  serviceAccount:
    name: crd-read-sa
    namespace: default
```

#### Certificate Management Provider

The manifest for the certificate management provider is given as follows. Install this manifest to the cluster

```yaml
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: ReadinessProvider
metadata:
  name: cert-manager
spec:
  checkRefs:
  - com.org1.k8s.certificate-management
  conditions:
  - name: certificate
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: certificates.cert-manager.io
  - name: certificate-request
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: certificaterequests.cert-manager.io
  - name: challenge
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: challenges.acme.cert-manager.io
  - name: cluster-issuer
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: clusterissuers.cert-manager.io
  - name: issuer
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: issuers.cert-manager.io
  - name: order
    resourceExistenceCondition:
      apiVersion: apiextensions.k8s.io/v1
      kind: CustomResourceDefinition
      name: orders.acme.cert-manager.io
  serviceAccount:
    name: crd-read-sa
    namespace: default
```

### Readiness Definition

In this example, let's define three readiness definitions namely `dev`, `canary` and `prod` as shown below.

| Readiness Definition   |      Required Checks                                          |
|------------------------|---------------------------------------------------------------|
| dev                    | package-management                                            |
| canary                 | package-management, secret-management                         |
| prod                   | package-management, secret-management, certificate-management |

The manifests for the above readiness definitions are as follows

```yaml
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: Readiness
metadata:
  name: dev-class
spec:
  checks:
  - name: com.org1.k8s.package-management
    type: basic
    category: Packaging  
---
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: Readiness
metadata:
  name: canary-class
spec:
  checks:
  - name: com.org1.k8s.package-management
    type: basic
    category: Packaging
  - name: com.org1.k8s.secret-management
    type: basic
    category: Security
---
apiVersion: core.tanzu.vmware.com/v1alpha2
kind: Readiness
metadata:
  name: prod-class
spec:
  checks:
  - name: com.org1.k8s.certificate-management
    type: basic
    category: Security
  - name: com.org1.k8s.package-management
    type: basic
    category: Packaging
  - name: com.org1.k8s.secret-management
    type: basic
    category: Security
```

Let's add the given manifests to the kubernetes cluster.

### Install packages

Before proceeding further into the demo, let's check the status of each of the resources that we have created.

```bash
shell> kubectl get readiness,readinessprovider
NAME                                           READY   AGE
readiness.core.tanzu.vmware.com/canary-class   false   45s
readiness.core.tanzu.vmware.com/dev-class      false   45s
readiness.core.tanzu.vmware.com/prod-class     false   45s

NAME                                                      STATE     AGE
readinessprovider.core.tanzu.vmware.com/cert-manager      failure   63s
readinessprovider.core.tanzu.vmware.com/kapp-controller   failure   104s
readinessprovider.core.tanzu.vmware.com/secretgen         failure   77s
```

We can see all the providers are in failure state and all the readiness definition are not ready.

#### Install kapp-controller

Install kapp-controller into the cluster by running the following command.

```bash
kubectl apply -f https://github.com/vmware-tanzu/carvel-kapp-controller/releases/latest/download/release.yml
```

Few minutes after installation, query the readiness resources. We can see the kapp-controller transitioned to success state and the dev-class is set to ready.

```bash
shell> kubectl get readiness,readinessprovider
NAME                                           READY   AGE
readiness.core.tanzu.vmware.com/canary-class   false   5m26s
readiness.core.tanzu.vmware.com/dev-class      true    5m26s
readiness.core.tanzu.vmware.com/prod-class     false   5m26s

NAME                                                      STATE     AGE
readinessprovider.core.tanzu.vmware.com/cert-manager      failure   5m44s
readinessprovider.core.tanzu.vmware.com/kapp-controller   success   6m25s
readinessprovider.core.tanzu.vmware.com/secretgen         failure   5m58s
```

#### Install secrentgen-controller

Install secretgen-controller into the cluster using the following command.

```bash
kubectl apply -f https://github.com/carvel-dev/secretgen-controller/releases/latest/download/release.yml
```

After few minutes, we can see the secretgen provider in success state and the canary class (in addition to dev class) is marked as ready.

```bash
shell> kubectl get readiness,readinessprovider
NAME                                           READY   AGE
readiness.core.tanzu.vmware.com/canary-class   true    9m33s
readiness.core.tanzu.vmware.com/dev-class      true    9m33s
readiness.core.tanzu.vmware.com/prod-class     false   9m33s

NAME                                                      STATE     AGE
readinessprovider.core.tanzu.vmware.com/cert-manager      failure   9m51s
readinessprovider.core.tanzu.vmware.com/kapp-controller   success   10m
readinessprovider.core.tanzu.vmware.com/secretgen         success   10m
```

#### Install cert-manager

Install cert-manager using the following command.

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
```

In few minutes, we can see all the providers in success state and all the readiness definitions in ready state.

```bash
shell> kubectl get readiness,readinessprovider
NAME                                           READY   AGE
readiness.core.tanzu.vmware.com/canary-class   true    11m
readiness.core.tanzu.vmware.com/dev-class      true    11m
readiness.core.tanzu.vmware.com/prod-class     true    11m

NAME                                                      STATE     AGE
readinessprovider.core.tanzu.vmware.com/cert-manager      success   11m
readinessprovider.core.tanzu.vmware.com/kapp-controller   success   12m
readinessprovider.core.tanzu.vmware.com/secretgen         success   12m
```

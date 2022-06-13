Creating clusters with cluster-class
========================================

This developer handbook describe how to create a cluster using a clusterclass based CRD. 


## Set up and host a package repository for management cluster
First, make sure you have the push access to an OCI registry. Export its URL as `OCI_REGISTRY` environment variable.

### Set up an OCI Registry with GCP

For example, let's suppose we have a project with ID `my-project-1527816345739` on GCP. Then we can set the registry URL
as `gcr.io/my-project-1527816345739/tkg/management`. Please refer to the
[GCP Registry documentation](https://cloud.google.com/container-registry/docs/overview) to decide the actual
URL.

Export the registry URL as `OCI_REGISTRY`

```shell
export OCI_REGISTRY=gcr.io/my-project-1527816345739/tkg/management
```

### Build and Ship Packages to the Registry

Then build and publish all packages according to
the [development workflow](https://github.com/vmware-tanzu/tanzu-framework/blob/main/docs/packages/dev-workflow.md).

```shell
make build-cli && make docker-build && make docker-publish && make kbld-image-replace
```

* `build-cli` compiles the tanzu CLI and its plugins
* `docker-build` builds container images for all packages
* `docker-publish` uploads the container images you built to the OCI registry. Make sure you have authenticated the
  docker per [here](https://cloud.google.com/container-registry/docs/advanced-authentication#gcloud-helper).
* `local-registry` starts a local registry at `localhost:5001`
* `kbld-image-replace` replaces `packages/*/kbld-config.yaml` with the resolved image paths

Note: in theory `make docker-all` will trigger all `docker-build`, `docker-publish` and `kbld-image-replace`. But
locally for me, it only runs the first target with this version
of [Makefile](https://github.com/vmware-tanzu/tanzu-framework/blob/0f572c2ee0d209f538d33b3c5951a664f4076a82/Makefile#L740)
.

After that, run the local registry. Then, build the package and repo bundles for a management clusters.

```shell
make local-registry && make package-push-bundles-repo PACKAGE_REPOSITORY=management
```

## Link clusters to the package repository and install packages
### Spin up a kind cluster as the management cluster

Create kind cluster with CAPx providers and kapp-controller installed. Since this utility script install multiple
infrastructure providers,
environment variables should be exported such as `AWS_B64ENCODED_CREDENTIALS`, `VSPHERE_PASSWORD`, `VSPHERE_USERNAME`.

```shell
export AWS_B64ENCODED_CREDENTIALS=<encoded aws credentials>
./hack/kind/deploy_kind_with_capi_and_kapp.sh
```

Login to the kind cluster with tanzu CLI.

```shell
tanzu login --kubeconfig '${HOME}/.kube/config' --context kind-test-cluster --name kind-test-cluster
```

### Connect the kind cluster to the package repository

Add the `management` repository with packages we just built to the management cluster,

```shell
export PACKAGE_REPO_URL=gcr.io/my-project-1527816345739/tkg/management/management@sha256:0bd76eb7c622fa32cbb851b30d691dc797105445f5e42ec8dd2dc41b34366fdf
tanzu package repository update management --url ${PACKAGE_REPO_URL} --create
```

Ensure that the `management` repository has been reconciled successfully.

```shell
$ tanzu package repository list
  NAME        REPOSITORY                                                 TAG                                                                      STATUS               DETAILS
  management  gcr.io/my-project-1527816345739/tkg/management/management  sha256:0bd76eb7c622fa32cbb851b30d691dc797105445f5e42ec8dd2dc41b34366fdf  Reconcile succeeded
```

At this point, it is expected to see all packages that are available to install.
```shell
$ kubectl get package -A 
NAMESPACE   NAME                                                      PACKAGEMETADATA NAME                           VERSION      AGE
default     addons-manager.tanzu.vmware.com.0.23.0-dev                addons-manager.tanzu.vmware.com                0.23.0-dev   10s
default     cliplugins.tanzu.vmware.com.0.23.0-dev                    cliplugins.tanzu.vmware.com                    0.23.0-dev   10
...
```

In order to create a cluster with cluster-api-aws, we need to install the `tkg-clusterclass-aws` package first. It is possible to hit
package reconcile error if k8s version is 1.24 or above with a older version of kapp-controller such as v0.31.0 (Troubleshoot 3). Let's grep it first.

```shell
$  k get packages -A | grep tkg-clusterclass-aws
default     tkg-clusterclass-aws.tanzu.vmware.com.0.23.0-dev          tkg-clusterclass-aws.tanzu.vmware.com          0.23.0-dev   1h17m22s
```

Then install the `tkg-clusterclass-aws` package.
```shell
$ tanzu package install tkg-clusterclass-aws --package-name tkg-clusterclass-aws.tanzu.vmware.com --version 0.23.0-dev

```

Verify the package is successfully installed with
```shell
$ k get pkgi -A
NAMESPACE   NAME                      PACKAGE NAME                               PACKAGE VERSION   DESCRIPTION           AGE
default     tkg-clusterclass-aws   tkg-clusterclass-aws.tanzu.vmware.com   0.23.0-dev        Reconcile succeeded   91s
```

You are expected to see the `tkg-aws` clusterclass now.
```shell
$ k get cc
NAME                      AGE
tkg-aws-clusterclass   107s
```
## Create workload clusters with clusterclass

First create the `tkg-system` namespace.
```shell
$ k create ns tkg-system
namespace/tkg-system created
```


Create a manifest for the cluster with `tkg-aws` as the cluster class, something like below.










## Troubleshooting

### 1. Couldn't push the imgpkg bundle

When generating repo bundles with `package-push-bundles-repo` with an M1 chips, it is likely to hit

```shell
cd hack/packages/package-tools && go run main.go repo-bundle generate --repository=management --registry=gcr.io/my-project-1527816345739/tkg/management --package-values-file= --version=v0.23.0-dev --sub-version=
Error: couldn't generate package-values-sha256.yaml: couldn't push the imgpkg bundle:
Usage:
  package-tooling repo-bundle generate [flags]
```

This is because `make tools` doesn't download the correct `imgpkg` binary given `ARCH=arm64` but a `Not found` file. As
workaround,
download the `imgpkg-darwin-amd64` and move it under `hack/tools/bin` manually.

### 2. NewImage validation error

When generating the package bundle with `package-push-bundles-repo`, one can possibly hit the validation error.

```shell
$ cd hack/packages/package-tools && go run main.go package-bundle generate --all --thick --repository=management --version=v0.23.0-dev --sub-version= --registry=gcr.io/my-project-1527816345739/tkg/management
Generating "featuregates" package bundle...
Error: couldn't generate imgpkg lock output file: couldn't generate imgpkg lock output file: couldn't run kbld command to generate imgpkg lock output file: kbld: Error: Validating config/ (kbld.k14s.io/v1alpha1) cluster:
  Validating Overrides[0]:
    Expected NewImage to be non-empty
```

This is because the newImage field is still empty in the package `kbld-config.yml` file,
for
example [here](https://github.com/vmware-tanzu/tanzu-framework/blob/a83c583ea8376137dda1abc0192f0443b04974fb/packages/featuregates/kbld-config.yaml#L5)
.
You need to build and publish package images that include the step `make kbld-image-replace`, following
the [development workflow](https://github.com/vmware-tanzu/tanzu-framework/blob/main/docs/packages/dev-workflow.md). 

### 3. Kapp cannot reconcile packages
When running `tanzu package install`, it is likely to hit the following error when running 1.24 k8s.
```shell
Error: resource reconciliation failed: Preparing kapp: Expected to find one service account token secret, but found none. Reconcile failed: Error (see .status.usefulErrorMessage for details)
Error: exit status 1
```
Root cause is [here](https://github.com/vmware-tanzu/carvel-kapp-controller/issues/687). Use higher version of kapp-controller in the script `deploy_kind_with_capi_and_kapp.sh`.





# Testing packages

## Testing clusterclass packages

- Build management package bundles and package repository and publish it

```text
# Configure the registry where you want to publish the packages and package repo bundles
# You will need push permission to this repository
export OCI_REGISTRY=gcr.io/eminent-nation-87317/tkg/test/repo/management

# Build and publish management repo bundle
make package-push-bundles-repo PACKAGE_REPOSITORY=management
```

As part of the logs you will find a URL to image for package repository. Example: `gcr.io/eminent-nation-87317/tkg/test/repo/management/packages/management/management@sha256:164dd18d9642969f5636126c0e5b67d296d56339d8bd7c0ca49fff0cf8e20ad3`

- Create Kind cluster for testing

```text
# Create kind cluster with CAPx providers and kapp-controller installed
./hack/kind/deploy_kind_with_capi_and_kapp.sh
```

- Use `tanzu login` to login to newly created kind cluster as management cluster

```text
# login to the kind cluster. Ignore the CLIPluginList error as it's not needed to be synced
tanzu login --kubeconfig /Users/anujc/.kube/config --context kind-test-cluster --name kind-test-cluster
```

- Install the `PackageRepository` and `Package` to the kind cluster.

```text
tanzu package repository update management --url <PACKAGE-REPOSITORY-URL-GENERATED-EARLIER> --create
```

After it get Reconciled successfully, new package will be available on the cluster.

```text
~> k get packagerepository -A
NAMESPACE   NAME         AGE    DESCRIPTION
default     management   119m   Reconcile succeeded
~> k get package -A
NAMESPACE   NAME                                               PACKAGEMETADATA NAME                    VERSION      AGE
default     clusterclass-aws.tanzu.vmware.com.0.16.0-dev       clusterclass-aws.tanzu.vmware.com       0.16.0-dev   1h58m50s
default     clusterclass-azure.tanzu.vmware.com.0.16.0-dev     clusterclass-azure.tanzu.vmware.com     0.16.0-dev   1h58m50s
default     clusterclass-docker.tanzu.vmware.com.0.16.0-dev    clusterclass-docker.tanzu.vmware.com    0.16.0-dev   1h58m50s
default     clusterclass-vsphere.tanzu.vmware.com.0.16.0-dev   clusterclass-vsphere.tanzu.vmware.com   0.16.0-dev   1h58m50s
```

Install the package for testing with `tanzu package install` command and passing `values-file` for the package. Sample [values file for aws clusterclass](/packages/clusterclass-aws/test-data/sample-values.yaml).

```text
tanzu package install clusterclass-aws-pkg --package-name clusterclass-aws.tanzu.vmware.com --version 0.16.0-dev --values-file sample-values.yaml
```

After it get Reconciled successfully, clusterclass resource should get created.

```text
~> k get cc -A
NAMESPACE   NAME                      AGE
default     tkg-aws-clusterclass      96m
```

Create a `cluster` resource with required corresponding clusterclass and variable defined under topology and use kubectl apply to test the cluster creation. Sample [cluster file for aws](/packages/clusterclass-aws/test-data/cluster.yaml).

```text
kubectl apply -f cluster.yaml
```

## Deleting installed packages and package repository to reset the kind cluster

Below commands should clean kind cluster and user can continue to rebuild packagerepository after fixing the issues and reinstall package repository and packages as mentioned step-4 above.

```text
kubectl delete packageinstall clusterclass-aws-pkg
kubectl delete packagerepository management
```

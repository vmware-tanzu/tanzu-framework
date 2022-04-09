# Dev workflow

This document provides guidance on how to build and publish package and repo bundles and test them on your local cluster.

1. Build CLI plugins

   CLI plugins need to be built by running the below make target, because they are used to generate
   `core-management-plugins` and `standalone-plugins` package bundle. There is WIP to remove this step and build each
   package bundle individually.

   ```shell
      make build-cli
   ```

2. Build the container images for all the packages

   ```shell
      OCI_REGISTRY=${REGISTRY} make docker-all
   ```

   This would build, push a docker image for the component and also replaces the newImage path in the [kbld-config.yaml](../../packages/kbld-config.yaml)

3. Start local registry

   Local docker registry is used for pushing the package bundle to get the sha256, to use it later when producing repo bundle.

   Run the below make target to start the local registry

   ```shell
      make local-registry
   ```

4. Build the package bundles that belong to a particular package repository

   To build the package bundles for all the packages in the management package repository run:

   ```shell
      PACKAGE_REPOSITORY=management make push-package-bundles
   ```

   To build a particular package bundle, run the following command.
   Note that included in the package bundle will be a thick tarball that is useful for air-gapped environments.

   ```shell
      PACKAGE_NAME=my-package make package-bundle
   ```

   For environments where an external image registry is accessible, you may use the following command to build a package bundle:

   ```shell
      PACKAGE_NAME=my-package make package-bundle-thin
   ```

5. Build package repo bundle

   After building the package bundles, it is time to build package repository bundle, run the below make
   target:

   ```shell
      PACKAGE_REPOSITORY=management OCI_REGISTRY=${REGISTRY} make package-repo-bundle
   ```

6. Push package bundles

   After the package bundles are generated, now it's time to push them to an OCI registry, so that they can be consumed.
   Run the below make target to push all the package bundles in the specified package repository:

   ```shell
      PACKAGE_REPOSITORY=management OCI_REGISTRY=${REGISTRY} make push-all-package-bundles
   ```

   If you are interested in pushing only a few package bundles, you could do that by running

   ```shell
      OCI_REGISTRY=${REGISTRY} make push-package-bundles ${PACKAGE_BUNDLES}
   ```

   PACKAGE_BUNDLES variable should be comma-separated values and must not contain spaces.
   Example: PACKAGE_BUNDLES=featuregates,core-management-plugins

7. Push repo bundle

   To push the generated repo bundle to an OCI registry, run:

   ```shell
      PACKAGE_REPOSITORY=management OCI_REGISTRY=${REGISTRY} make push-package-repo-bundle
   ```

Follow the below steps to test the artifacts that are generated in previous steps on your local cluster

**Note**: A local cluster can be any k8s cluster with `kapp-controller` installed in it.
You can also use [management cluster](https://github.com/vmware-tanzu/tanzu-framework/blob/main/cmd/cli/plugin/managementcluster/README.md)
or [workload cluster](https://github.com/vmware-tanzu/tanzu-framework/blob/main/cmd/cli/plugin/cluster/README.md) which
comes with `kapp-controller` already installed in it to test the artifacts.

We can use the `package` plugin to add the repo bundle to the cluster and to install the packages

1. Add package repo bundle to the cluster

   ```shell
      tanzu package repository add repo --url ${REPO_BUNDLE_URL} --namespace ${NAMESPACE}
   ```

2. Install the package

   ```shell
      tanzu package install tkg --namespace ${NAMESPACE} --package-name ${PACKAGE_NAMA} --version ${PACKAGE_VERSION}
   ```

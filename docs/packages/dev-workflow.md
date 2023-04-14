# Dev workflow

This document provides guidance on how to build and publish package and repo bundles and test them on your local cluster.

> For detailed info about variables and targets used here, see [build-tooling](https://github.com/vmware-tanzu/build-tooling-for-integrations/blob/main/docs/build-tooling-getting-started.md).

1. Build the container images for all the packages

   Images need to be built by running the below make target, because they are used to generate
   package bundles.

   ```shell
      OCI_REGISTRY="****" make docker-build-all -f build-tooling.mk
   ```

2. Publish the container images for all the packages

   ```shell
      OCI_REGISTRY="****" make docker-publish-all -f build-tooling.mk
   ```
   This would push contianer images for the components to specified registry and also replaces the newImage path in the [kbld-config.yaml](../../packages/readiness/kbld-config.yaml)

3. Build the package bundles that belong to a particular package repository

   To build the package bundles for all the packages in the `management` package [repository](../../packages/package-values.yaml) run:

   ```shell
      OCI_REGISTRY="****" \
      PACKAGE_REPOSITORY="management" \
      PACKAGE_VERSION="dev.0.0" \
      PACKAGE_SUB_VERSION="1" \
      REGISTRY_USERNAME="****" \
      REGISTRY_PASSWORD="****" \
      REGISTRY_SERVER="****" \
      make package-bundle-generate-all -f build-tooling.mk
   ```

   To build a particular package bundle, run the following command. `PACKAGE_NAME` should specify the directory name under [packages](../../packages)
   Note that included in the package bundle will be a thick tarball that is useful for air-gapped environments.

   ```shell
      OCI_REGISTRY="****" \
      PACKAGE_NAME="readiness" \
      PACKAGE_REPOSITORY="management" \
      PACKAGE_VERSION="dev.0.0" \
      PACKAGE_SUB_VERSION="1" \
      REGISTRY_USERNAME="****" \
      REGISTRY_PASSWORD="****" \
      REGISTRY_SERVER="****" \
      make package-bundle-generate -f build-tooling.mk
   ```

4. Build package repo bundle [Optional]

   After building the package bundles, it is time to build package repository bundle, run the below make
   target:

   ```shell
      PACKAGE_REPOSITORY=management OCI_REGISTRY="****" make package-repo-bundle
   ```

5. Push package bundles 

   After the package bundles are generated, now it's time to push them to an OCI registry, so that they can be consumed.
   Run the below make target to push all the package bundles in the specified package repository:

   ```shell
      OCI_REGISTRY="****" \
      PACKAGE_REPOSITORY="management" \
      PACKAGE_VERSION="dev.0.0" \
      PACKAGE_SUB_VERSION="1" \
      REGISTRY_USERNAME="****" \
      REGISTRY_PASSWORD="****" \
      REGISTRY_SERVER="****" \
      make package-bundle-push-all -f build-tooling.mk
   ```

   If you are interested in pushing only a specific package bundle, you could do that by running

   ```shell
      OCI_REGISTRY="****" \
      PACKAGE_NAME="readiness" \
      PACKAGE_REPOSITORY="management" \
      PACKAGE_VERSION="dev.0.0" \
      PACKAGE_SUB_VERSION="1" \
      REGISTRY_USERNAME="****" \
      REGISTRY_PASSWORD="****" \
      REGISTRY_SERVER="****" \
      make package-bundle-push -f build-tooling.mk
   ```

6. Push repo bundle [Optional]

   To push the generated repo bundle to an OCI registry, run:

   ```shell
      PACKAGE_REPOSITORY=management OCI_REGISTRY="****" make repo-bundle-push -f build-tooling.mk
   ```

----

Follow the below steps to test the artifacts that are generated in previous steps on your local cluster

> **Note**: A local cluster can be any k8s cluster. There are 2 ways to install the packages

#### To install on a cluster without `kapp-controller` installed in it.

1. Download the package bundle. The sha tag can be noted from the registry where package was published.
   ```shell
      imgpkg pull -b ${OCI_REGISTRY}/readiness@sha256:1fb9f9c6f0c6ba1f995440885b02806551a79d9cef5b9c7c3d6f53a586facddd -o readiness-pkg
   ```

1. Install the package
   ```shell
      ytt -f readiness-pkg/config/ | kbld -f - -f readiness-pkg/.imgpkg/images.yml | kubectl apply -f-
   ```

#### To install on a cluster with `kapp-controller` installed in it.
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

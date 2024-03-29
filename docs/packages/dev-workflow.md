# Dev workflow

This document provides guidance on how to build and publish package and repo bundles and test them on your local cluster.

> For detailed info about variables and targets used here, see [build-tooling](https://github.com/vmware-tanzu/build-tooling-for-integrations/blob/main/docs/build-tooling-getting-started.md).

1. Build the container images for all the packages

   Images need to be built by running the below make target, because they are used to generate
   package bundles.

   ```shell
      OCI_REGISTRY="****" make docker-build-all
   ```

2. Publish the container images for all the packages

   ```shell
      OCI_REGISTRY="****" make docker-publish-all
   ```

   This would push container images for the components to specified registry and also replaces the newImage path in the [kbld-config.yaml](../../packages/readiness/kbld-config.yaml)

3. Build the package bundles that belong to a particular package repository

   To build the package bundles for all the packages in the `runtime-core` package [repository](../../packages/package-values.yaml) run:

   ```shell
      OCI_REGISTRY="****" \
      PACKAGE_REPOSITORY="runtime-core" \
      PACKAGE_VERSION="dev.0.0" \
      PACKAGE_SUB_VERSION="1" \
      REGISTRY_USERNAME="****" \
      REGISTRY_PASSWORD="****" \
      REGISTRY_SERVER="****" \
      make package-bundle-generate-all
   ```

   To build a particular package bundle, run the following command. `PACKAGE_NAME` should specify the directory name under [packages](../../packages)
   Note that included in the package bundle will be a thick tarball that is useful for air-gapped environments.

   ```shell
      OCI_REGISTRY="****" \
      PACKAGE_NAME="readiness" \
      PACKAGE_REPOSITORY="runtime-core" \
      PACKAGE_VERSION="dev.0.0" \
      PACKAGE_SUB_VERSION="1" \
      REGISTRY_USERNAME="****" \
      REGISTRY_PASSWORD="****" \
      REGISTRY_SERVER="****" \
      make package-bundle-generate
   ```

4. Build package repo bundle [Optional]

   After building the package bundles, it is time to build package repository bundle, run the below make
   target:

   ```shell
      PACKAGE_REPOSITORY="runtime-core" \
      OCI_REGISTRY="****" \
      REPO_BUNDLE_VERSION="v1.0.0" \
      REPO_BUNDLE_SUB_VERSION="0" \
      PACKAGE_VALUES_FILE=packages/package-values.yaml \
      make repo-bundle-generate
   ```

5. Push package bundles

   After the package bundles are generated, now it's time to push them to an OCI registry, so that they can be consumed.
   Run the below make target to push all the package bundles in the specified package repository:

   ```shell
      OCI_REGISTRY="****" \
      PACKAGE_REPOSITORY="runtime-core" \
      PACKAGE_VERSION="dev.0.0" \
      PACKAGE_SUB_VERSION="1" \
      REGISTRY_USERNAME="****" \
      REGISTRY_PASSWORD="****" \
      REGISTRY_SERVER="****" \
      make package-bundle-push-all
   ```

   If you are interested in pushing only a specific package bundle, you could do that by running

   ```shell
      OCI_REGISTRY="****" \
      PACKAGE_NAME="readiness" \
      PACKAGE_REPOSITORY="runtime-core" \
      PACKAGE_VERSION="dev.0.0" \
      PACKAGE_SUB_VERSION="1" \
      REGISTRY_USERNAME="****" \
      REGISTRY_PASSWORD="****" \
      REGISTRY_SERVER="****" \
      make package-bundle-push
   ```

6. Push repo bundle [Optional]

   To push the generated repo bundle to an OCI registry, run:

   ```shell
      PACKAGE_REPOSITORY="runtime-core" \
      OCI_REGISTRY="****" \
      REPO_BUNDLE_VERSION="v1.0.0" \
      REPO_BUNDLE_SUB_VERSION="0" \
      REGISTRY_USERNAME="****" \
      REGISTRY_PASSWORD="****" \
      REGISTRY_SERVER="****" \
      make repo-bundle-generate repo-bundle-push
   ```

----

Follow the below steps to test the artifacts that are generated in previous steps on your local cluster

> **Note**: A local cluster can be any k8s cluster. There are 2 ways to install the packages

**To install on a cluster without `kapp-controller` installed in it**

1. Download the package bundle. The image sha tag can be noted from the registry where package was published.

   ```shell
      imgpkg pull -b ${OCI_REGISTRY}/readiness@sha256:1fb9f9c6f0c6ba1f995440885b02806551a79d9cef5b9c7c3d6f53a586facddd -o readiness-pkg
   ```

2. Install the package

   ```shell
      ytt -f readiness-pkg/config/ | kbld -f - -f readiness-pkg/.imgpkg/images.yml | kubectl apply -f-
   ```

**To install on a cluster with `kapp-controller` installed in it**

> For more details about these steps, see kapp [packaging tutorial](https://carvel.dev/kapp-controller/docs/v0.31.0/packaging-tutorial)

1. Create a `PackageRepository` resource. The image sha tag can be noted from the registry where **repo bundle** was published.

   Create `repo.yaml`

   ```yaml
      ---
      apiVersion: packaging.carvel.dev/v1alpha1
      kind: PackageRepository
      metadata:
      name: tanzu-runtime-core-repo
      spec:
      fetch:
         imgpkgBundle:
            image: ${OCI_REGISTRY}/runtime-core@sha256:1fb9f9c6f0c6ba1f995440885b02806551a79d9cef5b9c7c3d6f53a586facddd
   ```

   ```bash
      kapp deploy -a repo -f repo.yaml -y
   ```

2. Install kapp-controller service account [Optional if already present]

   ```bash
      kapp deploy -a default-ns-rbac -f https://raw.githubusercontent.com/vmware-tanzu/carvel-kapp-controller/develop/examples/rbac/default-ns.yml -y
   ```

   This will create a service account named `default-ns-sa`.

3. Install the package

   Create `pkginstall.yaml`

   ```yaml
      ---
      apiVersion: packaging.carvel.dev/v1alpha1
      kind: PackageInstall
      metadata:
      name: tanzu-readiness-pkg
      spec:
      serviceAccountName: default-ns-sa # service account from step 2
      packageRef:
         refName: readiness.tanzu.vmware.com
   ```

   ```shell
      kapp deploy -a tanzu-readiness-package -f pkginstall.yaml -y
   ```

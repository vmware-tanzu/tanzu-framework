# Add PackageInstall CR to a Meta Package

This document provides guidance on how to add a PackageInstall CR to a meta package.
Check this [doc](./definitions.md#meta-package) to understand some packaging terminology such as
meta package in the context of this project.

`tkg`, `framework` are some examples of meta package in this repository.

## Steps to add PackageInstall CR to a Meta Package

Let's take the example of adding `featuregates` PackageInstall to `framework`.

1. Add PackageInstall CR to the meta package config

   To install an actual package and its underlying resources on a Kubernetes cluster, a PackageInstall CR must be
   created on the Kubernetes cluster. So, for the package to be installed when the meta package is installed,
   we need to add PackageInstall CR in the meta package config.

   More details on PackageInstall CR can be found [here](https://carvel.dev/kapp-controller/docs/v0.34.0/packaging/#package-install)

2. Add ServiceAccount needed to install the package through meta package

   Every PackageInstall CR must provide ServiceAccount name in the serviceAccountName field for kapp-controller to
   install the underlying package resources in the Kubernetes cluster. This ServiceAccount provided should have access
   to needed privileges for management of underlying package resources.

   More info about the kapp-controller security model can be found [here](https://carvel.dev/kapp-controller/docs/v0.34.0/security-model/)

   Ex: Directory structure after adding featuregates PackageInstall CR, ServiceAccount and RBAC rules to
   framework package

   ```plain
   packages/framework
   ├── Makefile
   ├── README.md
   ├── bundle
   │ └── config
   │     ├── pacakgeinstalls
   │     │ ├── featuregates-pi.yaml
   │     │ └── featuregates-service-account.yaml
   │     └── values.yaml
   ├── metadata.yaml
   └── package.yaml
   ```

3. Repeat steps 1 and 2 if the meta package is part of another meta package

Example: This [PR](https://github.com/vmware-tanzu/tanzu-framework/pull/1848) that adds featuregates PackageInstall to
the `framework` meta package and also adds `framework` PackageInstall to `tkg`
meta package.

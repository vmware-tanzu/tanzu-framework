# Definitions

This document provides definitions of some package terminology used in the Tanzu framework.
To learn about the Carvel package concepts visit this [page](https://carvel.dev/kapp-controller/docs/latest/packaging)

## Management package

Any package that performs the role of managing and operating a workload cluster is a Management package.
Management packages are purpose-built to be installed on the management cluster.

Example: Featuregates, addons-manager, etc.

## Management package repository

A management package repository is a repository for packages that are exclusively intended to be installed on a
management cluster.

## Meta package

A Package that contains PackageInstall CRs, Secrets and ServiceAccounts, RBAC rules that are needed to install other
packages and when the meta package is installed, it installs the other packages. A meta Package may describe
a specific version of a Distribution.

## Thick tarball

A tarball that contains a package bundle and its associated images.
Thick tarballs are suitable for air-gapped environments.

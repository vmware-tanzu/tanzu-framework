# Deploying the local changes to a Kind cluster

## Pre-requisites

1. Docker
2. Kind
3. ytt
4. Kubectl
5. Make

### Create Kind cluster

The kind cluster can be created by the following command.

```bash
KIND_CLUSTER_NAME=<cluster_name> KUBE_VERSION=<kubernetes_version> make create-kind-cluster
```

For example, the following command creates a kind cluster with name `kind1` and k8s version `v1.26.3`.

```bash
KIND_CLUSTER_NAME=kind1 KUBE_VERSION=v1.26.3 make create-kind-cluster
```

## Deploy local changes to a kind cluster

Run the following command to install the readiness framework in the Kind cluster created.

```bash
KIND_CLUSTER_NAME=<cluster_name> make deploy-local-readiness
```

Example:

```bash
KIND_CLUSTER_NAME=kind1 make deploy-local-readiness
```

## Run end-to-end tests

The end-to-end tests can be triggered on the local Kind cluster by running the following command.

```bash
make e2e-readiness
```

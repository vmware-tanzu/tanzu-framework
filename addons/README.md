# Tanzu Addons Manager

Tanzu Addons Manager manages the lifecycle of core addons like CNI, CPI, CSI, etc. It utilizes Kapp-controller's [packaging API](https://carvel.dev/kapp-controller/docs/latest/packaging/) and [App CR](https://carvel.dev/kapp-controller/docs/latest/app-spec/) to do the core addons lifecycle management.

## Watch

Tanzu Addons Manager watches the following

- Addons secret

- Cluster CR

- Kubeadm control plane

- BOM configmap

## Workflow of Tanzu Addons Manager

1. receives a request

2. reconciles the core package repository according to the TKR BOM configmap of the cluster

3. reconciles all the addons secrets
    - If it's a remote app (App CR that lives in one cluster but deploys resources in another cluster)
      - Creates/updates addon data values secret on mgmt cluster
      - Get the remote cluster kubeconfig
      - Create/updates remote App CR in mgmt cluster that deploys resources in a remote cluster
    - If it's not a remote app
      - Creates/updates addon data values secret on the cluster that the addons secret points to
      - Creates/updates packageInstall CR on the cluster that the addons secret points to

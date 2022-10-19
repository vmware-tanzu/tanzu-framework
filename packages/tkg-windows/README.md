# tkg-windows Package

The tkg-windows package provides the following services:

* RBAC objects for Kube-Proxy and Antrea CNI installation on Windows.

## Configuration

The following configuration values can be set to customize the tkg-windows installation.

### Global

| Value | Required/Optional | Description |
|-------|-------------------|-------------|
| `namespace` | Optional | Namespace the resources will be created (default: kube-system) |

## Components

* Role: node:antrea-read-secret
* RoleBinding: node:read-antrea-sa
* ClusterRole: node: kube-proxy
* ClusterRoleBinding: node:kube-proxy
* ServiceAccount: kube-proxy-windows

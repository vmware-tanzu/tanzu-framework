# cluster-api-provider-vsphere Package

This package provides the Cluster API implementation for vSphere using
[cluster-api-provider-vsphere](https://github.com/kubernetes-sigs/cluster-api-provider-vsphere).

## Components

* capv-controller-manager

## Configuration

The following configuration values can be set to customize the installation.

| Value                              | Required/Optional | Description                                                                 |
|------------------------------------|-------------------|-----------------------------------------------------------------------------|
| `capvControllerManager.httpProxy`  | Optional          | Configures the HTTP_PROXY environment variable on capv-controller-manager.  |
| `capvControllerManager.httpsProxy` | Optional          | Configures the HTTPS_PROXY environment variable on capv-controller-manager. |
| `capvControllerManager.noProxy`    | Optional          | Configures the NO_PROXY environment variable on capv-controller-manager.    |

## Usage Example

To learn more about cluster-api-provider-vsphere visit
<https://github.com/kubernetes-sigs/cluster-api-provider-vsphere>.

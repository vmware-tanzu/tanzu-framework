# cluster-api-provider-azure Package

This package provides the Cluster API implementation for Microsoft Azure using
[cluster-api-provider-azure](https://github.com/kubernetes-sigs/cluster-api-provider-azure).

## Components

* capz-controller-manager

## Configuration

The following configuration values can be set to customize the installation.

| Value                              | Required/Optional | Description                                                                 |
|------------------------------------|-------------------|-----------------------------------------------------------------------------|
| `capzControllerManager.httpProxy`  | Optional          | Configures the HTTP_PROXY environment variable on capz-controller-manager.  |
| `capzControllerManager.httpsProxy` | Optional          | Configures the HTTPS_PROXY environment variable on capz-controller-manager. |
| `capzControllerManager.noProxy`    | Optional          | Configures the NO_PROXY environment variable on capz-controller-manager.    |

## Usage Example

To learn more about cluster-api-provider-azure visit
<https://github.com/kubernetes-sigs/cluster-api-provider-azure>.

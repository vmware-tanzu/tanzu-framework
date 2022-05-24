# cluster-api-provider-aws Package

This package provides consistent deployment and day 2 operations of
"self-managed" and EKS Kubernetes clusters on AWS using
[cluster-api-provider-aws](https://github.com/kubernetes-sigs/cluster-api-provider-aws).

## Components

* capa-controller-manager

## Configuration

The following configuration values can be set to customize the installation.

| Value                              | Required/Optional | Description                                                                 |
|------------------------------------|-------------------|-----------------------------------------------------------------------------|
| `capaControllerManager.httpProxy`  | Optional          | Configures the HTTP_PROXY environment variable on capa-controller-manager.  |
| `capaControllerManager.httpsProxy` | Optional          | Configures the HTTPS_PROXY environment variable on capa-controller-manager. |
| `capaControllerManager.noProxy`    | Optional          | Configures the NO_PROXY environment variable on capa-controller-manager.    |

## Usage Example

To learn more about cluster-api-provider-aws visit
<https://github.com/kubernetes-sigs/cluster-api-provider-aws>.

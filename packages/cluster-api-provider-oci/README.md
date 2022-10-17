# cluster-api-provider-aws Package

This package provides consistent deployment and day 2 operations of
"self-managed" Kubernetes clusters on OCI using
[cluster-api-provider-oci](https://github.com/oracle/cluster-api-provider-oci).

## Components

* capoci-controller-manager

## Configuration

The following configuration values can be set to customize the installation.

| Value                                | Required/Optional | Description                                                                  |
|--------------------------------------|-------------------|------------------------------------------------------------------------------|
| `capociControllerManager.httpProxy`  | Optional          | Configures the HTTP_PROXY environment variable on capoci-controller-manager. |
| `capociControllerManager.httpsProxy` | Optional          | Configures the HTTPS_PROXY environment variable on capoci-controller-manager.  |
| `capociControllerManager.noProxy`    | Optional          | Configures the NO_PROXY environment variable on capoci-controller-manager.     |

## Usage Example

To learn more about cluster-api-provider-oci visit
<https://github.com/oracle/cluster-api-provider-oci>.

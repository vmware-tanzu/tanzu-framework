# readiness Package

This package provides the functionality to define and evaluate k8s clusters for suitability of running enterprise workloads using [readiness](https://INFO_NEEDED).

## Components

## Configuration

The following configuration values can be set to customize the readiness installation.

### Global

| Value | Required/Optional | Description |
|-------|-------------------|-------------|
| `namespace` | Optional | Target deployment namespace. Defaults to `default` namespace |

### readiness Configuration

| Value | Required/Optional | Description |
|-------|-------------------|-------------|
| `deployment.hostNetwork` | Optional | If true, use the host's network namespace. Defaults to `false` |
| `deployment.nodeSelector` | Optional | node selectors for deployment of controller-manager pods |
| `deployment.tolerations` | Optional | tolerations for deployment of controller-manager pods. Defaults to `NoSchedule` on `control-plane` nodes |
| `deployment.webhookServerPort` | Optional | The port that the webhook server serves at |
| `deployment.tlsCipherSuites` | Optional | Comma-separated list of cipher suites for the server. If omitted, the default Go cipher suites will be used. |

## Usage Example

Readiness and ReadinessProvider API concepts and usage guide can be found [here](../../docs/runtime-core/readiness-framework.md).

# Object Propagation Controller

This controller source objects into the target namespaces.

Each set of source objects is specified using:

- apiVersion - required
- kind - required
- namespace - required
- label selector - may be empty ("")

The target namespace is specified by a label selector, which may be empty ("").

The controller reads configuration provided via `--input` CLI parameter (default: `/dev/stdin`).
Example input:

```yaml
- source:
    apiVersion: v1
    kind: ConfigMap
    namespace: tanzu-system
    labelSelector: 'run.tanzu.vmware.com/propagated'
  target:
    namespaceLabelSelector: '!cluster.x-k8s.io/provider'
- source:
    apiVersion: v1
    kind: Secret
    namespace: tanzu-system
    labelSelector: 'run.tanzu.vmware.com/propagated'
  target:
    namespaceLabelSelector: '!cluster.x-k8s.io/provider'
```

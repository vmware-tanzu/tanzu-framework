vSphere CPI

The base files are obtained from https://github.com/kubernetes-sigs/cluster-api-provider-vsphere by running and extracting CPI
manifests from the result.

```
go run ./packaging/flavorgen -f vip > cluster-template.yaml
```
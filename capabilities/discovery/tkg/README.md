# TKG Discovery

This repository is a collection of constants and functions which expose Discovery queries aimed at making it simple to
extend and integrate with TKG.

Initialize a new TKG `DiscoveryClient` using `rest.Config`. Then, you can run any existing queries:

```go
tkg, err := NewDiscoveryClientForConfig(cfg)
if err != nil {
    log.Fatal(err)
}

if tkg.IsManagementCluster() {
    log.Info("Management cluster")
}
```

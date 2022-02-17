# TKR Service

This is the top level directory for the TKR Service.

## TKR Resolver Library

Usage:

```go
import (
"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
)

// skip ...

tkrResolver := resolver.New()
tkrResolver.Add( /* tkr1, tkr2, tkr3, osImage1, osImage2, osImage3 */)

// skip ...
k8sVersionPrefix := "1.22"
tkrSelector, _ := labels.Parse("!deprecated")
osImageSelector, _ := labels.Parse("os-name=ubuntu,ami-region=us-west-2")

query := data.Query{
ControlPlane: data.OSImageQuery{
K8sVersionPrefix: k8sVersionPrefix,
TKRSelector:      tkrSelector,
OSImageSelector:  osImageSelector,
},
MachineDeployments: map[string]data.OSImageQuery{
"md1": {
K8sVersionPrefix: k8sVersionPrefix,
TKRSelector:      tkrSelector,
OSImageSelector:  osImageSelector,
},
},
}

result := tkrResolver.Resolve(query)
```

The primary client of the Resolver package will be the TKR Resolver webhook on CAPI Cluster objects. This webhook will
have two parts: the cache reconciler and the webhook handler.

The cache reconciler will list all TKRs and OSImages on initialization and populate the cache. It would then watch TKRs
and OSImages and simply Add() or Remove() on incoming updates, thereby providing background cache refresh and
invalidation.

This will make the webhook handler efficient by avoiding unnecessary work (fetching data and populating indexes) on
every webhook request.

The provided CachingResolver implementation is safe for concurrent use.

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
query := data.Query{
  ControlPlane: data.OSImageQuery{
    K8sVersionPrefix: "1.22",
    TKRSelector:      tkrSelector,
    OSImageSelector:  osImageSelector,
  },
  MachineDeployments: map[string]data.OSImageQuery{
    "nodePool1": {
      K8sVersionPrefix: "1.22",
      TKRSelector:      tkrSelector,
      OSImageSelector:  osImageSelector,
    },
  },
}

result := tkrResolver.Resolve(query)
```

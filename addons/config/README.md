# Addon Config CR templates

This folder stores YTT template for generating default addon config CRs used by ClusterBootstrapTemplate. It will be invoked by the make targed at the root of tanzu-framework.

The make target needs 4 inputs, which are GVK of the config CR, and the current TKR version.

``` bash
make generate-package-config apiGroup=cni.tanzu.vmware.com kind=AntreaConfig version=v1alpha1 tkr=v1.23.3---vmware.1-tkg.1
```

## Folder structure

``` text
.
├── README.md
├── {Group}
│   └── {Version}
│       ├── {Kind}.yaml
│       └── {Kind}.yaml

```

## Templating

Typical templating needed would be `Name` and `Namespace`

``` yaml
metadata:
  name: #@ data.values.TKR_VERSION
  namespace: #@ data.values.GLOBAL_NAMESPACE
```

Other values should be included if the value should be different than the default value set for the config CR. You can refer to the kubebuilder definitions in `apis/` to find the default values.

# Addon Config CR templates

This folder stores YTT template for generating default addon config CRs used by ClusterBootstrapTemplate. It will be invoked by the make targed at the root of tanzu-framework.

The make target takes 4 required inputs, which are group, version and kind of the config CR, and the current TKR version. The global namespace is taken as an optional input, defaulted to `tkg-system` if not provided.

``` bash
make generate-package-config apiGroup=cni.tanzu.vmware.com kind=AntreaConfig version=v1alpha1 tkr=v1.23.3---vmware.1-tkg.1 namespace=tkg-system
```

## Folder structure

``` text
| templates
├── {Group1}
│   └── {Version1}
│       ├── {Kind1}.yaml
│       └── {Kind2}.yaml
│   └── {Version2}
│       ├── {Kind1}.yaml
│       └── {Kind2}.yaml
├── {Group2}
│   └── {Version1}
│       ├── {Kind1}.yaml
│       └── {Kind2}.yaml
......
```

## Templates

Typical templating needed would be `Name` and `Namespace`

``` yaml
metadata:
  name: #@ data.values.TKR_VERSION
  namespace: #@ data.values.GLOBAL_NAMESPACE
```

Other values are included if the value should be different from the default value set for the config CR. You can refer to the kubebuilder definitions in `apis/` to find the default values.

## Expected

Each template should have it's corresponding expected output.

For all the expected outputs, `GLOBAL_NAMESPACE=tkg-system` and `TKR_VERSION=v1.23.3---vmware.1-tkg.1` are used as sample input

## Managed component
Date 07/30/2021: have converted Kind from a unmanaged component to a managed component, including changes:
- created cayman base branch:  https://gitlab.eng.vmware.com/core-build/cayman_kubernetes-sigs_kind/-/commits/v0.8.1+vmware.0/kubernetes-sigs_kind/
- added configurables as dependencies in kubernetes-sigs_kind config file. Refer: kubernetes-sigs_kind/v1.21.2+vmware.1_v0.8.1.yaml
- put k8s version and kind version in the component name.
- added edcd, coreDNS, kubernetes as child node of Kind, which is required in Kind image and needed to be build as prerequisite.

## How to release Kind
### To use Kind version 0.8.1
1. create new Kind component config file based on file <previous K8s version>_v0.8.1/yaml, ex. kubernetes-sigs_kind/v1.21.2+vmware.1_v0.8.1.yaml
2. Bump
    - Kind version to <current k8s version>_v0.8.1
    - Component references' versions.

### To use Kind version > 0.8.1
Create new cayman base branch for Kind in https://gitlab.eng.vmware.com/core-build/cayman_kubernetes-sigs_kind, refer branch v0.8.1+vmware.0
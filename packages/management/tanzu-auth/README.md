# Tanzu Auth Package

Tanzu Auth management package provides federated user authentication to a fleet of clusters via [pinniped](https://pinniped.dev/).

## Components

* tanzu-auth-controller
* tanzu-auth-cascade-controller

## Usage Example

Build the tanzu-auth-controller-manager and deploy the image to a registry

```bash
# take note of the image output, this value will be needed in a later step
cd path/to/tanzu-framework/addons/pinniped/config-controller 
./hack/run.sh
# ...
# The push refers to repository [harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager]
# 56d0b8a656ce: Layer already exists
# 0b031aac6569: Layer already exists
# dev: digest: sha256:e65505c2e3bab863436ca24ace6c25dc029f32f4b349804fe9e462c2ba2f09e3 size: 739
#T arget cluster 'https://10.206.80.174:6443' (nodes: tkg-mgmt-vc-control-plane-xmzch, 1+)
# resolve | final: harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager:dev -> harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager@sha256:e65505c2e3bab863436ca24ace6c25dc029f32f4b349804fe9e462c2ba2f09e3
```

To simply validate the `ytt` templates:

```bash 
# ytt -f ./bundle/config -v image=<some.image>
ytt -f ./bundle/config -v image=harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager:dev
```

Use `kbld` to resolve template images to a sha:

```bash 
ytt -f ./bundle/config -v image=harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager:dev | kbld -f -
# image reference will be replaced with:
# - image: harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager@sha256:e65505c2e3bab863436ca24ace6c25dc029f32f4b349804fe9e462c2ba2f09e3
```

Generate a new `imgpkg` file with `kbld`:

```bash
ytt -f ./bundle/config -v image=harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager:dev | kbld -f - --imgpkg-lock-output bundle/.imgpkg/images.yml
```

Deploy the app as `tanzu-auth` with `kapp`:

```bash 
ytt -f ./bundle/config -v image=harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager:dev | kbld -f - --imgpkg-lock-output bundle/.imgpkg/images.yml | kapp deploy --app tanzu-auth -f- 
```

Investigate the deployed application:

```bash
kapp inspect --app tanzu-auth --tree
# Target cluster 'https://10.206.80.174:6443' (nodes: tkg-mgmt-vc-control-plane-xmzch, 1+)
#
# Resources in app 'tanzu-auth'
#
# Namespace   Name                                                 Kind                Owner    Conds.  Rs  Ri  Age
# tanzu-auth  tanzu-auth-controller-manager                        Deployment          kapp     2/2 t   ok  -   7m
# tanzu-auth   L tanzu-auth-controller-manager-557b9df496          ReplicaSet          cluster  -       ok  -   7m
# tanzu-auth   L.. tanzu-auth-controller-manager-557b9df496-nzcgt  Pod                 cluster  4/4 t   ok  -   7m
# tanzu-auth   L tanzu-auth-controller-manager-557b9df496-nzcgt    PodMetrics          cluster  -       ok  -   2s
# tanzu-auth  tanzu-auth-controller-manager-sa                     ServiceAccount      kapp     -       ok  -   7m
# (cluster)   tanzu-auth-controller-manager                        ClusterRole         kapp     -       ok  -   7m
# (cluster)   tanzu-auth                                           Namespace           kapp     -       ok  -   7m
# (cluster)   tanzu-auth-controller-manager                        ClusterRoleBinding  kapp     -       ok  -   7m
#
# Rs: Reconcile state
# Ri: Reconcile information
#
# 8 resources
#
# Succeeded
```

And view its logs 

```bash 
kapp logs --app tanzu-auth -f
```

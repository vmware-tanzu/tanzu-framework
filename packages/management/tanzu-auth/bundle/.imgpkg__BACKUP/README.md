# Tanzu Auth

The Tanzu Auth package is responsible for the deployment of the tanzu-auth-controller-manager binary.

## Working Notes

This package is in progress with this section indicating the working state of a manual deployment of this package.  This 
doc assumes you have a functioning TKG cluster and local kubeconfig. 

Ensure you have built and pushed the `tanzu-auth-controller-manager` container to a registry.  You will need a reference
to the image in a later step.

```bash
cd path/to/tanzu-framework/addons/pinniped/tanzu-auth-controller
# note the output of a line like:
# The push refers to repository [harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager]
# resolve | final: harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager:dev -> harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager@sha256:e65505c2e3bab863436ca24ace6c25dc029f32f4b349804fe9e462c2ba2f09e3
./hack/run.sh
```

The rest of this doc will assume `repository=harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager`

Now to build and deploy the `tanzu-auth-controller-manager-package`.

```bash
# begin in this directory
cd path/to/tanzu-framework/packages/management/tanzu-auth
```

To validate the `ytt` templates

```bash
# referencing the image from the registry above
ytt -f ./bundle/config -v image=<tanzu-auth-controller-manager-image>
# specifically
ytt -f ./bundle/config -v image=harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager@sha256:e65505c2e3bab863436ca24ace6c25dc029f32f4b349804fe9e462c2ba2f09e3
```

Then use `ytt` and `kbld` to generate a new `.imgpkg/images.yaml` file

```bash 
# referencing the image from the registry above
ytt -f ./bundle/config -v image=<tanzu-auth-controller-manager-image> | kbld -f - --imgpkg-lock-output bundle/.imgpkg/images.yml
# specifically
ytt -f ./bundle/config \
  -v image=harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager@sha256:e65505c2e3bab863436ca24ace6c25dc029f32f4b349804fe9e462c2ba2f09e3 \
  | kbld -f - --imgpkg-lock-output bundle/.imgpkg/images.yml
```

You can use `imgpkg` to push the bundle up to the registry

```bash
imgpkg push --bundle harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager-package --file bundle
```

And pull it back down to examine the contents

```bash 
imgpkg pull --bundle harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager-package --output /tmp/tanzu-auth-controller-manager-package
```

To fully deploy the package onto the running cluster run:

```bash 
ytt -f ./bundle/config -v image=harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager@sha256:e65505c2e3bab863436ca24ace6c25dc029f32f4b349804fe9e462c2ba2f09e3 | kbld -f - --imgpkg-lock-output bundle/.imgpkg/images.yml | kapp deploy --app tanzu-auth -f-
# broken down to digest
ytt -f ./bundle/config \
    -v image=harbor-repo.vmware.com/tkgiam/ben/tanzu-auth-controller-manager@sha256:e65505c2e3bab863436ca24ace6c25dc029f32f4b349804fe9e462c2ba2f09e3 \
    | kbld -f - --imgpkg-lock-output bundle/.imgpkg/images.yml \
    | kapp deploy --app tanzu-auth -f-
```

Inspecting the `tanzu-auth` App should reveal running resources:

```bash
kapp inspect --app tanzu-auth --tree
# Target cluster 'https://10.206.80.174:6443' (nodes: tkg-mgmt-vc-control-plane-xmzch, 1+)
#
# Resources in app 'tanzu-auth'
#
# Namespace   Name                                                 Kind                Owner    Conds.  Rs  Ri  Age
# tanzu-auth  tanzu-auth-controller-manager                        Deployment          kapp     2/2 t   ok  -   23s
# tanzu-auth   L tanzu-auth-controller-manager-6989c587cb          ReplicaSet          cluster  -       ok  -   23s
# tanzu-auth   L.. tanzu-auth-controller-manager-6989c587cb-qmgn8  Pod                 cluster  4/4 t   ok  -   23s
# tanzu-auth   L tanzu-auth-controller-manager-6989c587cb-qmgn8    PodMetrics          cluster  -       ok  -   1s
# tanzu-auth  tanzu-auth-controller-manager-sa                     ServiceAccount      kapp     -       ok  -   25s
# (cluster)   tanzu-auth-controller-manager                        ClusterRole         kapp     -       ok  -   25s
# (cluster)   tanzu-auth                                           Namespace           kapp     -       ok  -   25s
# (cluster)   tanzu-auth-controller-manager                        ClusterRoleBinding  kapp     -       ok  -   25s
#
# Rs: Reconcile state
# Ri: Reconcile information
#
# 8 resources
#
# Succeeded
```

And check the logs on the running pod

```bash
kapp logs --app tanzu-auth -f
```

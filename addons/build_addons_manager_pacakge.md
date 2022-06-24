# Building addons-manager package

To copy and paste below examples, make sure interactive mode is enabled if using zsh. Usually the case in mac os.

    setopt interactivecomments


## Set Variables
The following variables need to be set. The values used
below are only examples and should be modified for your development environment

    #user needs to provide this. Of the form accessible.registry.url/some/path
    export DEV_REGISTRY=<user-must-provide-this>

    # build version. Of the form "dev.x" where x is a number that should be incremented for every build
    #to assure the image will be updated on cluster
    export VERSION=dev.1

    # url of oci registry where images will be pushed to.  You should replace this with an accessible registry url.
    export OCI_REGISTRY=${DEV_REGISTRY}/$(whoami)


    # name to be used when pushing oci image
    export OCI_IMAGE_NAME=addons-controller



    # package version. Should use value appropriate for your development environment
    export PACKAGE_VERSION=v1.0.0

    # full oci image name
    export IMG=${OCI_REGISTRY}/${OCI_IMAGE_NAME}:${PACKAGE_VERSION}-${VERSION}

    # directory under tanzu-framework/packages which contains the package bundle config files
    export PACKAGE_REPOSITORY=management

    # directory under tanzu-framework/package/${PACKAGE_REPOSITORY}/ which contains the bundle config files
    export PACKAGE_NAME=addons-manager

    #variables needed by Makefile
    export PACKAGE_SUB_VERSION=$VERSION
    export PACKAGE_BUNDLES=${PACKAGE_NAME}
    export BUILD_VERSION=${PACKAGE_VERSION}


## Build and push OCI image
The OCI image of the addon-controller binary (manager) is built and pushed
to the OCI_REGISTRY so that it can be referenced by the addons-manager
package bundle.
from anywhere in the repo directory structure

    # switch to addons directory in repo
    cd $(git rev-parse --show-toplevel)/addons

    # build oci  image
    make docker-build

    # verify image
    docker image history $IMG

    # push image to $OCI_REGISTRY
    make docker-publish

    #(optional) verify the image was pushed
    docker pull $IMG

## Create package
    # switch to toplevel directory of repo
    cd $(git rev-parse --show-toplevel)

    # create packages/kbld-config.yaml file
    make kbld-image-replace

    # create thin package
    make package-bundle-thin

    # check package-bundles created
    tree -L  3 $(git rev-parse --show-toplevel)/build/

    # push package bundle to $OCI_REGISTRY
    make push-package-bundles

    # package bundle has been pushed to
    export PACKAGE_BUNDLE_URL=${OCI_REGISTRY}/${PACKAGE_NAME}:${PACKAGE_VERSION}_${VERSION}
    echo package bundle was pushed to $PACKAGE_BUNDLE_URL

    # verify package was pushed as stated
    docker pull $PACKAGE_BUNDLE_URL
    docker image history $PACKAGE_BUNDLE_URL


## Using the new package in a tkg cluster
With admin access to the k8s cluster

    #check access
    kubectl get pods -A

    # look for the addons-manager package
    kubectl get packages -A | grep addons-manager

Next we need to edit the package to use the new package bundle

    # set the imgpkgBundle.image:  in the addons-manager package cr
    # using the namespace and package name obtained in the previous command
    kubectl edit package -n vmware-system-tkg addons-manager.tanzu.vmware.com.2.0.12

    # replace the value of spec.template.spec.fetch.imgpkgBundle.image with $OCI_REGISTRY/addons-manager:v1.0.0_dev.x
    # (see previous steps for actual value)
    # once you save the package, it might timeout: "the server was unable to return a response in the time alloted..."
    # if so repeat the kubectl edit package command

    # if necessary, change the spec.syncPeriod of the tanzu-addons-manager package install CR
    # so the changes to the packet are picked up faster
    kubectl edit packageinstalls  -n vmware-system-tkg   tanzu-addons-manager

    # wait until status is ReconcileSucceeded
    kubectl get packageinstalls -n vmware-system-tkg   tanzu-addons-manager
    
    #check pods again
    kubectl get pods -A | grep addons

    # examine the deployment to see the new image
    kubectl get deployments -A | grep addons

    # with values obtained
    kubectl get deployments  -n vmware-system-tkg tanzu-addons-controller-manager  -o yaml | grep url

    # you can also check the pods. the sha256 should be listed in the output of last command
    kubectl get pods -A | grep addons-controller
    kubectl  get pods -n vmware-system-tkg  tanzu-addons-controller-manager-xxxxxx -o yaml | grep imageID

    # you can look at logs of pod to verify it is working or to debug
    kubectl logs -n vmware-system-tkg tanzu-addons-controller-manager-xxxxxx

    # finally, if required,  reset the spec.syncPeriod of the package installation back to original value
    kubectl edit packageinstalls  -n vmware-system-tkg   tanzu-addons-manager

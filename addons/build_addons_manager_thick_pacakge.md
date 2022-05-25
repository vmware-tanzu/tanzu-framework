# Building addons-manager package

## Set Variables
The following variables need to be set. The values used 
below are only examples and should be modified for your development environment

    # local registry information
    export registry_hostname=localhost
    export registry_port=5001
    
    export registry_oci_image=registry:2.7
    export local_registry_name=local_registry

    # url of oci registry where images will be pushed to.  replace $(whoami) with anystring you want if prefered.
    export OCI_REGISTRY=${registry_hostname}:${registry_port}
    
    # name to be used when pushing oci image 
    export OCI_IMAGE_NAME=addons-controller
 
    # build version. Of the form "dev.x" where x is a number that should be incremented for every build
    #to assure the image will be updated on cluster
    export VERSION=dev.1

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
    # if the following is not set, it will be derived from shell git describe --always --dirty --tags, and appropriate build
    # files will need to be edited.  Best to use "latest" for local builds
    export IMG_TAG_OVERRIDE=latest

## Start a local oci registry

Registry listens on localhost port 5001.

    export registry_oci_image=registry:2.7
    export local_registry_container=local_registry
    docker container stop ${local_registry_container} && docker container rm -v ${local_registry_container} || true
	docker run -d -p ${registry_port}:5000 --name ${local_registry_container} ${registry_oci_image}
    docker container logs ${local_registry_container}
    docker container list | grep "${local_registry_container}\|PORTS"

## Build and push OCI image
The OCI image of the addon-controller binary (manager) is built and pushed
to the OCI_REGISTRY so that it can be referenced by the addons-manager
package bundle.

    # switch to addons directory in repo
    cd $(git rev-parse --show-toplevel)/addons
    
    # build oci  image
    make docker-build
    
    # verify image
    docker image history $IMG

    # push image to localhost:5001
    make docker-publish
    
## Create thick package 
    # siwtch to toplevel directory of repo
    cd $(git rev-parse --show-toplevel)

    # create packages/kbld-config.yaml file
    make kbld-image-replace

    # create thin and thick bundles
    make package-bundle 
    
    # check package-bundles "-thick" file created
    tree -L  3 $(git rev-parse --show-toplevel)/build/

## Stop local registry
Optional, but if not needed best to stop local registry. 

    docker container stop ${local_registry_container} && docker container rm -v ${local_registry_container} || true

## Use thick package 
The created thick pacakge (_ addons-manager-xxx-thick.tar.gz_), may now be pushed to an external regidtry of your choice

    export EXTERNAL_REGISTRY=<some_other_oci_registry_image_url>

    export PACKAGE_BUNDLE_NAME=${PACKAGE_NAME}-${PACKAGE_VERSION}_${VERSION}
    export PATH_TO_THICK_PACKAGE=$(git rev-parse --show-toplevel)/build/package-bundles/${PACKAGE_REPOSITORY}/${PACKAGE_BUNDLE_NAME}-thick.tar.gz
    export EXTERNAL_REGISTRY_IMAGE=${EXTERNAL_REGISTRY}/${PACKAGE_NAME}
    imgpkg copy --tar ${PATH_TO_THICK_PACKAGE} --to-repo ${EXTERNAL_REGISTRY_IMAGE}
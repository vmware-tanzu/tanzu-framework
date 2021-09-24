# How to deploy management and workload cluster in air-gapped environment?

tkgctl library has a config variable `TKG_CUSTOM_IMAGE_REPOSITORY` which can be set as config value in config file or as an environment variable.

Setting up `TKG_CUSTOM_IMAGE_REPOSITORY` variable to point to internal docker registry which uses the CA signed certificate is sufficient while creating management and workload cluster with some assumption that all image are available in the provided custom repository.

Library also supports config variables `TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE` and `TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY` for using private docker registry which uses a self signed CA certificate. Library configures containerd service in kind cluster node and management cluster nodes to use the base64-encoded CA certificate specified via `TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE` to verify the private docker registry, or skip TLS verification if `TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY` is set to true.

## Publishing all images to custom image repository

1. Install [yq](https://github.com/mikefarah/yq) and [jq](https://stedolan.github.io/jq/)
1. Use script [gen-publish-images.sh](/hack/gen-publish-images.sh) to generate script which can publish images to custom image repository.

   ```sh
   # setup variable
   export TKG_CUSTOM_IMAGE_REPOSITORY="custom-image-repository.io"

   # Generate file for airgapped image publish
   ./gen-publish-images.sh > publish-images.sh

   # verify generated script
   cat publish-images.sh

   # make it executable
   chmod +x publish-images.sh

   # run the script to pull, retag and publish images to custom repository
   ./publish-images.sh
   ```

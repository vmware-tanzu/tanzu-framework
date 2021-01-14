# TKG Addons for Pinniped

## How to install and use the Pinniped addon using ytt

- Create the ytt value yaml(`values.yaml`) based on the example from [here](./examples), and substitute the values for your case
- Run `ytt --ignore-unknown-comments -f ../ytt-common-libs -f ../pinniped/templates -f values.yaml > pinniped.yaml`
- Run `kubectl apply -f pinniped.yaml`. Once you have applied the `pinniped.yaml` you should be able to see required 
  resources are creating. The post deploy job 
  will also be running to configure the Pinniped and Dex installation to make all things work. Wait for a while until the 
  post deploy job to complete.
  - If using external OIDC provider, you might want to configure the OIDC provider by providing correct login redirect url to Dex service
- Copy the Pinniped-cli from [here](./post-deploy/hack/bin/pinniped-cli) or download from Pinniped [Github](https://github.com/vmware-tanzu/pinniped/releases)
- Run `./pinniped-cli get kubeconfig` to get the updated kubeconfig which could be distributed to authorized users. The 
  users with that kubeconfig will be redirected to configured authentication page.
  
## Images

### Post deploy job images

- Dev image: gcr.io/kubernetes-development-244305/gdaniel/tkg-pinniped-post-deploy:with-dex
- Latest tested image: gcr.io/kubernetes-development-244305/gdaniel/tkg-pinniped-post-deploy:latest

### Pinniped addon template images

- Dev image: gcr.io/kubernetes-development-244305/gdaniel/tkg-addons-pinniped-templates:dev
- Latest tested image: gcr.io/kubernetes-development-244305/gdaniel/tkg-addons-pinniped-templates:latest
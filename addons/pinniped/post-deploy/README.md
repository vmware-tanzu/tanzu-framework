# tkg-pinniped-post-deploy

**Note:** This doc is out of date

This repo contains the implementation for Pinniped post deployment configuration on TKG. TKG provides an automated way 
for users to install Pinniped on TKG clusters, however there are some configurations need to be seamlessly handled. e.g. 
configure `FederationDomain` issuer and `JWTAuthenticator` issuer. 

## Workflow

### Management cluster

- User creates `Secret` with ytt values of Pinniped to notify addons-manager the installation of Pinniped and Dex addon.
- The addons-manager figures out the correct ytt template and create `App` for kapp-controller to consume. The logic
  within this repo will be packaged as a container and wrapped in above ytt template as a `Job`. The `Job` will be run
  as the last step to configure Pinniped and Dex.

#### What are handled by `tkg-pinniped-post-deploy`

- Autodetect the IP address or DNS name of Pinniped supervisor service
- Create or update `Certificate` of Pinniped with valid service IP address or DNS name of Pinniped supervisor
- Create or update `JWTAuthenticator` with valid issuer IP address or DNS name of Pinniped supervisor
- Create or update `FederationDomain` with valid issuer IP address of DNS name of Pinniped supervisor
- Create `ConfigMap` with Pinniped information under kube-public namespace
- Create or update `Certificate` of Dex with valid service IP address or DNS name of Dex service endpoint
- Create or update `OIDCIdentityProvider` with valid service IP address or DNS name of Dex service endpoint and Dex's CaBundle
- Update `ConfigMap` of Dex with valid issuer IP address or DNS name of Pinniped supervisor, valid service IP address or DNS name of Dex service endpoint and randomly generated client secret
- Update `PinnipedOidcSecret` with the same client secret
  
### Workload cluster

- User creates `Secret` with ytt values of Pinniped to notify addons-manager the installation of Pinniped addon.
- The addons-manager figures out the correct ytt template and create `App` for kapp-controller to consume. The logic
  within this repo will be packaged as a container and wrapped in above ytt template as a `Job`. The `Job` will be run
  as the last step to configure Pinniped.

#### What are handled by `tkg-pinniped-post-deploy`

- Create or update `JWTAuthenticator` with the Pinniped supervisor info passed in. User should be able to fetch the Pinniped
  supervisor info from pinniped-info ConfigMap on management cluster.

### List of changes

#### Pinniped
* FederationDomain
  * SupervisorServiceEndpoint
* PinnipedCertificate
  * SupervisorServiceEndpoint
* JWTAuthenticator
  * SupervisorServiceEndpoint
  * PinnipedCaBundle
* ConfigMap
  * SupervisorServiceEndpoint
  * PinnipedCaBundle
  
#### Dex
* DexCertificate
  * DexEndpoint
* IDP
  * DexEndpoint
  * DexCaBundle
* ConfigMap
  * SupervisorEndpoint
  * DexEndpoint
  * ClientSecret
* PinnipedOidcSecret
  * ClientSecret

## How to run the post deploy from local env

- Set the `KUBECONFIG` pointing to the cluster
- `make run` (**Note**: You might want to substitute the flags in Makefile used by the `make run` to satisfy your test)

## How to run everything all together

### OIDC

- Login to your OIDC provider dashboard, e.g. https://www.okta.com/, create the web application and remember the clientID, secret and the url(e.g. https://dev-xxxxxx.okta.com). 
- Get the ytt value examples from [here](../examples), update the OIDC related values.
- Render the YAML by using ytt command
  - Go to addons root dir
  - Run `ytt --ignore-unknown-comments -f ./ytt-common-libs -f ./pinniped/templates -f ./pinniped/examples/mc-vsphere-oidc.yaml > tmp.yaml`
- Run `kubectl apply -f tmp.yaml`. You should be able to see all things are deployed and post deploy job is completed after a while
- Login to your OIDC provider dashboard, e.g. https://www.okta.com/, direct to the application and make sure the login redirect url is pointing to the Pinniped supervisor service
  It could be external IP address or DNS name. E.g. https://<pinniped-supervisor-svc-ip>/callback
- Run `./hack/bin/pinniped-cli get kubeconfig > tmp.kubeconfig` to get the kubeconfig configured by Pinniped
- Run `kubectl --kubeconfig=./tmp.kubeconfig get pods`, you should be redirected to the browser which asks the login from OIDC provider

### LDAP 

- Get the ytt value examples from [here](../examples), update the LDAP related values including LDAP host, userSearch and groupSearch policies.
- Render the YAML by using ytt command
  - Go to addons root dir
  - Run `ytt --ignore-unknown-comments -f ./ytt-common-libs -f ./pinniped/templates -f ./pinniped/examples/mc-vsphere-ldap.yaml > tmp.yaml`
- Run `kubectl apply -f tmp.yaml`. You should be able to see all things are deployed and post deploy job is completed after a while
- Run `./hack/bin/pinniped-cli get kubeconfig > tmp.kubeconfig` to get the kubeconfig configured by Pinniped
- Run `kubectl --kubeconfig=./tmp.kubeconfig get pods`, you should be redirected to the browser which asks the login from LDAP provider

## How to build docker images

**Note**: The dev image is under: `gcr.io/kubernetes-development-244305/gdaniel/tkg-pinniped-post-deploy:with-dex`. 

You could also build your own images by using `make build-images`

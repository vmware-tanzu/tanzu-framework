# Deploying the local changes to a Kind cluster

## Pre-requisites
1. Docker
2. Kind
3. ytt
4. Kubectl

## Step 1 - Create a Kind cluster
A simple Kind cluster can be created for testing using the following command.

```
kind create cluster
```

## Step 2 - Build docker image

Run the following command from the `tanzu-framework` directory.

```
make docker-build-all
```

## Step 3 - Load readiness controller image

Load the readiness controller image into the kind nodes by running the following command.

```
kind load docker-image readiness-controller-manager:latest
```

## Step 4 - Deploy the manifests

Run the following command to deploy CRDs and bring up the readiness controller 

```
ytt -f packages/readiness/bundle/config | kubectl apply -f-
```
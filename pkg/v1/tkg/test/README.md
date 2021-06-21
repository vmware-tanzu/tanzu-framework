# TKG Integration tests

## What are these tests?
TKG integration tests validates the `tkgctl` library used underneath by the cluster, management-cluster plugins and also by the TMC to integrate with TKG. 
These tests leverage CAPD provider to be able to run without needing access to any external infrastructure.

## How to run these tests locally?
They are developed using the [Ginkgo testing framework](https://github.com/onsi/ginkgo). Here are the steps to run them locally:
```
$ make tools
$ E2E_CONFIG=pkg/v1/tkg/test/config/docker.yaml hack/tools/bin/ginkgo -v -trace pkg/v1/tkg/test/tkgctl/docker
```

If you want to run a single test 
```
E2E_CONFIG=pkg/v1/tkg/test/config/docker.yaml hack/tools/bin/ginkgo -v -trace --focus="<test spec name/regex>" pkg/v1/tkg/test/tkgctl/docker
```

## Troubleshooting
If you are running these CAPD based integration tests on MAC, make sure that there are enough resources allocated to docker in the docker settings, here are some recommendations: 
```
Memory: 6GB+
Swap: 2GB+
Disk image size: 100GB+
```
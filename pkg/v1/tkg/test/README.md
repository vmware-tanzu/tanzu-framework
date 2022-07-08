# TKG Integration tests

## What are these tests?

TKG integration tests validates the tkg client library used by the cluster, management-cluster plugins.
Some of these tests leverages CAPD provider to be able to run without needing access to any external cloud infrastructure.

## How to run these tests locally?

They are developed using the [Ginkgo testing framework](https://github.com/onsi/ginkgo). Here are the steps to run them locally:

```sh
make tools
E2E_CONFIG=pkg/v1/tkg/test/config/docker.yaml
hack/tools/bin/ginkgo -v -trace pkg/v1/tkg/test/tkgctl/docker
```

If you want to run a single test

```sh
E2E_CONFIG=pkg/v1/tkg/test/config/docker.yaml
hack/tools/bin/ginkgo -v -trace --focus="<test spec name/regex>" pkg/v1/tkg/test/tkgctl/docker
```

In case of TKGS test cases, make sure sshuttle is running to connect tkgs cluster, also set HTTP_PROXY and HTTPS_PROXY values. Generate kube config file using the `kubectl vsphere log`, the token in tkgs kube config expires every 24hours, so need to regenerate kube config file. For the tkgs E2E_CONFIG config file, refer existing tkgs config from pkg/v1/tkg/test/config/tkgs.yaml and do required changes.
example:

```sh
sshuttle -r root@10.199.56.233 192.163.0.0/8 #let it keep running while running test cases
E2E_CONFIG=pkg/v1/tkg/test/config/tkgs.yaml
HTTP_PROXY=192.163.1.163:3128
HTTPS_PROXY=192.163.1.163:3128
hack/tools/bin/ginkgo -v -trace pkg/v1/tkg/test/tkgctl/tkgs
```

## Troubleshooting

If you are running the CAPD based integration tests on an Apple system, it is recommended that Docker Desktop is allocated the following resources at
minimum:

- Memory: 6GB+
- Swap: 2GB+
- Disk image size: 100GB+

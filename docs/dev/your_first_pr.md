# Your First PR

Thanks for choosing to help improve the Tanzu Framework! This guide will
help you build your first successful pull request.

## Opening an Issue

Your new idea starts with a new GitHub issue. You'll link your pull request to
your issue once you're ready to submit it. This will help us appropriately
prioritize the release into which your PR should get merged into.

[Click here](https://github.com/vmware-tanzu/tanzu-framework/issues/new/choose) to
open your first issue.

## Getting the Code

Now that you've created your issue, your next step is to fork the repo.

Your pull request will merge your fork into our repository.

Click on the "Fork" button at the top right of the page to fork
Tanzu Framework into your GitHub account.

✅ **NOTE**: You'll need your terminal for the steps below.

## Preparing Your Development Environment

Next, we'll need to prepare your development environment.

Before doing anything, go ahead and clone your fork with `git clone`.

### Download and install Golang and GVM

Tanzu Kubernetes Grid is written in [Golang](https://golang.org), or Go
for short. While it is currently built against Go v${VERSION} this might change
in the future.

Golang Version Manager, or [GVM](https://github.com/moovweb/gvm), is a tool
that enables you to maintain multiple versions of Go on your computer. If
you've ever developed a Rails or Python app, `rvm` and `pyenv` are similar.

Here, we will install Golang and GVM.

First, we will need to install a "base" version of Golang, as GVM can
compile Golang from source, and Golang compiles itself in Golang:

```sh
# Mac
brew install golang

# Windows
choco install golang

# Linux
curl -o /tmp/golang.tar.gz https://golang.org/dl/go1.17.1.linux-amd64.tar.gz && \
   tar -xzf /tmp/golang.tar.gz -C /usr/local &&
   export PATH="$PATH:/usr/local/go/bin"
```

Next, we'll install GVM and use it to install Go v1.17:

```sh
GOLANG_VERSION=1.17
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
source ~/.gvm/scripts/gvm
gvm install "go$GOLANG_VERSION" && gvm use "go$GOLANG_VERSION"
```

### Install dependencies

```sh
cd tanzu-framework
go get -v -t -d ./...
```

Next, `cd` into `tanzu-framework` and download `tanzu-framework` dependencies:

### Configure the Bill of Materials

`make configure-bom`

The Bill-of-Materials, or the BoM, outlines the components required to download TKG components, such
as Docker images and ClusterAPI providers. We will need to generate this file so that
the `tanzu` client can initialize itself. Dummy values will be used during testing.

### Compile CLI components

```sh
make manager build-cli
```

Next, we'll build the CLI components. The `manager` binary is used as an entrypoint
for various TKG components.

## Create a test

### Ginkgo and Gomega

Tanzu Framework uses [Ginkgo](https://onsi.github.io/ginkgo/) to create tests in
[Gherkin](https://cucumber.io/docs/gherkin/) syntax. Gherkin is a syntax
used for enabling [behavior-driven development](https://en.wikipedia.org/wiki/Behavior-driven_development).
Matchers in Ginkgo are provided by [Gomega](https://onsi.github.io/gomega).

The test for most code in this repository is located inside of the `_test`
version of the file from which it originated. For instance, if I wanted to
modify `CreateCluster` inside of `pkg/v1/tkg/tkgctl/create_cluster.go`,
I would look for its test in `pkg/v1/tkg/tkgctl/create_cluster_test.go`.

### New `interface{}`s and fakes

Most `interface{}`s have a corresponding test double. Test doubles are generated
automatically by [Counterfeiter](https://github.com/maxbrunsfeld/counterfeiter/).
Counterfeiter uses an annotating comment above the `interface{}` for whom a double
should be created, and it creates a double for you.

If you need to implement a new `interface{}` in `pkg/v1/tkg/cool_thing/thing.go`, ensure that you:

1. Add this comment above it:

   ```go
   //go:generate counterfeiter -o ../fakes/cool_thing/your_interface.go --fake-name CoolThing YourInterface`
   ```

2. Run `go generate ./...`

This will create a double for `YourInterface` inside of `pkg/v1/fakes/cool_thing/your_interface.go`.

❌ **WARNING**: **Do not edit the test double yourself!** ❌

## Get your test passing

```sh
go test ./...
```

As you develop the code for your new feature or fix, you should `go test`
from time to time. Tests take about 15 seconds to complete on an
2020 MacBook Pro with an Intel i7 CPU and 32GB RAM.

### Run an end-to-end test locally

Once you're done implementing your feature, we encourage you to test it out
with a real Tanzu Kubernetes Grid cluster. Fortunately, this can be done entirely
locally!

✅ **NOTE**: If you are using Docker for Mac or Docker for Windows, ensure
that it is configured to use at least 100GB of disk space. ✅

### Rebuild the CLI components

```sh
make build-cli
```

When you're ready to test locally, rebuild the CLI with the `make` command above.

## Create the config files

```sh
# management cluster
cat >/tmp/config.yaml <<-CONFIG
providers:
- name: docker
  url: ./pkg/v1/providers/infrastructure-docker/v0.3.23/infrastructure-components.yaml
  type: InfrastructureProvider
CLUSTER_NAME: test
INFRASTRUCTURE_PROVIDER: docker
CLUSTER_PLAN: dev
ENABLE_CEIP_PARTICIPATION:  false
ENABLE_MHC: false
SERVICE_CIDR: "10.232.0.0/16"
CLUSTER_CIDR: "192.168.1.0/24"
CNI: calico
SIZE: 2
CONFIG

# workload cluster
cat >/tmp/config-workload.yaml <<-CONFIG
providers:
- name: docker
  url:  /Users/ncarlos/src/tanzu-framework/pkg/v1/providers/infrastructure-docker/v0.3.23/infrastructure-components.yaml
  type: InfrastructureProvider
CLUSTER_NAME: test-workload
CLUSTER_PLAN: dev
INFRASTRUCTURE_PROVIDER: docker
SIZE: 2
ENABLE_MHC: false
CLUSTER_CIDR: "192.168.0.0/16"
SERVICE_CIDR: "10.232.1.0/24"
CNI: calico
CONFIG
```

First, create the config files for your management and (if needed) workload cluster so that it uses
the Docker ClusterAPI provider using a `dev` plan (or `prod`) if your change
requires it.

⚠️  **WARNING**: The `docker` provider is not officially supported by VMware,
and we do not recommend using it for production workloads! ⚠️

✅ **NOTE**: This workflow will also work if your feature requires a specific
provider, like AWS or vSphere. Your `config.yaml` will look different,
however. ✅

## Create the management cluster

```sh
./artifacts/$(uname)/amd64/cli/core/latest/tanzu-core-darwin_amd64 \
  management-cluster create [CLUSTER_NAME] -f /tmp/config.yaml
```

Next, create the management cluster with the `tanzu` CLI binary that you built
locally. This will take about five to ten minutes to complete.

## Create the workload cluster (if needed)

```sh
./artifacts/$(uname)/amd64/cli/core/latest/tanzu-core-darwin_amd64 \
  cluster create [CLUSTER_NAME] -f /tmp/config.yaml
```

Next, use the same config file to create a workload cluster from the management
cluster if your feature necessitates it.

## Hack

Test your feature! Remember that you can get the Kubeconfig for your
workload cluster by running:

```sh
./artifacts/$(uname)/amd64/cli/core/latest/tanzu-core-darwin_amd64 \
  cluster kubeconfig get [CLUSTER_NAME] -f /tmp/config.yaml
```

## Teardown

Because we are running this locally, tearing everything down is insanely easy:

```sh
kind get clusters | xargs -I {} kind delete cluster --name {}
```

❌ **WARNING**: This command will delete _all_ Kind clusters on your machine!
Use `grep -Ev` to prevent this.

## Submit your PR

You're now ready to submit your PR! 🚀

[Click here](https://github.com/vmware-tanzu/tanzu-framework/pulls/new) to do so.
Remember to merge the branch in your fork against `main` on ours.

## Add some documentation

If your change requires documentation, consider adding it in the `docs/` directory.

Before doing so, download `markdownlint` from Homebrew, Chocolatey, or `npm` and ensure that it is
passing!

```sh
# ❌ incorrect
$: markdownlint docs/dev/your_first_pr.md
docs/dev/your_first_pr.md:6 MD025/single-title/single-h1 Multiple top-level headings in the same document [Context: "# Opening an Issue"]
docs/dev/your_first_pr.md:15 MD025/single-title/single-h1 Multiple top-level headings in the same document [Context: "# Getting the Code"]
docs/dev/your_first_pr.md:26 MD025/single-title/single-h1 Multiple top-level headings in the same document [Context: "# Preparing Your Development E..."]
docs/dev/your_first_pr.md:95 MD025/single-title/single-h1 Multiple top-level headings in the same document [Context: "# Create a test!"]
docs/dev/your_first_pr.md:95:16 MD026/no-trailing-punctuation Trailing punctuation in heading [Punctuation: '!']
docs/dev/your_first_pr.md:130 MD025/single-title/single-h1 Multiple top-level headings in the same document [Context: "# Get your test passing!"]
docs/dev/your_first_pr.md:130:24 MD026/no-trailing-punctuation Trailing punctuation in heading [Punctuation: '!']
docs/dev/your_first_pr.md:140 MD025/single-title/single-h1 Multiple top-level headings in the same document [Context: "# Run an end-to-end test local..."]
docs/dev/your_first_pr.md:140:33 MD026/no-trailing-punctuation Trailing punctuation in heading [Punctuation: '!']
docs/dev/your_first_pr.md:246 MD025/single-title/single-h1 Multiple top-level headings in the same document [Context: "# Submit your PR"]
```

```sh
# ✅ Correct
$: markdownlint docs/dev/your_first_pr.md
$: # all clear!
```

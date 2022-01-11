# Release notes process

This document provides guidance on how to gather release notes from the pull
requests that have user-visible changes that will be part of the release.

We would be using a tool called [release-notes](https://github.com/kubernetes/release/tree/master/cmd/release-notes)
that is responsible for fetching, contextualizing, and rendering the release
notes for Tanzu Framework.

## Installation

The simplest way to install the release-notes CLI is via `go install`:

```sh
go install k8s.io/release/cmd/release-notes@latest
```

## Usage

To generate release notes for a commit range, we would be needing some basic
information that we need to pass to the `release-notes` command.

* GitHub API token.
* Branch or tag name.
* The commit hash to start processing from (inclusive).
* The commit hash to end processing at (inclusive).

Then run the below command to generate the release notes for a commit range:

```sh
export GITHUB_TOKEN=${GITHUB_API_TOKEN}
release-notes \
    --github-base-url https://github.com \
    --org vmware-tanzu-private \
    --repo core \
    --branch ${tag_or_branch} \
    --required-author "" \
    --start-sha ${start_sha} \
    --end-sha ${end_sha}
```

The above command writes the release notes to a file that will be printed in
output of this command. The contents of the generated file should be pasted in
release description when publishing the release.

For more information on the `release-notes` command and the options to run it,
check [here](https://github.com/kubernetes/release/tree/master/cmd/release-notes)

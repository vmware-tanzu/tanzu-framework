# Release process

The person who is responsible for cutting a release, pushes a tag to the
GitHub repo to trigger the automation. A GitHub action then does the
following things:

1. Runs a full build via `make all` to build the UI and CLI.
1. Runs the tests.
1. Uploads the CLI artifacts to GCP buckets.
1. Creates a draft github release.

After the GitHub action runs successfully without failure, the person who is
cutting the release:

1. Runs the [release-notes](release-notes-gathering-process.md) command to gather the
   release notes from the pull requests that are part of the release.
1. Goes to the releases page and adds the release notes to the Draft release
   and publishes the release.

# Providers Sync Tools

## Overview

Tanzu-framework is moving towards packaging cluster-api componentry. While this
transition is happening, tanzu-framework will have cluster-api components in
two locations: `packages/cluster-api*` and `providers`.

These tools exist to ensure cluster-api related packages in the
`packages` directory stay in sync with the content in
`providers`.

## How it works

The tooling in this directory vendirs the cluster-api packages into a git
ignored directory:

`hack/provider-sync-tools/<some cluster-api package>/build/upstream`

The tooling then renders the files into `hack/provider-sync-tools/<some
cluster-api-package>/build/rendered` using ytt. The render includes two sets of
overlays.

1. The first set of overlays come from the package's overlays.

1. The second set of overlays are specific to edits needed to
   `providers`, and exist in the `hack/provider-sync-tools/<some
   cluster-api package>/overlay`. An example of such an edit would be the
   controller image in the CRDs of the package.

This tooling will prevent edits made directly to the cluster-api related
folders in `providers`. Instead of directly editing those files,
overlays should be created in `hack/provider-sync-tools/<some cluster-api
package>/overlay`. Running the `make all` target in this folder will render the
package contents along with the overlays, and then place the edited files in
`providers`.

The validate task will ensure that the rendered content from the package and
overlays have no diff with the content of `providers`.

## What to do in common scenarios

1. Making a tweak to `providers/<cluster-api-folder>`
    - add an overlay file to `hack/provider-sync-tools/<some cluster-api
      package>/overlays`.
      NOTE: There are many files in `providers/<cluster-api-folder>/` and
      only some of them are copied from upstream cluster-api, typically CRDs.
      These sync tools are intended to sync the files that come from upstream
      and the tweaks to files that come from upstream. Files in
      `providers/<cluster-api-folder>` that are not from upstream should
      be edited in place, i.e. cluster configuration files.
    - run `make all` to and check to see that `providers/<cluster-api
      folder>` contents looks as desired.
    - commit the changes

1. Bumping cluster-api to a newer version
    - edit the package(s) `packages/<some-cluster-api-package>/`
      vendir file to match the desired version
    - run the unit tests in the package(s), ensuring they're passing.
    - `cd hack/provider-sync-tools && make all`
    - ensure the `providers/<cluster-api-folder>` looks as desired.
    - commit the changes

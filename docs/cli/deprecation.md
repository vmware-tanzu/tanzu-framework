# Tanzu CLI deprecation policy

Tanzu CLI is comprised of several different commands and plugins.
Sometimes, a release might remove flags or CLI commands (collectively
"CLI elements").
This document sets out the deprecation policy for Tanzu CLI.

CLI elements must function after their announced deprecation for no less than
one minor release.

## How do we deprecate things?

A **deprecation event** would coincide with a release. The rules for
deprecation are as follows:

1. Deprecated CLI elements must display warnings when used.
1. The warning message should include a functional alternative to the
   deprecated command or flag if anything.
1. The warning message should include the release for when the command/flag
   will be removed.
1. The deprecation should be documented in the Release notes to make users
   aware of the changes.

**Example warning message**:

Command "foo" is deprecated, will be removed in version "x.y.z". Use "bar"
instead.

This [file](../../pkg/v1/cli/deprecation.go) in the cli package has a utility
function that can be used to deprecate a command.

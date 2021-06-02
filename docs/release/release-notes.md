# Release notes

This document provides guidance on providing release notes for changes made to
the core repo. Release notes act as a direct line of communication with the
users, making them aware of the changes and act as reference points for users 
about to install or upgrade to a particular release.

## Table of Contents
* [Does my pull request need a release note?](#does-my-pull-request-need-a-release-note)
* [Contents of a release note](#contents-of-a-release-note)
* [Applying a Release Note](#applying-a-release-note)
* [Reviewing Release Notes](#reviewing-release-notes)

## Does my pull request need a release note?

Any pull request with user-visible changes are required to add release notes,
this could mean:

- User facing, critical bug-fixes
- Notable feature additions
- Output format changes
- Deprecations or removals
- Metrics changes
- Dependency changes
- API changes
- CLI Changes  
- Configuration schema change
- Fix of a vulnerability (CVE)

## Contents of a release note

A release note needs a clear, concise description of the change in simple plain language.
This includes:

* an indicator if the pull request Added, Changed, Fixed, Removed, Deprecated 
  functionality or changed enhancement.
* the name of the affected API or configuration fields, CLI commands/flags.
* a link to relevant user documentation about the enhancement/feature.

Your release note should be written in clear and straightforward sentences.
Not all users are familiar with the technical details of your pull request,
so consider what they need to know when you write your release note.

Write the release note like you are in the future like:
* "Added" instead of "add"
* "Fixed" instead of "fix"
* "Bumped" instead of "bump"

Some examples of release notes:
* The command foo has been deprecated, will be removed in version "1.5.0".
  Users of this command should use "bar" instead.
* Fixed a bug that prevents CLI from initializing.

## Applying a Release Note
Any pull request with user visible changes, should include a release-note 
section in the pull request description.

For pull requests with a release note:

    ```release-note
    Your release note here
    ```

For pull requests with no release note:

    ```release-note
    NONE
    ```

## Reviewing Release Notes
Release note should be reviewed as a dedicated step in the overall pull request
review process.

A release note needs to be changed or rephrased if one of the following cases
apply:
* The release note does not communicate the full purpose of the change.
* The release note is grammatically incorrect.
* The release does not comply with the guidelines on the contents of the 
  release note.


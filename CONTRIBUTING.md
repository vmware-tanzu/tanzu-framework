Table of Contents
=================

* [Contributing](#contributing)
* [Communication](#communication)
* [Propose a Change](#propose-a-change)
  * [Pull Request Etiquette](#pull-request-etiquette)
  * [Issues Lifecycle](#issues-lifecycle)
  * [DCO Sign off](#dco-sign-off)
  * [Commit Messages](#commit-messages)

# Contributing

We’d love to accept your patches and contributions to this project. Please
review the following guidelines you'll need to follow in order to make a
contribution.

If you are interested in going beyond a single PR, take a look at our 
[Contribution Ladder](docs/contribution-ladder.md) and learn how to become a 
reviewer.

# Communication

We prefer communicating asynchronously through GitHub issues and the [#TBD
Slack channel](https://kubernetes.slack.com/archives/CQCDFHWFR). In order to be
inclusive to the community, if a conversation related to an issue happens
outside of these channels, we appreciate summarizing the conversation's context
and adding it to an issue.

# Propose a Change

Pull requests are welcome for all changes. When adding new functionality, we
encourage including test coverage. If significant effort will be involved, we
suggest beginning by submitting an issue so any high level feedback can be
addressed early.

Please submit feature requests and bug reports by using GitHub issues and filling
the supplied template with as much detail as you can.

Before submitting an issue, please search through open ones to ensure others
have not submitted something similar. If a similar issue exists, please add any
additional information as a comment.

## Pull Request Etiquette

* Before submitting a pull request, please make sure you verify the changes
  locally. The `Makefile` in this repository provides useful targets such as
  `lint` and `test` to make verification easier.
* Prefer small commits and small pull requests over large changes.
  Small changes are easier to digest and get reviewed faster. If you find
  that your change is getting large, break up your PR into small, logical
  commits. Consider breaking up large PRs into smaller PRs, if each of them
  can be useful on its own.
* Pull requests *must* include a `Fixes #NNNN` or `Updates #NNNN` comment. Remember
  that `Fixes` will close the associated issue, and `Updates` will link the PR to it.
* Have good commit messages. Please see [Commit Messages](#commit-messages)
  section for guidelines.

## Issues Lifecycle

Once an issue is labeled with `in-progress`, a team member has begun
investigating it. We keep `in-progress` issues open until they have been
resolved and released. Once released, a comment containing release information
will be posted in the issue's thread.

## Commit Messages

- Commit messages should include a short (72 chars or less) title summarizing the change.
- They should also include a body with more detailed explanatory text, wrapped to 72 characters.
  - The blank line separating the summary from the body is critical (unless you omit the body entirely).
- Commit messages should be written in the imperative: "Implement feature" and not "Implemented feature".
- Bulleted points are fine.
- Typically a hyphen or asterisk is used for the bullet, followed by a single space.

## Merging Commits

Maintainers should prefer to merge pull requests with the [Squash and merge](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-request-merges#squash-and-merge-your-pull-request-commits) option.
This option is preferred for a number of reasons.
First, it causes GitHub to insert the pull request number in the commit subject
which makes it easier to track which PR changes landed in.
Second, a one-to-one correspondence between pull requests and commits makes it
easier to manage reverting changes.

At a maintainer's discretion, pull requests with multiple commits can be merged
with the [Create a merge commit](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-request-merges)
option. Merging pull requests with multiple commits can make sense in cases 
where a change involves code generation or mechanical changes that can be 
cleanly separated from semantic changes. The maintainer should review commit
messages for each commit and make sure that each commit builds and passes
tests.

## Building From Source

To build the project from source, please consider the docs on [local development](docs/dev/build.md).

## DCO Sign off

All authors to the project retain copyright to their work. However, to ensure
that they are only submitting work that they have rights to, we are requiring
everyone to acknowledge this by signing their work.

To sign your work, just add a line like this at the end of your commit message:

```
Signed-off-by: Joe Beda <joe@heptio.com>
```

This can easily be done with the `--signoff` option to `git commit`.

By doing this you state that you can certify the following (from https://developercertificate.org/):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

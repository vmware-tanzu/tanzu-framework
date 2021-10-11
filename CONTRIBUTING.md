# Contributing

## Table of Contents

* [Contributing](#contributing)
* [Communication](#communication)
* [Propose a Change](#propose-a-change)
  * [Pull Request Etiquette](#pull-request-etiquette)
  * [Issues Lifecycle](#issues-lifecycle)
  * [Commit Messages](#commit-messages)
* [Contributor License Agreement](#contributor-license-agreement)

--------------

We’d love to accept your patches and contributions to this project. Please
review the following guidelines you'll need to follow in order to make a
contribution.

If you are interested in going beyond a single PR, take a look at our
[Contribution Ladder](docs/contribution-ladder.md) and learn how to become a
reviewer.

## Communication

We prefer communicating asynchronously through GitHub issues and the [#TBD
Slack channel](https://kubernetes.slack.com/archives/CQCDFHWFR). In order to be
inclusive to the community, if a conversation related to an issue happens
outside of these channels, we appreciate summarizing the conversation's context
and adding it to an issue.

## Propose a Change

✅ Is this your first pull request? Check out
[how to create your first pull request](./docs/dev/your_first_pr.md) first. Also,
thanks for contributing!

Pull requests are welcome for all changes. When adding new functionality, we
encourage including test coverage. If significant effort will be involved, we
suggest beginning by submitting an issue so any high level feedback can be
addressed early.

Please submit feature requests and bug reports by using GitHub issues and filling
the supplied template with as much detail as you can.

Before submitting an issue, please search through open ones to ensure others
have not submitted something similar. If a similar issue exists, please add any
additional information as a comment.

### Pull Request Etiquette

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

### Issues Lifecycle

Once an issue is labeled with `in-progress`, a team member has begun
investigating it. We keep `in-progress` issues open until they have been
resolved and released. Once released, a comment containing release information
will be posted in the issue's thread.

### Commit Messages

* Commit messages should include a short (72 chars or less) title summarizing the change.
* They should also include a body with more detailed explanatory text, wrapped to 72 characters.
  * The blank line separating the summary from the body is critical (unless you omit the body entirely).
* Commit messages should be written in the imperative: "Implement feature" and not "Implemented feature".
* Bulleted points are fine.
* Typically a hyphen or asterisk is used for the bullet, followed by a single space.

### Merging Commits

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

### Building From Source

To build the project from source, please consider the docs on [local development](docs/dev/build.md).

## Contributor License Agreement

All contributors to this project must have a signed Contributor License
Agreement (**"CLA"**) on file with us. The CLA grants us the permissions we
need to use and redistribute your contributions as part of the project; you or
your employer retain the copyright to your contribution. Before a PR can pass
all required checks, our CLA action will prompt you to accept the agreement.
Head over to [https://cla.vmware.com/](https://cla.vmware.com/) to see your
current agreement(s) on file or to sign a new one.

We generally only need you (or your employer) to sign our CLA once and once
signed, you should be able to submit contributions to any VMware project.

Note: if you would like to submit an "_obvious fix_" for something like a typo,
formatting issue or spelling mistake, you may not need to sign the CLA.

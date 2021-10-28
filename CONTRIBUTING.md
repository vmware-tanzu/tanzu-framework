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
[Contribution Ladder](docs/community/contribution-ladder.md) and learn how to become a
reviewer.

## Communication

We prefer communicating asynchronously through GitHub issues. We currently do
not have a slack channel or email list (hopefully soon).

## Propose a Change

We suggest beginning by submitting an issue so any high level feedback can be
addressed early, particularly if significant effort will be involved.

Please submit feature requests and bug reports by using GitHub issues and filling
in the supplied template with as much detail as you can.

Before submitting an issue, please search through open ones to ensure others
have not submitted something similar. If a similar issue exists, please add any
additional information as a comment.

## Contribute a Change

Pull requests are welcome for all changes, whether they are improving
documentation, fixing a bug, adding or enhancing a feature, or fixing a typo.

Changes to the behavior of the `tanzu-framework` itself will require that you
build and test your changes.

When adding new functionality, or fixing bugs, add appropriate test coverage
where possible. Different parts of the code base have different strategies and
patterns for testing, some of which may be in flux at any point in time.
Consider commenting on the issue to seek input or  or opening a draft pull
request to seek feedback on approaches to testing a particular change.

To build the project from source, please consider the docs on [local development](docs/dev/build.md).

### Commit Messages

* Commit messages should include a short (72 chars or less) title summarizing the change.
* They should also include a body with more detailed explanatory text, wrapped to 72 characters.
  * The blank line separating the summary from the body is critical (unless you omit the body entirely).
* Commit messages should be written in the imperative: "Implement feature" and not "Implemented feature".
* Bulleted points are fine.
* Typically a hyphen or asterisk is used for the bullet, followed by a single space.

## Pull Request Process

### Creating a Pull Request

Use the pull request template to provide a complete description of the change.
The template aims to capture important information to streamline the review
process, ensure your changes are captured in release notes, and update related
issues. Your pull request description and any discussion that follows is a
contribution itself that will help the community and future contributors
understand the project better.

* Before submitting a pull request, please make sure you verify the changes
  locally. The `Makefile` in this repository provides useful targets such as
  `lint` and `test` to make verification easier.
* Prefer small commits and small pull requests over large changes.
  Small changes are easier to digest and get reviewed faster. If you find
  that your change is getting large, break up your PR into small, logical
  commits. Consider breaking up large PRs into smaller PRs, if each of them
  can be useful on its own.
* Have good commit messages. Please see [Commit Messages](#commit-messages)
  section for guidelines.
* Pull requests *should* reference an existing issue and include a `Fixes #NNNN`
  or `Updates #NNNN` comment. Remember that `Fixes` will close the associated
  issue, and `Updates` will link the PR to it.

### Getting your Pull Request Reviewed, Approved, and Merged

Once a pull request has been opened, the following must take place before it is merged:

* It needs the `ok-to-merge` label
* It needs to pass all the checks.
* It needs to be approved by a [CODEOWNER](https://github.com/vmware-tanzu/tanzu-framework/blob/main/CODEOWNERS) for all files changed

While these steps will not always take place in the same order, the following describes the process for a typical pull request once it is opened:

1. Review is automatically requested from CODEOWNERS.
2. Assignee is added to pull request to ensure it gets proper attention throughout the process.
   Typically one of the CODEOWNERS will assign themselves, but they may choose to delegate to someone else.
3. Triage removes adds `ok-to-merge` if the pull request is generally aligned with product goals and does not conflict with current milestones; otherwise they may add a comment and a `do-not-merge/*` label.
4. Assignee may request others to do an initial review; anyone else may review
5. Reviewers leave feedback
6. Contributor updates pull request to address feedback
7. Requested reviewer approves pull request
8. Assignee approves pull request
9. Assignee merges pull request or requests another member to merge it if necessary.

During the review process itself, direct discussion among contributors and reviewers is encouraged.

Throughout the process, and until the pull request has been merged, the following should be transparent to the contributor:

* Has the pull request been assigned to anyone yet?
* Has the pull request been labeled with `ok-to-merge` or `do-not-merge/*`?
* Has someone been requested to review the pull request?
* Has the PR been approved by a reviewer?
* Has the PR been approved by the approver?

If any of the above is unclear, and there has been no new activity for 2-3 days,
the contributor is encouraged to seek further information by commenting and
mentioning the assignee or @vmware-tanzu/tanzu-framework-reviewers if there is
no assignee or they themselves are unresponsive.

### Merging a Pull Request

Maintainers should prefer to merge pull requests with the [Squash and merge](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-request-merges#squash-and-merge-your-pull-request-commits) option.
This option is preferred for a number of reasons.
First, it causes GitHub to insert the pull request number in the commit subject
which makes it easier to track which PR changes landed in.
Second, a one-to-one correspondence between pull requests and commits makes it
easier to manage reverting changes.

At a maintainer's discretion, pull requests with multiple commits can be merged
with the [Rebase and merge](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-request-merges#rebase-and-merge-your-pull-request-commits)
option. Merging pull requests with multiple commits can make sense in cases
where a change involves code generation or mechanical changes that can be
cleanly separated from semantic changes. The maintainer should review commit
messages for each commit and make sure that each commit builds and passes
tests.

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

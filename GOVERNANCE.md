# Tanzu Framework Repository Governance
This document defines the project governance for Tanzu Framework, an open source project by VMware.

## Overview
Framework offers an open source repository that will serve multiple Tanzu products including Tanzu Community Edition,
the Tanzu Community Edition.

Tanzu Community Edition is an open source and free to use community-supported Kubernetes distribution.
Please refer to the [Tanzu Community Edition repository](https://github.com/vmware-tanzu/tce) for more information.

## Framework Repository
Framework exists in a single repository and is governed by VMware and maintained under the vmware-tanzu organization.
* Framework: The TKG hub. Multiple Tanzu products build atop our Framework. Includes APIs, shared libraries, and
the Tanzu CLI and tools for integration.

## Community
* Users: Members that consume Framework via any medium (Slack, GitHub, mailing lists, etc.).
* Contributors: Members who contribute to Framework through documentation, code reviews, responding to
issues, participation in proposal discussions, contributing code, etc.
* Maintainers: Framework leaders are current employees of VMware. They are responsible for the overall
health and direction of the project; final reviewers of PRs and responsible for releases. Some Maintainers are
responsible for one or more components within Framework codebase, acting as technical leads, product managers and
engineering managers for that component. Maintainers are expected to contribute code and documentation, review PRs
including ensuring quality of code, triage issues, proactively fix bugs, and perform maintenance tasks for these
components. If a maintainer leaves VMware, he/she will also leave the maintainer position. The CODEOWNERS file in
the root directory specifies the maintainer list of people or team (via alias) for the responsibilities to the code
in the tree structure.

## Proposal Process
One of the most important aspects in any open source community is the concept of proposals. All large changes to the
codebase and / or new features, including ones proposed by maintainers, should be preceded by a proposal in our
community repo. This process allows for all members of the community to weigh in on the concept (including the technical
details), share their comments and ideas, and offer to help. It also ensures that members are not duplicating work or
inadvertently stepping on toes by making large conflicting changes.

The project roadmap is defined by accepted proposals.

Proposals should cover the high-level objectives, use cases, and technical recommendations on how to implement. In general,
the community member(s) interested in implementing the proposal should be either deeply engaged in the proposal process or
be an author of the proposal.

The proposal should be documented as a separated markdown file and pushed to the [design folder](docs/design) via PR. The name of the file should follow the name pattern <short meaningful words joined by '-'>-design.md,
e.g: `restore-hooks-design.md`.

Use the [Proposal Template](docs/dev/_proposal.md) as a starting point.

## Proposal Lifecycle
The proposal PR follows the GitHub lifecycle of the PR to indicate its status:

* Open: Proposal is created and under review and discussion.
* Approved: Proposal has been reviewed and approved.
* Rejected: Proposal has been reviewed and rejected.
* Merged: Proposal has been approved and code is merged in the repo.
* Closed: Proposal has been finished by the lifecycle either Merged or Rejected.

### Lazy Consensus
To maintain velocity in a project, the concept of Lazy Consensus is practiced. Ideas and / or proposals should be shared by
maintainers via GitHub with the appropriate maintainer groups (e.g., @vmware-tanzu/core-maintainers) tagged. Out of respect
for other contributors, major changes should be listed in the [ROADMAP](ROADMAP.md) to centralize the direction of the project.
Author(s) of proposals, pull requests, issues, etc. will specify a time period of no less than five (5) working days for comment
and remain cognizant of popular observed world holidays.

Other maintainers may request additional time for review, but should avoid blocking progress and abstain from delaying progress
unless absolutely needed. The expectation is that blocking progress is accompanied by a guarantee to review and respond to the
relevant action(s) (proposals, PRs, issues, etc.) in short order. All pull requests need to be approved by two (2) maintainers.

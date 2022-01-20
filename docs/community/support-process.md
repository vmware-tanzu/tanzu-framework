# Framework Support Process

## Definitions

* **Framework member** - is a person who is an active contributor and with
  write access to the `tanzu-framework` repo.
* **Critical urgent issue** - is an issue that severely impacts the use of the
  software and has no workarounds. This kind of issue should be labeled with
  [severity-1](severity-definitions.md#severity-1) label.

## Weekly rotation

The Framework members use a weekly rotation to manage community support.
Each week, a different member is the point person for triaging the new issues
and guiding severity-1 issues to the best state to be tackled.

The point person is not expected to solve every critical(severity-1) issue or
be on-call 24x7. Instead, they will communicate expectations for the critical
urgent issues to the community and ensure the issues are in the best position
to be addressed.

The point person is not expected to be involved with normal non-critical
priority(other than severity-1) issues.

## Start of Week

Support schedule will be provided to ensure everyone knows when their support
week is occurring.

The schedule will consist of members who provide a week of their expertise to
ensure new issues receive the labeling they need.

They also work closely with the community to ensure the issue is properly
detailed and have steps for reproducing the issue, if appropriate.

Point people are responsible for ensuring they are active and managing the
issue backlog during their scheduled week.

## During the Week

The point person will monitor:

* New issues labeled with `needs-triage`.
* Currently, open issues labeled as `investigating` and `triage/needs-info`.
* Open issues labeled as `severity-1`.

### GitHub issue flow

Generally speaking, new GitHub issues will fall into one of several categories.
The point person will use the following process for each issue:

* Feature request
  * Label the issue with `kind/feature`.
  * Determine the area of the Framework the issue belongs to and add appropriate
    [area](https://github.com/vmware-tanzu/tanzu-framework/labels?q=area) label.
  * Remove `needs-triage` label.
* Bug
  * Label the issue with `kind/bug`.
  * Determine the area of the Framework the issue belongs to and add appropriate
    [area](https://github.com/vmware-tanzu/tanzu-framework/labels?q=area) label.
  * If the issue is critical urgent, it should be labeled as `severity-1`.
  * Remove `needs-triage` label.
* User question/problem that does not fall into one of the previous categories
  * Assign the issue to yourself.
  * When you start investigating/responding, label the issue with `investigating`.
  * Add context for both the user and future support people.
  * Use the `triage/needs-info` label to indicate an issue is waiting for
  information from the user. If you do not get a response in 20 days then close
  the issue with an appropriate comment.
  * If you resolve the issue, add the resolution as a comment on the issue and
  close it.
  * If the issue ends up being a feature request or a bug, update the labels
  and follow the appropriate process for it.

## End of Week

The point person will ensure all GitHub issues worked on during the week are
labeled with `investigating` and `triage/needs-info` (if appropriate), and have
updated comments, so the next person can pick them up.

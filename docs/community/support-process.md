# Framework Support Process

## Weekly rotation
The Framework members use a weekly rotation to manage community support. 
Each week, a different member is the point person for triaging the new issues
and guiding high priority Severity-1 issues to the best state to be tackled.

The point person is not expected to solve every [critical(severity-1)](severity-definitions.md#severity-1)
issue or be on-call 24x7. Instead, they will communicate expectations for the
critical urgent issues to the community and ensure the issues are in the best
position to be addressed.

The point person is not expected to be involved with normal non-critical 
priority(other than severity-1) issues. The critical urgent issues and any
issues that need discussion can be discussed in the weekly SIG meet.

## Start of Week
We will provide a support schedule to ensure everyone knows when their support 
week is occurring.

The schedule will consist of members who provide a week of their expertise to
ensure new issues receive the labeling they need.

They also work closely with the community to ensure the issue is properly 
detailed and have steps for reproducing the issue, if appropriate.

Point people are responsible for ensuring they are active and managing the 
issue backlog during their scheduled week.

## During the Week
We will monitor
* Framework Repository issues
  * New issues labeled with `needs-triage`.
  * Currently, open issues labeled as `investigating` and `triage/needs-info`.
  * Open issues labeled as `severity-1`.

### GitHub issue flow
Generally speaking, new GitHub issues will fall into one of several categories.
We use the following process for each:
* Feature request
  * Label the issue with `kind/feature`.
  * Remove `needs-triage` label and add `triage/accepted` label.
  * Determine the area of the Framework the issue belongs to and add appropriate area label.
* Enhancement request
  * Label the issue with `kind/enhancement`.
  * Remove `needs-triage` label and add `triage/accepted` label.
  * Determine the area of the Framework the issue belongs to and add appropriate area label.
* Bug
  * Label the issue with `kind/bug`.
  * Remove `needs-triage` label and add `triage/accepted`.
  * Determine the area of the Framework the issue belongs to and add appropriate area label.
  * If the issue is critical urgent, it should be labeled as `severity-1`.
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
We ensure all GitHub issues worked on during the week are labeled with 
`investigating` and `triage/needs-info` (if appropriate), and have updated
comments, so the next person can pick them up. The point person should discuss
the week's developments during the SIG meeting including critical urgent(severity-1) issues
and any important points.

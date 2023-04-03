Cherry-pick process
===================

- Know the target release branches your PR must be cherry-picked into
- Raise your PR against main branch and add label `cherry-pick/release-0.28` where `release-0.28` is branch name, where you'd like to cherry-pick this change
- Get your PR against main tested, reviewed and merged. Once merged, github action workflow will create cherry-pick PR against desired release branch.
- The workflow adds a comment on the PR about the status of cherry-pick workflow for success (comment the resulting cherry-pick PR), failure or cancelled (if workflow is cancelled).
- If there are merge conflicts while raising the cherry-pick PR, the workflow fails and adds respective comment on the merged PR. In this case, please raise the cherry-pick PR manually.

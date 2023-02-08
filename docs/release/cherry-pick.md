Cherry-pick process
===================

- Know the target release branches your PR must be cherry-picked into
- Raise your PR against main branch and add label `cherry-pick/release-0.28` where `release-0.28` is branch name, where you'd like to cherry-pick this change
- Get your PR against main tested, reviewed and merged. Once merged, github action workflow will create cherry-pick PR against desired release branch.
- You can specify multiple labels for multiple cherry-pick PRs, just add respective labels for example: `cherry-pick/release-0.28` `cherry-pick/release-0.29`
- The workflow expects no conflicts while raising cherry-pick PRs, otherwise you'll have to create the cherry-picks manually. Monitor the workflow in your original PR against main for success/failure.

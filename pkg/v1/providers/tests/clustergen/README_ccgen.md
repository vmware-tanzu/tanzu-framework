# Cluster-Class Cluster Generation Comparison System

## Introduction

The clustergen tool has been augmented with the ability to generate cluster
resources using the old (pure client-side ytt-processing) and the new
(clusterclass-based using dryrun functionality) from the same set of cluster
configuration file

```text
make clustergen
```

produces two sets of outputs for each set of cluster configuration cases,
cleanses and normalizes the yaml, then generates diffs between the two sets of
outputs. This is used to gauge the degree to compatibility of the
clusterclass-based generation compared to the old outputs as we iterate on the
clusterclass authoring.

Note: the tool requires a modified version of clusterctl CLI that includes
a 'alpha topology-dryrun' and 'alpha generate-normalized-topology' commands,
buildable with `make clustectl` from
<https://github.com/vuil/cluster-api/tree/ccgen>
The branch will be periodically updated to bring in any relevant changes
upstream.

To limit the set of test cases to run this tool with, the CASES envvar can be
set to a space delimited set of dddddddd.case names found in the testdata/
directory.

e.g. this picks the first 10 cases using the azure infra
export CASES=$(for i in `grep azure:v1. testdata/*.case | \
    cut -d: -f1 | uniq | cut -d/ -f7 | head -10`; do echo -n "$i "; done; echo)

TODO: more details to follow.

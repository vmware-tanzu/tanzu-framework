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
a 'alpha topology-dryrun' and a new 'alpha generate-normalized-topology'
commands, buildable with `make clustectl` from
<https://github.com/vuil/cluster-api/tree/cgen_113>
The branch will be periodically updated to bring in any relevant changes
upstream. The customizations allows for output that is more deterministic and
diff-friendly and compared against old outputs.

To limit the set of test cases to run this tool with, the CASES envvar can be
set to a space delimited set of dddddddd.case names found in the testdata/
directory.

e.g. this picks the first 10 cases using the azure infra
export CASES=$(for i in `grep azure:v1. testdata/*.case | \
    cut -d: -f1 | uniq | cut -d/ -f7 | head -10`; do echo -n "$i "; done; echo)

## Details on dry-run generation of ClusterClass-based configuration

Since the clustergen tests are expected to generate configuration entirely
client-side, we need a means to generate cluster resources that would be as
close an approximation of what the server will generate for a CC-based Cluster
configuration, without standing up the server at all!

To do this we need to do the following only during the running of clustergen:

1. access the cluster class definition and templates client side
2. simulate any server-side data injection into the Cluster object that any
   CC patches would need to access

Item 1 is done by adding the cconly/ directory into the list of locations used
generation of clusterclass resources locally via the ytt processor. The
processor accomplishes this via processing the clusterclass-tkg-(infra)-default.yaml
template.  Note that this is a reason why CC-specific files that are already
bundled into a Carvel package are still present in the providers/ directory)

Item 2 is accomplished via hard-coding some values for the Cluster's TKR_DATA
variable (in run_tests.sh)

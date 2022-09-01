# Cluster Generation Compliance Tests

## Introduction

The purpose of these tests is to verify that CLI / cluster template changes
produces the desired effect on the generated TKG cluster YAML docset.

This is done by comparing the generated output of each cluster creation attempt
in the dataset against their expected results.

The dataset corresponds to a collection of "tkg config cluster ...." commands
with various arguments.

A reasonably small dataset generated via the pairwise combinatorial testing
tool (PICT) provides good coverage of all supported configuration variables and
command line options across each provider version.

## Prerequisite

Install PyYAML:

```sh
pip3 install PyYAML
```

## Running tests

Run cluster generation tests

```sh
make generate-testcases cluster-generation-tests
```

This will build a test binary capable of generating cluster configuration. This
binary will then be used to produce cluster configurations using various test data
These outputs, saved in testdata/generatd, can now be review for
discrepancies.

## Running A/B comparison tests

The above test workflow is the basis for comparing cluster generation outputs
produced by different commits of the repository.
Suppose you have a local branch with changes rebased off the top of origin/main

from the top level directory of the repository, do:

```sh
export CLUSTERGEN_BASE=`git ref-parse origin/main`
make clustergen
```

This will:

- generate a set of test configurations using the latest model files in HEAD
- build two test binaries, one off HEAD and one off commit set to
${CLUSTERGEN_BASE}, and produce cluster configuration based on the set of test
onfigurations with each binary.
- produce a diff between the two sets of cluster configuration outputs

One can use this approach to review the effect of changes in one's local branch
on how cluster configurations are affected, and is usually more effective than
reviewing the outputs from ```make cluster-generation-tests``` directly.

## Checklist when making changes affecting user visible changes to cluster creation

These changes can be:

1. Updates to the cluster generation logic in pkg/v1/tkg (these should be uncommon)
1. Provider template additions/changes

### Dealing with cluster creation code and template changes

Under some circumstances to get proper coverage on these changes, an updated
(most likely expanded) set of test cases should be generated with
[PICT](https://github.com/microsoft/pict) along with these changes.

You can build pict from the github source, or download a copy from

1. [MacOS](https://storage.googleapis.com/clustergen-tools/pict/pict.darwin)
1. [Linux](https://storage.googleapis.com/clustergen-tools/pict/pict.linux)

```sh
make generate-testcases
```

if PICT tool is not in PATH,

```sh
make PICT=location_of_pict_binary generate-testcases
```

What test cases are generated is determined by the contents of the model
files in the param_models/ directory. See "Updating Parameter Model Files"
for more details.

### Updating Parameter Model Files

Full documentation on the format and capabilities of the model files are in
[the PICT docs](https://github.com/Microsoft/pict/blob/master/doc/pict.md).

To partition the test cases by provider type, a file per provider is provided
in the param_models/ directory. During "make generate-testcase" PICT processes
every file. The outputs of these runs are cleansed into a parsable CSV format.

Each line in the CSV files is turned into a DDDDDDDD.case test case file
containing the command line args and config values to use during the cluster
config generation. Some "special" values in the model files encodes specific
situations, namely:

```sh
--X: "NOTPROVIDED"
```

if a value is assign as NOTPROVIDED for a CLI arg --X, --X will not be provided
in the command line invocation.

```sh
VAR: "''"
```

if a value is assign as ''

```sh
VAR: ''
```

will be written to the tkg config file used in the cluster generation

```sh
VAR: "NA"
```

if a value is assign as NA, the VAR configuration value will not be present in
the tkg config file used in the cluster generation.

```sh
VAR: "80"
VAR: "443"
```

Note numeric values should be quoted in the model file as well.

### Additional notes/tips

Since ytt-based templates has a potential to make cross-version, even
cross-provider changes, pay special attention to whether the changes affected
the expected number of providers and versions.

### Balancing test coverage with the number of test cases

When introducing **optional** args/config variables, one should avoid modifying
the main set of param model files (those name cluster_(infratype).model), which
are responsible for the bulk of the test cases generated. The reasoning here is
that any working configuration among the test cases generated from these will
continue to work precisely because these new args/config variables are
optional. Outputs from these existing test cases will in fact highlight the
differences caused by the upcoming change handling these new settings using
default values because none are specified.  (We suggest reviewing these changes
first before proceeding to creating/modifying test cases.)

Next, new test cases can then be generated from more targeted model files (an
example of which is cluster_optional.model). The purpose of these files are to

1. introduce different values for the new options to be tested
1. keep variation in the configuration aspects that are already thoroughly
   covered by the test cases generated from the main model files to an absolute
   minimum
1. provide mostly legitimate values for all non-optional parameters for all
   target infrastructures if the changes being tested are infra-agnostic

Combine testing of multiple sets of optional args when possible to further
reduce the number of test cases while promoting the testing of cross-feature
interaction. When not possible, new model files can be used for other test
scenarios. The only requirement is that these files are placed in param_models/
and are named cluster_[a-zA-Z_]+.model

In some cases, when several config variables introduced have interdependencies,
constraint statements may have to be introduced in the model file to minimize
the likelihood of incompatible combinations of values being generated.

Using these separate model files has the effect of providing more targeted
testing of the new features without altering too many existing test cases or
generating too many new ones.

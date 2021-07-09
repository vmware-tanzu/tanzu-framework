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


## Running tests

Install PyYAML:
```
pip3 install PyYAML
```

Tests can be run using the top-level Makefile:

To run tests using the legacy templating mechanism:
```
make YTT=false cluster-generation-tests
```

To run tests using the ytt processor
```
make cluster-generation-tests
```

Each should exit with return code zero if generated yamls are found to be
functionally equivalent to those stored in the dataset


## Checklist when making changes affecting user visible changes to cluster creation

These changes can be:

1. Introduction/changes to CLI command-line arguments, config variables.
1. CLI logic to modify cluster yamls during creation (these should be uncommon)
1. Provider template additions/changes

### Dealing with Introduction/Changes to CLI command-line arguments, config variables.

To get proper coverage on these changes, an updated (most likely expanded) set
of test cases should be generated with PICT (https://github.com/microsoft/pict)

You can build pict from the github source, or download a copy from

1. https://build-artifactory.eng.vmware.com/artifactory/list/kscom-generic-local/staging/artifacts/pict/pict.darwin
2. https://build-artifactory.eng.vmware.com/artifactory/list/kscom-generic-local/staging/artifacts/pict/pict.linux

```
make generate-testcases
```

if PICT tool is not in PATH,

```
make PICT=location_of_pict_binary generate_testcases
```

What test cases are generated is determined by the contents of the model
files in the param_models/ directory. See "Updating Parameter Model Files"
for more details.

### Dealing with CLI logic / cluster template changes

For these changes, do the following:

```
1. make cluster-generation-tests
2. review the diffs reported
3a. if they are expected as part of the changes, commit said diff.
3b. else fix issue causing the discreptancies
```

repeat this sequence until "make cluster-generation-tests" passes

### Updating Parameter Model Files

Full documentation on the format and capabilities of the model files are in:
https://github.com/Microsoft/pict/blob/master/doc/pict.md

To partition the test cases by provider type, a file per provider is provided
in the param_models/ directory. During "make generate-testcase" PICT processes
every file. The outputs of these runs are cleansed into a parsable CSV format.

The lastgen folder stores the uncleansed output of the most recent runs, and
serves as a reference (PICT calls them seed files) when generating cases in the
future. They can potentially minimize the number of changes to the testcase
list, so make sure to incorporate them in any PR that regenerates new test
cases.

Note that, at least for now, the seed files are pretty useless in scenarios
where new (especially mandatory) options/args are introduced because pretty
much all test cases will have to be updated to accomodate them. Likewise, one
will likely see changes in terms of numerous testcase files getting deleted and
possibly even more new files added. This is due to the test case name being a
hash on the config/args used in the cluster creation the latter is likely
different now due to the arg/config additions.
In these situations, all test case additions and deletions, along with changes
to the expected output files, should be submitted to thoroughly update the test
dataset.

Each line in the CSV files is turned into a DDDDDDDD.case test case file
containing the command line args and config values to use during the cluster
config generation. Some "special" values in the model files encodes specific
situations, namely:

```
--X: "NOTPROVIDED"
```

if a value is assign as NOTPROVIDED for a CLI arg --X, --X will not be provided
in the command line invocation.

```
VAR: "''"
```

if a value is assign as ''
```
VAR: ''
```
will be written to the tkg config file used in the cluster generation


```
VAR: "NA"
```
if a value is assign as NA, the VAR configuration value will not be present in
the tkg config file used in the cluster generation.


### Additional notes/tips

Since ytt-based templates has a potential to make cross-version, even
cross-provider changes, pay special attention to whether the changes affected
the expected number of providers and versions.

When unintended differences are detected, it is possible that either the cli or
cli providers branches or both are out-of-date. Consider if a rebase can
address it.

Since this situation happens quite frequently, try to keep changes to the test
output files in tests/clustergen/testdata/expected/ in their own commit so they
can be discarded and regenerated easily after a rebase.

Also, it may be a good idea to run 'make cluster-generation-tests' without your
changes to ensure the test passes as it is supposed to.

### Balancing test coverage with the number of test cases.

Since testcases are named based on the hash of the cli args and config settings
used in the cluster generation, introducing any new command line option or
config variable will cause big changes to existing test case.

Thus, whenever possible, such as when making CLI changes involving new
**optional** args/config variables, one should avoid modifying the main set of
param model files (those name cluster_(infratype).model), which are responsible
for the bulk of the test cases generated. The reasoning here is that any working
configuration among the test cases generated from these will continue to work
precisely because these new args/config variables are optional. Outputs from
these existing test cases will in fact highlight the differences caused by the
upcoming change handling these new settings using default values because none
are specified.  (We suggest reviewing these changes first before proceeding to
creating/modifying test cases.)

Next, new test cases can then be generated from more targeted model files (an
example of which is cluster_optional.model). The purpose of these files are to

1. introduce different values for the new options to be tested
2. keep variation in the configuration aspects that are already thoroughly
   covered by the test cases generated from the main model files to an absolute
   minimum
3. provide mostly legitimate values for all non-optional parameters for all
   target infrastructures if the changes being tested are infra-agnostic

Combine testing of multiple sets of optional args when possible to further
reduce the number of test cases while promoting the testing of cross-feature
interaction. When not possible, new model files can be used for other test
scenarios. The only requirement is that these files are placed in param_models/
and are named cluster_[a-zA-Z_]+.model

Using these separate model files has the effect of providing more targeted
testing of the new features without altering too many existing test cases of or
generating too many new ones.

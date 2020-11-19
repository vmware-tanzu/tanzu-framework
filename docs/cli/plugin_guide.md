# Mission statement
The Tanzu CLI aims to provide a consistent experience for users of the Tanzu portfolio.
This document is intended to provide direction for for designing Tanzu CLI plugins
that adhere to the Tanzu CLI patterns.

# Overall Principles

Use [Tanzu First Pane of Glass: Shared Taxonomy](https://docs.google.com/document/d/1K8_p4Ve09AXkFCU5GJNpthYac4L1fSVwfEr5mz-HzYk/edit)
as the dictionary for the Tanzu CLI. That Taxonomy document is intended to be
the authoritative source of truth for the Tanzu taxonomy.

# Command Structure

The Tanzu CLI should follow the pattern of:

```
tanzu [global flags] noun [sub-noun] verb RESOURCE [flags]
```

* All commands belong to a "group" to help with discoverability. New commands
should be added to an existing group.
* Any nouns being added must exist in the Shared Taxonomy document
* Global flags are maintained by the Tanzu CLI governance group
* All commands and flags should have a description that adheres to the following:
  * Fits on an 80 character wide screen
  * Begins with a capital letter and does not end with a period
* * Compound words should be `-` delimited: `management-cluster` over `managementcluster`
* Adding new commands and subcommands must be reviewed by the Tanzu SIG CLI group

### Top Level Noun:
* Tread lightly when adding another top level noun
* Any new top level nouns must exist in the Shared Taxonomy dictionary
* Adding new top level nouns must be reviewed by the Tanzu SIG CLI group.

### Sub-noun:
* Sub-nouns do not need to be reviewed by the governance group
* Commands should not have more than one sub-noun

### Flags:
* A user should only be required to explicitly set 2 flags; most flags should be
 sanely defaulted
* Add as many flags as necessary configure the command
* Consider using a config file if the number of flags exceeds 5

### Verbs:
* Use the standard CRUD actions as much as possible
* create, delete, get, list, update

### Examples
* Any complex command should have examples demostrating its functionality.
```
# my example foo command
tanzu foo --bar
```

# Repositories

The core framework exists in https://github.com/vmware-tanzu-private/core any
plugins that are considered open source should exist in that repository as well.

Other repositories should follow the model seen in
https://gitlab.eng.vmware.com/olympus/cli-plugins and vendor the core repository.
Ideally these plugins should exist in the same area as the API definitions.

# Alpha Commands

If you want to add functionality that is backed by an alpha or unstable API, you
can expose it in the CLI via a top level `tanzu alpha` command.

# CLI Behavior

### Components
CLI commands should utilize the plugin component library in `pkg/cli/component` for interactive features like prompts or table printing.

### Templates
TBD 

### Asynchronous Requests
Commands should be written in such a way as to return as quickly as possible.
When a request is not expected to return immediately,
then the command should return immediately with an exit code indicating the
server's response, and rely on the user to poll
the resource via a `get` command.

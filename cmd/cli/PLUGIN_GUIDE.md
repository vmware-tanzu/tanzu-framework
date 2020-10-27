Use [Tanzu First Pane of Glass: Shared Taxonomy](https://docs.google.com/document/d/1K8_p4Ve09AXkFCU5GJNpthYac4L1fSVwfEr5mz-HzYk/edit) as the dictionary
for the Tanzu CLI. That Taxonomy document is intended to be the authoritative source of truth for the Tanzu taxonomy.

# Principles

The Tanzu CLI should follow the pattern of:

```
tanzu [global flags] noun [sub-noun] verb RESOURCE [flags]
```

* Any nouns being added must exist in the Shared Taxonomy document
* Global flags are maintained by the Tanzu CLI governance group

### Top Level Noun:
* Tread lightly when adding another top level noun
* Any new top level nouns must exist in the Shared Taxonomy dictionary
* Adding new top level nouns must be reviewed by the Tanzu CLI “governance group”

### Sub-noun:
* Sub-nouns do not need to be reviewed by the governance group
* Commands should not have more than one sub-noun 

### Flags:
* A user should only be required to explicitly set 2 flags; most flags should be sanely defaulted
* Add as many flags as necessary configure the command
* Consider using a config file if the number of flags exceeds 5

### Verbs:
* Use the standard CRUD actions as much as possible
* create, delete, get, list, update

# Repositories

The core framework exists in https://github.com/vmware-tanzu-private/core any plugins that are considered open source should exist in that repository as well.

Other repositories should follow the model seen in https://gitlab.eng.vmware.com/olympus/cli-plugins and vendor the core repository. Ideally these plugins should exist in the same area as the API definitions. 

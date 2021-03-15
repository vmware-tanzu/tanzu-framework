# Tanzu CLI Styleguide

## Tanzu CLI mission statement
The Tanzu CLI aims to provide a consistent experience for users of the Tanzu portfolio

This document is intended to provide direction for designing Tanzu CLI plugins so
that they adhere to established patterns

------------------------------

## Design Principles

### Prioritize human users first
* Default to human readable output, but support plaintext and json/yaml options
* Use straightforward, simple language and syntax
* Use machine readable output where it does not impact usability


### Prioritize discoverability, consistency and predictability
* Consistent language, structure, and as much as possible, behaviors across interfaces 

### Be as helpful as possible
* Invest time in your help and error text
* Provide examples
* Command line completion should be used whenever possible
* Provide context info whenever possible (ie feedback messages, etc)
 
### The Tanzu CLI should be declarative whenever possible
* Most APIs should be made declarative, and the CLI commands should supply basic `apply` operations on those data models

### The CLI should contain as little logic as possible
* Avoid complex client logic, maintaining state; prefer server side logic

### The CLI must be accessible 
* We should not exclude users by deprioritizing accessibility
* Features like search filters and output limits are useful for folks using assistive technology
* Screen-readers and automation are both served by being thoughtful with the use of non machine-readable characters (emojis, ascii spinners, etc.)

------------------------------


## Designing commands

The Tanzu CLI uses the following pattern

```
tanzu [global flags] noun [sub-noun] verb RESOURCE [flags]
```

Example
```
tanzu cluster create NAME
tanzu management-cluster kubeconfig get --admin
```

### Global Flags
Global flags are maintained by the [Tanzu CLI SIG](https://github.com/vmware-tanzu-private/community/tree/main/sigs/api-cli) and adding to the current global set should be managed through the SIG

### Nouns 
Any nouns being added must exist in the Shared [Taxonomy document](https://github.com/vmware-tanzu-private/core/blob/main/hack/linter/cli-wordlist.yml)
* Introducing nouns to support the creation of new commands and/or subcommands should be reviewed by the [Tanzu CLI SIG](https://github.com/vmware-tanzu-private/community/tree/main/sigs/api-cli)

Compound words should be - delimited
```
 management-cluster not managementcluster nor managementCluster
```

Nouns should not be combined with verbs
```
tanzu app get, not tanzu app-get
```

### Verbs
Use the standard CRUD verbs whenever possible
* Tanzu CLI uses create, delete, get, list, update
If at all possible, use verbs from the [command reference list](link)
* New verbs must be reviewed by the Tanzu CLI SIG
Opposing commands should take the form of antonyms 

Example
```
‘assign quota’ and ‘unassign quota’
```

### Sub-Noun
Plugin specific sub-nouns do not need to be reviewed by the governance group
Please review the command reference list when using sub-nouns, to make sure your word is not already in use

### Resource 
Commands should not nest more than one layer of resources

### Positional Arguments 
* There should no more than 1 positional argument
* Ideally, the positional argument should be the name/subject of the thing the command refers to (not some other flaggable property)

Example
```
tanzu cluster create CLUSTER-NAME [flags]
```

### Flags 
* Use standard names for flags if there is one (flags used in the cli are documented here)
* Where possible, set reasonable defaults for flag-able options that align with expected workflows
* A user should only be required to explicitly set a max of 2 flags 
* Add as many flags as necessary to configure the command
* Consider supporting the use of a config file if the number of flags exceeds 5
* Flags should be tab completed. The Tanzu CLI uses the cobra framework, which has tooling to help with this [Cobra shell completion docs](https://github.com/spf13/cobra/blob/master/shell_completions.md)

### Prompting
* If you are prompting for a secret, do not echo it in the terminal when the user types
  * Commands should accept secrets via environment variables
*  Commands should not prompt when TTY is not present
  * The component library can provide this check, pending resolution of issue #330

### Command principals
* Support verbosity flags
* Provide a way to disable interactive prompting (--quiet, --force, --yes)

### Components - Command
Available for plugins written in golang
CLI commands should utilize the plugin component library in [pkg/cli/component](https://github.com/vmware-tanzu-private/core/tree/main/pkg/v1/cli/component) for interactive features like prompts or table printing
Available input components
* Prompt
* Select 

------------------------------


## Designing feedback

### Feedback principals
Useful for everyone, critical to the usability of screen-readers and automation
Support verbosity flags (-v, --verbose)
--format to define preferred output
* Useful for humans and machine users, ie table, value, json, csv, yaml, etc
--filter
* To refine the output criteria 
For list commands, consider supporting --limit to set a maximum number of responses
* --page-size to define the number of resources per page if using pagination
* --sort-by to specify what field to organize the list by
If using color to communicate information, don't let it be the only indicator

### Confirmation prompting
* To protect against unintended destructive actions, the CLI should confirm user intentions before executing commands such as ‘delete’
* It can be helpful to describe the impact of an action if not obvious
* The default action should be to cancel if no input is provided

Example
```
$ tanzu cluster delete MY_CLUSTER

This action impacts all apps in this cluster. Deleting the cluster will remove associated apps
Really delete the cluster MY_CLUSTER [y/N]
MY_CLUSTER has not been deleted
```

* To disable all interactive prompts when running cli commands (helpful for scripting purposes) the CLI should support a --yes or -y, option

Example
```
tanzu cluster delete MY_CLUSTER --yes
```

### Confirmation feedback
When executing a command, repeat back the command and context to the user  

Example
```
$ tanzu app create NAME
Creating app NAME in namespace NAMESPACE as user USERNAME  
```

### Progress reporting
For long running processes, don't go for a long period without output to the user. Outputting something like ’Processing…’, or a spinner can go a long way towards reassuring the user that their command went through   

### Warnings
In the confirmation feedback, include a notice for experimental or beta commands  

Example
```
“This is a EXPERIMENTAL command and may change without notice.”
“This is a BETA command and may change without notice.”
```
------------------------------


## Designing outputs

### stdout
* Primary output for both human and machine users  

### stderr
* Output for messages, warnings, errors.
* Stderr output is shown to the user, but not included when piping commands together  

### Exit codes
* When a command succeeds without error, the process should exit with code 0  
* When a command fails for any reason (invalid flags, failed API request, etc.), the process should exit with code 1  
* When a multi-step command fails at any step, the process should exit with code 1  
* When an operation is meant to be idempotent (like creating a resource  that already exists, or deleting a resource that’s already been deleted), the process should exit with code 0  


### Feedback when a process completes
### For asynchronous commands 
  * Exit 0 - the command was successful issued AND
  * Return a line telling the user how to check status of the command

### For synchronous commands 
  * Feedback could be an error message with exit code 1 OR
  * Confirmation of completion (i.e. “OK”, “done”) with exit code 0  

### Tables
* Tables are the default output format for list commands
* Column headers are upper cased  
* Columns are separated by three spaces from the longest key/value
* If there aren’t any  values for a column, the column heading should be displayed and values  left empty - don’t conditionally remove or hide a column
* If there are no values for a table, the table is replaced with a message, like ‘No apps found’

Example
```
NAME             STATUS   CPU   MEM   NAMESPACE            WORKSPACE
calculator-app   running  53%   18%   mortgage-calc-dev     
calculator-bpp   running  93%   21%   mortgage-calc-test     
calculator-cpp   running  23%   77%   mortgage-calc  
```

### Time format
The Tanzu CLI, like kubectl uses the ISO 8601 standard for date and time
* Compact, consistent width, app/server's timezone

Example
```  
2021-03-02T15:43:12.41-0700
```

### Key:value pairs
* Key and value always lowercase
* Key followed by a colon (:)
* Values  left-aligned 
* Value column separated by three spaces from the longest key
* If no value for a key, key is displayed, value left empty 

Example
```
name:        morgage-calculator-app
namespace:   morgage-cal-dev
workspace:   mort-calc-dev
status:
url:         http://myapplicationurl.com  
```

### Color
*  Colors can be disabled using an environment variable (NO_COLOR=TRUE)
*  Colors are always disabled when the session is not a TTY session. This allows for the piping of CLI output into other commands (e.g. grep) or machine reading without including stray color characters (pending issue #369)
*  Usage tips are always in plain text, even when referencing text that might normally be colorized
```
TODO Examples of:
help text including command example 
error text describing next steps/commands
success message suggesting next steps
```
* Warning and error *messages* are in plain text
```
TODO - add screenshot
```

#### Don't add color to anything outside of the following conventions to convey contex:

*  Red = warning, danger  
Warnings and error messages are colorized and bold
```
TODO - Add screenshot
```

*  Green = success, informational  
Confirmation of completion when a command runs is colorized and bold.
```
TODO- add screenshot
```

*  Cyan = stability, calm, informational  
In command feedback: resources, and user name is colorized and bold
Interactive prompting: user input is colorized, as is the preceding question mark.

```
TODO- add screenshot
```
```
TODO- add screenshot
```


### Animation
* Disable if stdout is not an interactive terminal
  * The component library can provide this check pending resolution of issue #369

### Symbols / Emojis
* Currently no standards or guidance
* Recommendation is to discuss plans for emoji/symbol use with SIG
* Disable if stdout is not an interactive terminal
  * The component library can provide this check pending resolution of issue #369

### Components - Output
* Available for plugins written in golang
CLI commands should utilize the plugin component library in pkg/cli/component for interactive features like prompts or table printing.
Available output components
  * Table

------------------------------


## Designing help text
A command will display help if passed -h, --help, or if no options are passed and a command expects them

Provide a support path for feedback and issues 
* A github link or website URL in the top level help encourage user feedback  

All commands and flags should have description text that
* Fits on an 80 character wide screen, to prevent word wrap
* Begins with a capital letter and does not end with a period

Any complex command should have examples demonstrating its functionality  

Example
```
$ tanzu login -h
# Login to TKG management cluster using endpoint
	tanzu login --endpoint "https://login.example.com"  --name mgmt-cluster

# Login to TKG management cluster by using kubeconfig path and context for the management cluster
	tanzu login --kubeconfig path/to/kubeconfig --context path/to/context --name mgmt-cluster

# Login to an existing server
	tanzu login --server mgmt-cluster
```
------------------------------


## Designing error and warning text
Use output warnings sparingly. When used often they can create a lot of noise and users may learn to ignore them 

Try to write clear and concise errors, which tell the user what action (if any) must be taken. Ideally users won’t have to ask for help or check documentation to recover 

Make errors informative. A great error message should contain the following
* Error code / Error title, if applicable
* Error description - what happened 
* How to fix the error
* URL to documentation for user’s specific error, if applicable

Example
```
$ myapp dump -o myfile.out
Error: EPERM - Invalid permissions on myfile.out
Cannot write to myfile.out, file does not have write permissions
Fix with: chmod +w myfile.out
https://github.com/tanzu/core/help/2323
```
```
$ Tanzu namespace get EXAMPLE
Error: Namespace EXAMPLE not found. Try 'tanzu namespace list' to see available options
```

Use context in error messages to ease recovery
* If a parameter is invalid or missing, it is a chance to be helpful by telling the user exactly what they missed 
```
EXAMPLE “You forgot to enter the --name, apps in this namespace include App1, App2, etc...”
```

Make it easy to submit bug reports and feedback
* Consider including a github URL, or link to a form in error or help text
* If possible pre-populate form information to make the report as useful while not burdening the user


------------------------------

### Plugins

For information about developing plugins, see the [Plugin Guide](https://github.com/vmware-tanzu-private/core/blob/main/docs/cli/plugin_guide.md)

### Contributions to the Style Guide

A styleguide is never done, and should change to meet the changing needs
To propose changes please create an issue, and add it to the CLI SIG agenda to discuss

------------------------------

### Precedent
Olympus Design System   
PKS styleguide
cf-cli styleguide


### Accessibility Guidelines
* Web Content Accessibility Guidelines [WCAG](https://www.w3.org/TR/2008/REC-WCAG20-20081211/)
* US section 508 [link](https://www.section508.gov/)

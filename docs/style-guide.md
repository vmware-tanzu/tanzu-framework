# Documentation Style Guide

## Table of Contents

* [Content best practices](#content-best-practices)
    * [Use present tense](#use-present-tense)
    * [Use active voice](#use-active-voice)
    * [Use simple and direct language](#use-simple-and-direct-language)
    * [Address the reader as "you"](#address-the-reader-as-you)
    * [Avoid Latin phrases](#avoid-latin-phrases)
* [Patterns to avoid](#patterns-to-avoid)
    * [Avoid using "we"](#avoid-using-we)
    * [Avoid jargon and idioms](#avoid-jargon-and-idioms)
    * [Avoid statements about the future or that will soon be out of date](#avoid-statements-about-the-future-or-that-will-soon-be-out-of-date)
    * [Language](#language)
* [Documentation formatting standards](#documentation-formatting-standards)
    * [Use camel case for API objects](#use-camel-case-for-api-objects)
    * [Use angle brackets for placeholders](#use-angle-brackets-for-placeholders)
    * [Use bold for user interface elements](#use-bold-for-user-interface-elements)
    * [Use italics to define or introduce new terms](#use-italics-to-define-or-introduce-new-terms)
    * [Use code style for filenames, directories, paths, object field names and namespaces](#use-code-style-for-filenames-directories-paths-object-field-names-and-namespaces)
    * [Use punctuation inside quotes](#use-punctuation-inside-quotes)
* [Inline code formatting](#inline-code-formatting)
    * [Use code style for inline code and commands](#use-code-style-for-inline-code-and-commands)
    * [Starting a sentence with a component tool or component name](#starting-a-sentence-with-a-component-tool-or-component-name)
    * [Use normal style for string and integer field values](#use-normal-style-for-string-and-integer-field-values)
* [Code snippet formatting](#code-snippet-formatting)
    * [Don't include the command prompt](#dont-include-the-command-prompt)
    * [Separate commands from output](#separate-commands-from-output)
* [Tanzu word list](#tanzu-word-list)
* [Markdown elements](#markdown-elements)
    * [Headings](#headings)
    * [Paragraphs](#paragraphs)
    * [Links](#links)
    * [Lists](#lists)
    * [Tables](#tables)

_This style guide is adapted from the [Kubernetes documentation style guide](https://kubernetes.io/docs/contribute/style/style-guide/)._

This page outlines writing style guidelines for documentation related to this repo. Use this guide as a reference when
writing documents. Note that these are guidelines, not rules. Use your best judgement when writing and feel free to
propose changes to these guidelines by creating PRs.

## Content best practices
### Use present tense

|Do|Don't|
|--- |--- |
|This `command` starts a proxy.|This command will start a proxy.|

Exception: Use future or past tense if it is required to convey the correct meaning.

### Use active voice

|Do|Don't|
|--- |--- |
|You can explore the API using a browser.|The API can be explored using a browser.|
|The YAML file specifies the replica count.|The replica count is specified in the YAML file.|

Exception: Use passive voice if active voice leads to an awkward sentence construction.

### Use simple and direct language

Use simple and direct language. Avoid using unnecessary phrases, such as saying "please."

|Do|Don't|
|--- |--- |
|To create a ReplicaSet, ...|In order to create a ReplicaSet, ...|
|See the configuration file.|Please see the configuration file.|
|View the Pods.|With this next command, we'll view the Pods.|

### Address the reader as "you"

|Do|Don't|
|--- |--- |
|You can create a Deployment by ...|We'll create a Deployment by ...|
|In the preceding output, you can see...|In the preceding output, we can see ...|

### Avoid Latin phrases

Prefer English terms over Latin abbreviations.

|Do|Don't|
|--- |--- |
|For example, ...|e.g., ...|
|That is, ...|i.e., ...|

Exception: Use "etc." for et cetera.

## Patterns to avoid

### Avoid using "we"

Using "we" in a sentence can be confusing, because the reader might not know
whether they're part of the "we" you're describing.

|Do|Don't|
|--- |--- |
|Version 1.4 includes ...|In version 1.4, we have added ...|
|Kubernetes provides a new feature for ...|We provide a new feature ...|
|This page teaches you how to use Pods.|In this page, we are going to learn about Pods.|

### Avoid jargon and idioms

Many readers speak English as a second language. Avoid jargon and idioms to help them understand better.

|Do|Don't|
|--- |--- |
|Internally, ...|Under the hood, ...|
|Create a new cluster.|Turn up a new cluster.|

### Avoid statements about the future or that will soon be out of date

Avoid making promises or giving hints about the future. If you need to talk about
a beta feature, put the text under a heading that identifies it as beta
information.

Also avoid words like “recently”, "currently" and "new." A feature that is new today might not be
considered new in a few months.

|Do|Don't|
|--- |--- |
|In version 1.4, ...|In the current version, ...|
|The Federation feature provides ...|The new Federation feature provides ...|

### Language

This documentation uses U.S. English spelling and grammar.

## Documentation formatting standards

### Use camel case for API objects

When you refer to an API object, use the same uppercase and lowercase letters
that are used in the actual object name. Typically, the names of API
objects use
[camel case](https://en.wikipedia.org/wiki/Camel_case).

Don't split the API object name into separate words. For example, use
PodTemplateList, not Pod Template List.

Refer to API objects without saying "object," unless omitting "object"
leads to an awkward sentence construction.

|Do|Don't|
|--- |--- |
|The Pod has two containers.|The pod has two containers.|
|The Deployment is responsible for ...|The Deployment object is responsible for ...|
|A PodList is a list of Pods.|A Pod List is a list of pods.|
|The two ContainerPorts ...|The two ContainerPort objects ...|
|The two ContainerStateTerminated objects ...|The two ContainerStateTerminateds ...|

### Use angle brackets for placeholders

Use angle brackets for placeholders. Tell the reader what a placeholder represents.

1. Display information about a Pod:

        kubectl describe pod <pod-name> -n <namespace>

   If the pod is in the default namespace, you can omit the '-n' parameter.

### Use bold for user interface elements

|Do|Don't|
|--- |--- |
|Click **Fork**.|Click "Fork".|
|Select **Other**.|Select "Other".|

### Use italics to define or introduce new terms

|Do|Don't|
|--- |--- |
|A _cluster_ is a set of nodes ...|A "cluster" is a set of nodes ...|
|These components form the _control plane_.|These components form the **control plane**.|

### Use code style for filenames, directories, paths, object field names and namespaces

|Do|Don't|
|--- |--- |
|Open the `envars.yaml` file.|Open the envars.yaml file.|
|Go to the `/docs/tutorials` directory.|Go to the /docs/tutorials directory.|
|Open the `/_data/concepts.yaml` file.|Open the /\_data/concepts.yaml file.|


### Use punctuation inside quotes

|Do|Don't|
|--- |--- |
|events are recorded with an associated "stage."|events are recorded with an associated "stage".|
|The copy is called a "fork."|The copy is called a "fork".|

Exception: When the quoted word is a user input.

Example:
* My user ID is “IM47g”.
* Did you try the password “mycatisawesome”?

## Inline code formatting

### Use code style for inline code and commands

For inline code in an HTML document, use the `<code>` tag. In a Markdown
document, use the backtick (`` ` ``).

|Do|Don't|
|--- |--- |
|The `kubectl run` command creates a Deployment.|The "kubectl run" command creates a Deployment.|
|For declarative management, use `kubectl apply`.|For declarative management, use "kubectl apply".|
|Use single backticks to enclose inline code. For example, `var example = true`.|Use two asterisks (`**`) or an underscore (`_`) to enclose inline code. For example, **var example = true**.|
|Use triple backticks (\`\`\`) before and after a multi-line block of code for fenced code blocks.|Use multi-line blocks of code to create diagrams, flowcharts, or other illustrations.|
|Use meaningful variable names that have a context.|Use variable names such as 'foo','bar', and 'baz' that are not meaningful and lack context.|
|Remove trailing spaces in the code.|Add trailing spaces in the code, where these are important, because a screen reader will read out the spaces as well.|

### Starting a sentence with a component tool or component name

|Do|Don't|
|--- |--- |
|The `kubeadm` tool bootstraps and provisions machines in a cluster.|`kubeadm` tool bootstraps and provisions machines in a cluster.|
|The kube-scheduler is the default scheduler for Kubernetes.|kube-scheduler is the default scheduler for Kubernetes.|

### Use normal style for string and integer field values

For field values of type string or integer, use normal style without quotation marks.

|Do|Don't|
|--- |--- |
|Set the value of `imagePullPolicy` to `Always`.|Set the value of `imagePullPolicy` to "Always".|
|Set the value of `image` to `nginx:1.16`.|Set the value of `image` to nginx:1.16.|
|Set the value of the `replicas` field to `2`.|Set the value of the `replicas` field to 2.|

## Code snippet formatting

### Don't include the command prompt

|Do|Don't|
|--- |--- |
|kubectl get pods|$ kubectl get pods|

### Separate commands from output

Verify that the Pod is running on your chosen node:

```
kubectl get pods --output=wide
```

The output is similar to this:

```
NAME     READY     STATUS    RESTARTS   AGE    IP           NODE
nginx    1/1       Running   0          13s    10.200.0.4   worker0
```

## Tanzu word list

A list of Tanzu-specific terms and words to be used consistently across the repo.

|Term|Usage|
|--- |--- |
|Kubernetes|Kubernetes should always be capitalized.|
|Docker|Docker should always be capitalized.|
|Tanzu|Tanzu should always be capitalized.|
|Tanzu Kubernetes Grid|Tanzu Kubernetes Grid should always be capitalized.|
|TKG|TKG should always be capitalized.|
|CLI|CLI should always be capitalized.|
|VMware|VMware should always be correctly capitalized.|
|On-premises|On-premises or on-prem rather than on-premise or other variations.|
|Plugin|Plugin rather than plug-in or other variations.|

## Markdown elements

### Headings
People accessing this documentation may use a screen reader or other assistive technology (AT). [Screen readers](https://en.wikipedia.org/wiki/Screen_reader) are linear output devices, they output items on a page one at a time. If there is a lot of content on a page, you can use headings to give the page an internal structure. A good page structure helps all readers to easily navigate the page or filter topics of interest.

|Do|Don't|
|--- |--- |
|Include a title on each page or blog post.|Include more than one title headings (#) in a page.|
|Use ordered headings to provide a meaningful high-level outline of your content.|Use headings level 4 through 6, unless it is absolutely necessary. If your content is that detailed, it may need to be broken into separate articles.|
|Use sentence case for headings. For example, **Extend kubectl with plugins**|Use title case for headings. For example, **Extend Kubectl With Plugins**|

### Paragraphs

|Do|Don't|
|--- |--- |
|Try to keep paragraphs under 6 sentences.|Write long-winded paragraphs.|
|Use three hyphens (`---`) to create a horizontal rule for breaks in paragraph content.|Use horizontal rules for decoration.|

### Links

|Do|Don't|
|--- |--- |
|Write hyperlinks that give you context for the content they link to. For example: Certain ports are open on your machines. See [check required ports](#check-required-ports) for more details.|Use ambiguous terms such as “click here”. For example: Certain ports are open on your machines. See [here](#check-required-ports) for more details.|
|Write Markdown-style links: `[link text](URL)`. For example: `[TKG documentation](https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/index.html)` and the output is  [TKG documentation](https://docs.vmware.com/en/VMware-Tanzu-Kubernetes-Grid/index.html).|Write HTML-style links: `Visit our tutorial!`|


### Lists

Group items in a list that are related to each other and need to appear in a specific order or to indicate a correlation between multiple items. When a screen reader comes across a list—whether it is an ordered or unordered list—it will be announced to the user that there is a group of list items. The user can then use the arrow keys to move up and down between the various items in the list.

- End each item in a list with a period if one or more items in the list are complete sentences. For the sake of consistency, normally either all items or none should be complete sentences.

- Ordered lists that are part of an incomplete introductory sentence can be in lowercase and punctuated as if each item was a part of the introductory sentence.

- Use the number one (`1.`) for ordered lists.

- Use (`+`), (`*`), or (`-`) for unordered lists - be consistent within the same document.

- Leave a blank line after each list.

- Indent nested lists with four spaces (for example, ⋅⋅⋅⋅).

- List items may consist of multiple paragraphs. Each subsequent paragraph in a list item must be indented by either four spaces or one tab.

### Tables

The semantic purpose of a data table is to present tabular data. If you need to create a table, create the table in
markdown.

```
Parameter | Description | Default
:---------|:------------|:-------
`timeout` | The timeout for requests | `30s`
`logLevel` | The log level for log output | `INFO`
```

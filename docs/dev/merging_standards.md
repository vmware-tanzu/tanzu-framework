# Merging Standards

Many aspects of Framework are intended to be standardized - things like CLI flags, log output and terminology are all intended to be standard to ensure cohesion across the product.

This document intends to layout steps on how to consider Pull Requests which may alter these areas.

## CLI Words

We have standard [CLI words](../../pkg/v1/cli/command/plugin/lint/cli-wordlist.yml) defined in this repo to make it simple to keep our CLI words intact and allow our CLI plugins to vet themselves.

The easiest way to check if a command aligns with our CLI nouns and verbs standards is to run the plugins hidden `lint` comand. Every plugin offered by Tanzu CLI can lint itself against our standard words list. For any plugin, simply run:

`tanzu <plugin> lint`

Alternatively, the list can always just [be referenced directly](https://github.com/vmware-tanzu/tanzu-framework/blob/main/hack/linter/cli-wordlist.yml).

### Format

The [CLI style-guide](docs/cli/style_guide.md) provides guidance on how CLI nouns and verbs should be structured, please make sure to consult this while evaluating CLI word changes.

In short, words meet the following criteria:

* English, using US spelling
* Clear and unambiguous
* Concise without using acronyms that are not an industry standard (URL is good, MC is not)

Ideally, words will be the appropriate part of speech; nouns should be nouns, verbs should be verbs.

Ideally, merge requests proposing new words will define those terms.

Any additions that contradict existing terms should be adjusted and changes which do not conform to this list always require input from *Product*. This list is parsed and used to enforce standardization so it is important to treat changes here deliberately, like changing a model.

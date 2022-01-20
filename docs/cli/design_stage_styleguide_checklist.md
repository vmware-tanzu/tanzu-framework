# Design-Stage Styleguide Checklist

Follow this checklist to design commands that are consistent with the [CLI Style Guide](style_guide.md).

Follow the [Build-Stage Checklist](build_stage_styleguide_checklist.md) when implementing in code.

## Command Structure

- [ ] **Do the commands follow the pattern described in the [CLI Style Guide](style_guide.md#designing-commands)?**  (importance: high)
  - noun - verb - resource - flags
- [ ] **Is the number of nested noun layers <= 2?** (importance: medium)
- [ ] **Is the number of required flags <= 2?** (importance: medium)
- [ ] **Is the number of optional flags <= 5?** (importance: medium)

## UI Text / Taxonomy

- [ ] **Do any commands require adding new nouns (resources) to the [existing taxonomy](/hack/linter/cli-wordlist.yml)?** (importance: high)
  - Check in with the CLI SIG to add new top-level resources
- [ ] **Are the nouns in each command used in a manner consistent with usage in existing commands?** (importance: high)
- [ ] **Do any commands require adding new verbs (actions) to the [existing taxonomy](/hack/linter/cli-wordlist.yml)?**    (importance: medium)
  - Not a problem, but if there is a verb in the [existing taxonomy](/hack/linter/cli-wordlist.yml) that could be used, please use it
- [ ] **Are the verbs in each command used in a manner consistent with usage in existing commands?** (importance: medium)
- [ ] **Do any commands require adding new flags to the [existing taxonomy](/hack/linter/cli-wordlist.yml)?** (importance: low)
  - Not a problem, but if there is a flag in the [existing taxonomy](/hack/linter/cli-wordlist.yml) that could be used, please use it

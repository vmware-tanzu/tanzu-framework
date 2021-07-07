# codegen

Tanzu codegen plugin is a place for generators for generating utility code and
Kubernetes YAML, and is based on controller-tools. The generators specify
what to generate and output rules specify where to write the results.

## Generators
Generators look for special marker comments in the Go code and generate the
utility code and config YAML. To use a specific generator it needs to be 
specified through the CLI option and multiple generators can be used when 
running the `tanzu codegen generate` command. The generated utility code 
or the YAML needs to be written somewhere, this is controlled by the output 
rules. To learn more about output rules and how to configure them check 
[here](https://master.book.kubebuilder.io/reference/controller-gen.html#output-rules)

### Feature
Feature generator is for generating Feature CRs from marker comments that are 
associated with a type declaration.

Command to use the Feature generator:

```
tanzu codegen generate paths=${path_to_scan} feature output:feature:artifacts:config=${outputDir}
```
## tanzu management-cluster completion zsh

generate the autocompletion script for zsh

### Synopsis


Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for every new session, execute once:
# Linux:
$ management-cluster completion zsh > "${fpath[1]}/_management-cluster"
# macOS:
$ management-cluster completion zsh > /usr/local/share/zsh/site-functions/_management-cluster

You will need to start a new shell for this setup to take effect.


```
tanzu management-cluster completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --log-file string   Log file path
  -v, --verbose int32     Number for the log level verbosity(0-9)
```

### SEE ALSO

* [tanzu management-cluster completion](tanzu_management-cluster_completion.md)	 - generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 5-Nov-2021

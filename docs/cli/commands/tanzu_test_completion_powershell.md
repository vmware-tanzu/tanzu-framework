## tanzu test completion powershell

generate the autocompletion script for powershell

### Synopsis


Generate the autocompletion script for powershell.

To load completions in your current shell session:
PS C:\> test completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
tanzu test completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [tanzu test completion](tanzu_test_completion.md)	 - generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 5-Nov-2021

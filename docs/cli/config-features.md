# Configuration of Features

The Tanzu CLI offers the ability to configure CLI plugins based on feature flags. Some feature flags control behavior options that are part of
a production-ready release. By default, the CLI will install with only shippable features. 
However, some beta features may be available by updating config options. 
To use the feature configuration, you need to know the name of the plugin and the feature flag for the functionality you want to activate (or deactivate). 
For production-ready features, you can always find the feature flag in the configuration file under the plugin that uses it. For beta features,
you may need to inquire what the feature flag is.
We recommend you make a back-up of your original config file before using this command.
  
  To update, use the config command:
```sh
tanzu config set features.<plugin-name>.<feature-name> <true|false>
```
An example might be: 
```sh
tanzu config set features.management-cluster.dual-stack true
```
For global options, an example might be: 
```sh
tanzu config set features.global.debug true
```

The options for the command are:

* `plugin-name`: a valid plugin that uses the feature name you are setting; there is no validation. Use `global` for CLI-wide features.
* `feature-name`: the feature name you are setting; there is no validation.
* `true|false`: consumers of the feature config value expect a boolean for features; there is no validation. 

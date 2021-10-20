# Configuration of Features

The Tanzu CLI offers the ability to configure CLI plugins based on feature flags. Some feature flags control behavior options that are part of a production-ready release,
some flags control features that are still in development. By default, the CLI will install with only shippable features. However, some beta features may be available by 
updating config options. To use the feature configuration, you need to know the name of the plugin and the feature flag for the functionality you want to activate 
(or deactivate). For production-ready features, you can always find the feature flag in the configuration file under the plugin that uses it. For beta features, you may need to
 inquire what the feature flag is (or you may find it in the configuration file set to `false`). 
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
* `true|false`: the code that consumes the feature config value expects a boolean for features and may throw an error if it encounters non-boolean text; there is no validation.

For developers making use of this feature:

* To set a default value in the tanzu config file, add an entry to `DefaultCliFeatureFlags` (in pkg/v1/config/clientconfig.go)
* To let users know the feature is available but still under development, use a `false` value; when ready for production, change to `true`
* Users can change the value by using the command above, or by manually editing their tanzu config file
* Throughout your code, you may use `cfg.IsConfigFeatureActivated()` to check the flag value (in apis/config/v1alpha1/clientconfig.go)

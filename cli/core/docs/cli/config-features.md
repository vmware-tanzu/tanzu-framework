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

* To set a default value in the tanzu config file, pass `DefaultFeatureFlags` as part of PluginDescriptor when defining the plugin
* Users can change the value by using the command above, or by manually editing their tanzu config file
* Throughout your code, you may use `cfg.IsConfigFeatureActivated()` to check the flag value (in apis/config/v1alpha1/clientconfig.go)

If you want to make this feature available for a beta period:

* To let users know the feature is available but still under development, use a `false` default value; when ready for production, change to `true`. This will create an entry in
 their config file so they can see the flag name.
* We recommend using two flags, one for the beta period and one for production. For the beta period, simply append `-beta` to the flag name that you expect to use in production.
For example, your production flag might be `features.global.foobar` and for the beta you could use `features.global.foobar-beta`. There are two advantages to this approach:
(1) The user is clear when they are using a beta flag and when they are using a production flag,
(2) There are no transition issues between beta flag use and production. (If you use the same flag name for beta and for production, then when the production code runs the
 previous "beta" setting will be taken as the production setting. This would force either the user or an installation script to activate the flag from `false` to `true` using
  the `tanzu set config` command above. Using two different flag names, there is no such issue.)
NOTE: there is no code that detects `-beta`; it is simply a recommended naming convention.

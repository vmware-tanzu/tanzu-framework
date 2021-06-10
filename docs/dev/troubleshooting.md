# Troubleshooting

When running commands or developing locally, sometimes things may go awry and you may need
more information on the status of the cluster.

The following tips may be useful when troubleshooting aspects of Core.

### Verbose Output

The Tanzu CLI provides verbose output options, and setting this to its highest value (`-v9`)
can be useful when unexpected things happen.

`-v, --verbose int32     Number for the log level verbosity(0-9)`

### Clean and Rebuild CLI Plugins

Sometimes it is necessary to sanitize your plugin environment by cleaning and rebuild your artifacts.
Over time, one may run into issues due to the plugin artifacts being in an inconsistent state due to prior
development builds. Cleaning the plugin versions and starting with fresh plugin builds can help with unexpected
or difficult CLI issues.

The [Build doc](build.md) describes how to accomplish this.

### Open an Issue

If you stumble into a situation where a clean build of the main branch will not build cleanly, please open an
issue for the Core team so we can help.

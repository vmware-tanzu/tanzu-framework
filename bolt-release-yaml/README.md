# BOLT Release Yamls

This project contains all the YAML configurations that are used by the BOLT (Build Opensource Like Thunder ⚡️) service.  
The YAML configurations include component definitions, release definitions and additional configurations used by BOLT.  

This is the repository for component configuration and release configuration.

## Component configuration

This repository serves as [component](component/README.md) database, holding all configurations of all components.  
refer to [component API](https://gitlab.eng.vmware.com/TKG/bolt/bolt-cli/-/blob/main/api/v1alpha2/component.go) when modifying component configure.

## Release configuration

[Release](release/README.md) configurations, refer to [release API](https://gitlab.eng.vmware.com/TKG/bolt/bolt-cli/-/blob/main/api/v1alpha2/release.go) when modifying release configure.

## Bot configurations

[Bot](bot/README.md) is a special type of configuration, Bot been used to
automate release workflow such as dailybuild, upstream to downstream release automation, CVE
handling etc.  
refer to [bot api](https://gitlab.eng.vmware.com/TKG/bolt/bolt-cli/-/blob/main/api/v1alpha2/bot.go) when writing bot configure.

## Project Organization

The project is organized into the following folders:

* `bot`  
  **WIP**. Holds the configurations for bot integrations for a component.
* `component`  
  Acts as a database for all the components that are handled by BOLT. Each component gets its own folder under `component`
* `filters`  
  Holds the definition of the filters used by aggregator type components
* `release`  
  Holds the release configurations
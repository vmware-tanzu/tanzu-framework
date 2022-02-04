# tanzu-management-cluster-ng-api

This project was generated with [Angular CLI](https://github.com/angular/angular-cli) version 12.2.13.

tanzu-management-cluster-ng-api includes all swagger generated models and Angular HTTP methods which are auto-generated from `tanzu-framework/pkg/v1/tkg/web/api/spec.yaml`.

This library serves as an interface for REST APIs which exist in the Tanzu CLI Management Cluster plugin. 

## Development Notes

Any additions to the spec.yaml swagger contract which need to be reflected in the UI require that tanzu-management-cluster-ng-api is regenerated with the updated spec.

### Building tanzu-management-cluster-ng-api

Two npm scripts found in `./package.json` are required for generating this library:

`build-library` - installs all project dependencies and then runs `generate-api-client` and `ng-build`.

`generate-api-client` - references `tanzu-framework/pkg/v1/tkg/web/api/spec.yaml` to generate all typescript models 
and api client interfaces in `./lib/swagger`.

## Build Scripts

The following NPM scripts in package.json are most relevant to a developer:

`build-tanzu-management-cluster-ng-api` Generates the swagger models and HTTP methods from `spec.yaml`; executed by `build:all`.

`link-tanzu-management-cluster-ng-api` Generates symlink within tanzu-management-cluster-ng-api; executed by `build:all`. `npm link` must still be run on
consuming application to complete symlink.

`build:all` Installs dependencies, builds the project and generates symlink. The build artifacts will be stored in the `dist/` directory.

### Versioning

Versioning of this library is set in `tanzu-framework/pkg/v1/tkg/web-libraries/tanzu-management-cluster-ng-libs/package.json`.
It is recommended that you bump the minor version of the library any time that the swagger contract `spec.yaml` is modified which
results in regenerating this library.

# tanzu-management-cluster-ng-api

This project was generated with [Angular CLI](https://github.com/angular/angular-cli) version 12.2.13.

tanzu-management-cluster-ng-api includes all swagger generated models and HTTP methods which are auto-generated from `tanzu-framework/pkg/v1/tkg/web/api/spec.yaml`

This library is not yet published to an NPM registry, so consumption of the library is done through symlinks. See documentation on building below.

## Development Notes

Any additions to the spec.yaml swagger contract which need to be reflected in the UI require that tanzu-management-cluster-ng-api is regenerated with these changes.

### Building tanzu-management-cluster-ng-api

There are several NPM scripts that are available for building this library and creating the symlink to consume the library. For a quickstart
experience, the suggested NPM script to run is `npm run build:all` in `tanzu-framework/pkg/v1/tkg/web-libraries/tanzu-management-cluster-ng-libs/package.json`.
This NPM script will build all necessary artifacts and complete one half of the symlink process by linking the `/dist` folder of `tanzu-management-cluster-ng-api`.

## Build Scripts

The following NPM scripts in package.json are most relevant to a developer:

`build-tanzu-management-cluster-ng-api` Generates the swagger models and HTTP methods from `spec.yaml`; executed by `build:all`.

`link-tanzu-management-cluster-ng-api` Generates symlink within tanzu-management-cluster-ng-api; executed by `build:all`. `npm link` must still be run on
consuming application to complete symlink.

`build:all` Installs dependencies, builds the project and generates symlink. The build artifacts will be stored in the `dist/` directory.

### Consumption

Consumption of this library in the Tanzu Kickstart UI requires completion of the symlink process, which is done for you in the
`tanzu-framework/pkg/v1/tkg/web/package.json` build and run scripts (see `npm link tanzu-management-cluster-ng-api`).

### Versioning

Versioning of this library is set in `tanzu-framework/pkg/v1/tkg/web-libraries/tanzu-management-cluster-ng-libs/package.json`.
It is recommended that you bump the minor version of the library any time that the swagger contract `spec.yaml` is modified which
results in regenerating this library.

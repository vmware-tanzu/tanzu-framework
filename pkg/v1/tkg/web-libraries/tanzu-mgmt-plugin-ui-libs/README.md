# tanzu-mgmt-plugin-api-lib

This project was generated with [Angular CLI](https://github.com/angular/angular-cli) version 12.2.13.

TanzuUiApiLib includes all swagger generated models and HTTP methods which are auto-generated from `tanzu-framework/pkg/v1/tkg/web/api/spec.yaml`

This library is not yet published to an NPM registry, so consumption of the library is done through symlinks. See documentation on building below.

## Development Notes

Any additions to the spec.yaml swagger contract which need to be reflected in the UI require that TanzuUiApiLib is regenerated with these changes.

### Building TanzuUiApiLib

There are several NPM scripts that are available for building this library and creating the symlink to consume the library. For a quickstart
experience, the suggested NPM script to run is `npm run build:all` in `tanzu-framework/pkg/v1/tkg/web-libraries/tanzu-mgmt-plugin-ui-libs/package.json`.
This NPM script will build all necessary artifacts and complete one half of the symlink process by linking the `/dist` folder of `tanzu-mgmt-plugin-api-lib`.

## Build Scripts

The following NPM scripts in package.json are most relevant to a developer:

`build-tanzu-mgmt-plugin-api-lib` Generates the swagger models and HTTP methods from `spec.yaml`; executed by `build:all`.

`link-tanzu-mgmt-plugin-api-lib` Generates symlink within TanzuUiApiLib; executed by `build:all`. `npm link` must still be run on
consuming application to complete symlink.

`build:all` Installs dependencies, builds the project and generates symlink. The build artifacts will be stored in the `dist/` directory.

### Consumption

Consumption of this library in the Tanzu Kickstart UI requires completion of the symlink process, which is done for you in the
`tanzu-framework/pkg/v1/tkg/web/package.json` build and run scripts (see `npm link tanzu-mgmt-plugin-api-lib`).

### Versioning

Versioning of this library is set in `tanzu-framework/pkg/v1/tkg/web-libraries/tanzu-mgmt-plugin-ui-libs/package.json`.
It is recommended that you bump the minor version of the library any time that the swagger contract `spec.yaml` is modified which
results in regenerating this library.

# TkgKickstartUI

This project was generated with [Angular CLI](https://github.com/angular/angular-cli) version 8.3.20.

## Prerequisites for Building and Running UI on Local Machine

Node version 10.x.x
 - `node --version` to check which version you have 

Node version can be set and managed by using NVM (Node Version Manager):
 - `brew install nvm`
 - `nvm install 10` (or `nvm use 10` to temporarily set node version)

To build the UI locally via Make, run `make ui-build` which will install node modules and compile UI assets into 'dist' folder.

If an alternate NPM registry is required to obtain the node dependencies, it should be configured either
 - prior to running the make target, with 'npm config set registry <register-url>', or
 - providing the URL in the CUSTOM_NPM_REGISTRY environment variable.

If running UI locally without executing Makefile script, execute `npm install` from tkg-cli/web folder prior to starting or compiling UI.

## UI served on local Angular CLI server

Prerequisite - node modules have been install via `make ui-build` or `npm install` in '/tkg-cli/web' directory

Run `npm run start` from /tkg-cli/web folder. Navigate to `http://localhost:4200/` in a browser. The app will automatically reload if you change any of the source files.

 - See `Running UI Mock API server` to make mock API endpoints available when developing on local machine.

## Running UI Mock API server

Run `npm run start:insecure` from /tkg-cli/web/node-server folder. Node.js will serve mock API endpoints on `http://localhost:8008`.

## Running CLI Locally and Launching UI

See Prerequisites for Building and Running UI on Local Machine

To serve the tkg ui, under tkg-cli repo, run: make tkg
Then run tkg init --infrastructure=<aws/vsphere> --ui, the command just starts the UI server and it will not trigger any tkg init steps.

The ui will be served at http://127.0.0.1:8080

## Build

Prerequisite - node modules have been install via `make ui-build` or `npm install` in '/tkg-cli/web' directory

Run `npm run build:prod` to build the project. The build artifacts will be stored in the '/dist' directory.

## Running unit tests

Run `npm run test` to execute the unit tests via [Karma](https://karma-runner.github.io).

## Pre-commit testing

Run `make pull-ci` prior to creating a pull request to run all github CI tests (UI and Golang tests).

## Running end-to-end tests

Run `ng e2e` to execute the end-to-end tests via [Protractor](http://www.protractortest.org/).

## Further help

To get more help on the Angular CLI use `ng help` or go check out the [Angular CLI README](https://github.com/angular/angular-cli/blob/master/README.md).

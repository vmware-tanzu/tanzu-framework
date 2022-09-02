# Tanzu CLI Integration Library

This library provides functionality useful in the development of Tanzu CLI plugins. It includes:

1. Component library
1. Config library
1. Plugin helpers
1. Command helpers
1. Test helpers

## Component Library

This package implements reusable CLI user interface components, including:

- output writers (table, listtable, json, yaml, spinner)
- prompt
- selector
- question

## Config Library

This package implements helper functions to read, write and update the tanzu configuration file (`~/.config/tanzu/config.yaml`).

## Plugin Helpers

This package implements helper functions for new plugin creation. This is one of the main packages that each and every plugin will need to import to integrate with the Tanzu CLI.

## Command Helpers

This package implements command specific helper functions like command deprecation, etc.

## Test Helpers

This package implements helper functions to develop test plugins.

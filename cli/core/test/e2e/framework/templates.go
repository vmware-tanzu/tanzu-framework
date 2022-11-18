// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package test defines the integration and end-to-end test case for cli core
package framework

const ScriptBasedPluginTemplate = `#!/bin/bash

# Minimally Viable Dummy Tanzu CLI 'Plugin'

info() {
   cat << EOF
{
  "name": "%s",
  "target":"%s",
  "description": "%s",
  "version": "%s",
  "buildSHA": "%s",
  "group": "%s",
  "hidden": %s,
  "aliases": %s
}
EOF
  exit 0
}
version() { 
	echo "%s"
}

case "$1" in
    info)  $1;;
    version)   $1;;
    *) cat << EOF
%s functionality

Usage:
  tanzu %s [command]

Available Commands:
  info     plugin details
  version   plugin version
EOF
       exit 0
       ;;
esac
`

const ImagesTemplate = `---
apiVersion: imgpkg.carvel.dev/v1alpha1
images:
kind: ImagesLock`

const GeneratedValuesTemplate = `#@data/values
#@overlay/match-child-defaults missing_ok=True

---`

#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" || exit; pwd)"

normalize() {
echo "---" > /tmp/norm.yaml
ytt -f "${SCRIPT_ROOT}"/normalize.yaml -f "$1" >> /tmp/norm.yaml

# ytt found to output documents in different order under some circumstances.
# order the yaml documents by their content before comparing/committing them so we can get
# deterministic outputs every time.
#
# starting with start record pattern ---, accumulate lines until next ---, replacing newlines with
# stand-in string _^_, sort results, output result after replacing stand-in back to newline
awk '/^---/{if(s){print s}s=$0} !/^---/{s=s"_^_"$0}END{print s}' /tmp/norm.yaml | sort | sed $'s/_\^_/\\\n/g' > "$2"
}

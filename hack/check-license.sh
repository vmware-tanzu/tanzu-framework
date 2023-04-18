#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

# TODO(navidshaikh): Add shellcheck and merge

verify_license() {
  printf "Checking License in shell scripts and Makefiles ...\n"
  local required_keywords=("VMware, Inc." "SPDX-License-Identifier: Apache-2.0")
  local file_patterns_to_check=("*.sh" "Makefile" "*.mk")

  local result
  result=$(mktemp /tmp/tf-licence-check.XXXXXX)
  for ext in "${file_patterns_to_check[@]}"; do
    find . -type d -o -name "$ext" -type f -print0 |
      while IFS= read -r -d '' path; do
        for rword in "${required_keywords[@]}"; do
          if ! grep -q "$rword" "$path"; then
            echo "   $path" >> "$result"
          fi
        done
      done
  done

  if [ -s "$result" ]; then
    echo "No required license header found in:"
    sort < "$result" | uniq
    echo "License check failed!"
    echo "Please add the license in each listed file and verify using './hack/check-license.sh'"
    rm "$result"
    return 1
  else
    echo "License check passed!"
    rm "$result"
    return 0
  fi
}

verify_license || exit 1; exit 0

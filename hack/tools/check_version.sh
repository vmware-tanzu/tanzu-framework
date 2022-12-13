#!/usr/bin/env bash

# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Takes path , program, version as input
# exit 0 if the program, when called from path, has the version in its output.
VERSION=`$1/$2 version` 
# remove the "v" from the version string...
VERSION_NUM=`echo $VERSION | sed 's/v//g'`
EXPECTED_NUM=`echo $3 | sed 's/v//g'`

# We dont check v bc, we dont care about "v0.40.0" vs "0.40.0" and some
# programs print version strings out differently.
if echo $VERSION_NUM | grep -q $EXPECTED_NUM ; then
    echo "correct version *** $VERSION ***"
    exit 0
else
    echo "incorrect version of $2: *** $VERSION *** "
    echo "missing *** $3 ***"
    echo "delete $1/$2 and retry build"
    exit 1
fi

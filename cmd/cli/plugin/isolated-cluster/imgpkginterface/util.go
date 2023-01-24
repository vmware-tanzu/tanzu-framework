// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package imgpkginterface

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func PrintErrorAndExit(err error) {
	fmt.Printf("failed with error %s\n", err.Error())
	os.Exit(1)
}

func IsTKGRTMVersion(tag string) bool {
	tag = strings.TrimPrefix(tag, "v")
	versions := strings.Split(tag, ".")
	if len(versions) != 3 {
		return false
	}
	for _, version := range versions {
		if _, err := strconv.Atoi(version); err != nil {
			return false
		}
	}
	return true
}

func UnderscoredPlus(s string) string {
	return strings.Replace(s, "+", "_", -1)
}

func ReplaceSlash(s string) string {
	return strings.Replace(s, "/", "-", -1)
}

// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"testing"
)

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		desc           string
		versionFlag    string
		subVersionFlag string
		pkg            *Package
		concatenator   string
		want           formattedVersion
	}{
		{
			desc:           "version begins with 'v', read subVersion from package",
			versionFlag:    "v1.0.0",
			subVersionFlag: "",
			pkg: &Package{
				PackageSubVersion: "v12.0",
			},
			concatenator: "_",
			want: formattedVersion{
				concatenator: "_",
				concat:       "v1.0.0_v12.0",
				concatNoV:    "1.0.0_v12.0",
				version:      "v1.0.0",
				noV:          "1.0.0",
				subVersion:   "v12.0",
			},
		},
		{
			desc:           "version begins with a number, subVersion flag takes precedence over package subVersion",
			versionFlag:    "1.0.0",
			subVersionFlag: "1309",
			pkg: &Package{
				PackageSubVersion: "12.0",
			},
			concatenator: "+",
			want: formattedVersion{
				concatenator: "+",
				concat:       "1.0.0+1309",
				concatNoV:    "1.0.0+1309",
				version:      "1.0.0",
				noV:          "1.0.0",
				subVersion:   "1309",
			},
		},
		{
			desc:           "package subVersion is empty, but subVersion flag is not",
			versionFlag:    "v1.0.0",
			subVersionFlag: "vmware.1",
			pkg: &Package{
				PackageSubVersion: "",
			},
			concatenator: "_",
			want: formattedVersion{
				concatenator: "_",
				concat:       "v1.0.0_vmware.1",
				concatNoV:    "1.0.0_vmware.1",
				version:      "v1.0.0",
				noV:          "1.0.0",
				subVersion:   "vmware.1",
			},
		},
		{
			desc:           "subVersion flag is empty, but package subVersion is not",
			versionFlag:    "v1.0.0",
			subVersionFlag: "",
			pkg: &Package{
				PackageSubVersion: "vmware.2",
			},
			concatenator: "+",
			want: formattedVersion{
				concatenator: "+",
				concat:       "v1.0.0+vmware.2",
				concatNoV:    "1.0.0+vmware.2",
				version:      "v1.0.0",
				noV:          "1.0.0",
				subVersion:   "vmware.2",
			},
		},
		{
			desc:           "no subVersion has been provided",
			versionFlag:    "v1.0.0",
			subVersionFlag: "",
			pkg: &Package{
				PackageSubVersion: "",
			},
			concatenator: "+",
			want: formattedVersion{
				concatenator: "+",
				concat:       "v1.0.0",
				concatNoV:    "1.0.0",
				version:      "v1.0.0",
				noV:          "1.0.0",
				subVersion:   "",
			},
		},
		{
			desc:           "package information nor subVersion flag not provided",
			versionFlag:    "v1.0.0",
			subVersionFlag: "",
			pkg:            nil,
			concatenator:   "+++",
			want: formattedVersion{
				concatenator: "+++",
				concat:       "v1.0.0",
				concatNoV:    "1.0.0",
				version:      "v1.0.0",
				noV:          "1.0.0",
				subVersion:   "",
			},
		},
		{
			desc:           "no package information, but subVersion flag provided",
			versionFlag:    "v1.0.0",
			subVersionFlag: "may.04",
			pkg:            nil,
			concatenator:   "",
			want: formattedVersion{
				concatenator: "",
				concat:       "v1.0.0may.04",
				concatNoV:    "1.0.0may.04",
				version:      "v1.0.0",
				noV:          "1.0.0",
				subVersion:   "may.04",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			// Set global variables.
			version = tc.versionFlag
			subVersion = tc.subVersionFlag

			if got := formatVersion(tc.pkg, tc.concatenator); got != tc.want {
				t.Errorf("want %#v, got %#v", tc.want, got)
			}
		})
	}
}

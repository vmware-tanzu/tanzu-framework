// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package version provides utility types and functions to work with SemVer versions.
package version

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	utilversion "k8s.io/apimachinery/pkg/util/version"
)

// Version is a structured representation of a SemVer version string (including build metadata).
type Version struct {
	version       *utilversion.Version
	buildMetadata BuildMetadata
}

func (v Version) String() string {
	return "v" + v.version.String()
}

// ParseSemantic constructs a structured representation of a semantic version string including build metadata,
// producing comparable structural representation.
func ParseSemantic(s string) (*Version, error) {
	version, err := utilversion.ParseSemantic(s)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing semantic version from string '%s'", s)
	}
	buildMetadata := ParseBuildMetadata(version.BuildMetadata())

	return &Version{version: version, buildMetadata: buildMetadata}, err
}

// LessThan compares Version to other Version. Returns true if this Version represents a less recent
// version based on comparisons of both Versions' SemVer versioning components and build metadata.
func (v *Version) LessThan(other *Version) bool {
	if other == nil {
		return false
	}
	if v == nil {
		return true
	}

	if other.version.LessThan(v.version) {
		return false
	}
	if v.version.LessThan(other.version) {
		return true
	}

	return v.buildMetadata.LessThan(other.buildMetadata)
}

// Major returns the major release number.
func (v *Version) Major() uint {
	return v.version.Major()
}

// Minor returns the minor release number.
func (v *Version) Minor() uint {
	return v.version.Minor()
}

// BuildMetadata is a structured representation of build metadata in a SemVer version string
// (the part after '+' in "+vmware.1-...").
type BuildMetadata []string

// ParseBuildMetadata creates BuildMetadata  from a SemVer version string representation of build metadata
// (the part after '+' in "+vmware.1-...").
func ParseBuildMetadata(s string) BuildMetadata {
	var result BuildMetadata
	for s != "" {
		sepIndex := strings.IndexFunc(s, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		if sepIndex == 0 {
			sepIndex = 1
		}
		if sepIndex == -1 {
			sepIndex = len(s)
		}
		result = append(result, s[:sepIndex])
		s = s[sepIndex:]
	}

	return result
}

// LessThan compares BuildMetadata to other BuildMetadata. Returns true if this BuildMetadata represents a less recent
// build based on numeric (if possible) and alphanumeric comparison of build metadata components.
func (bm BuildMetadata) LessThan(other BuildMetadata) bool {
	if len(other) == 0 {
		return false
	}
	if len(bm) == 0 {
		return true
	}

	if bm[0] == other[0] {
		return bm[1:].LessThan(other[1:])
	}

	if bm0num, err := strconv.ParseInt(bm[0], 16, 0); err == nil {
		if other0num, err := strconv.ParseInt(other[0], 16, 0); err == nil {
			return bm0num < other0num
		}
	}

	return strings.Compare(bm[0], other[0]) < 0
}

// Prefixes returns the set of all possible version prefixes for a version string, including itself.
// Both label and SemVer formats are supported.
// E.g. for "v1.17.9---vmware.2" the result is:
// {"v1.17.9---vmware.2": "", "v1.17.9---vmware": "", "v1.17.9": "", "v1.17": "", "v1": ""}
func Prefixes(v string) labels.Set {
	result := labels.Set{}
	for ; v != ""; v = strings.TrimRightFunc(v, vSuffix()) {
		result[v] = ""
	}
	return result
}

func vSuffix() func(rune) bool {
	inDelimiter := false
	return func(r rune) bool {
		if !inDelimiter {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return true
			}
			inDelimiter = true
			return true
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
		return true
	}
}

var plusReplacer = strings.NewReplacer("+", "---")

// Label converts version string in SemVer format to label format.
func Label(v string) string {
	return "v" + plusReplacer.Replace(strings.TrimPrefix(v, "v"))
}

// WithV makes sure 'v' is prepended to the version string.
func WithV(s string) string {
	if strings.HasPrefix(s, "v") {
		return s
	}
	return "v" + s
}

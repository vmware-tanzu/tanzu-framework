// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSelectVersionStable(t *testing.T) {
	for _, test := range []struct {
		name     string
		versions []string
		max      string
	}{
		{
			name:     "basic patch",
			versions: []string{"v0.0.1", "v0.0.2"},
			max:      "v0.0.2",
		},
		{
			name:     "release candidates",
			versions: []string{"v0.0.1", "v1.3.0-rc.1", "v1.3.0-pre-alpha-1"},
			max:      "v0.0.1",
		},
		{
			name:     "release candidates same",
			versions: []string{"v0.0.1", "v1.3.0", "v1.3.0-rc.1", "v1.3.0-pre-alpha-1"},
			max:      "v1.3.0",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := Plugin{
				Name:     "foo",
				Versions: test.versions,
			}
			v := p.FindVersion(SelectVersionStable)
			require.Equal(t, test.max, v)
		})
	}
}

func TestSelectVersionAny(t *testing.T) {
	for _, test := range []struct {
		name     string
		versions []string
		max      string
	}{
		{
			name:     "basic patch",
			versions: []string{"v0.0.1", "v0.0.2"},
			max:      "v0.0.2",
		},
		{
			name:     "release candidates",
			versions: []string{"v0.0.1", "v1.3.0-rc.1", "v1.3.0-pre-alpha-1"},
			max:      "v1.3.0-rc.1",
		},
		{
			name:     "release candidates same",
			versions: []string{"v0.0.1", "v1.3.0", "v1.3.0-rc.1", "v1.3.0-pre-alpha-1"},
			max:      "v1.3.0",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := Plugin{
				Name:     "foo",
				Versions: test.versions,
			}
			v := p.FindVersion(SelectVersionAny)
			require.Equal(t, test.max, v)
		})
	}
}

func TestSelectVersionAlpha(t *testing.T) {
	for _, test := range []struct {
		name     string
		versions []string
		max      string
	}{
		{
			name:     "basic patch",
			versions: []string{"v0.0.1", "v0.0.2"},
			max:      "v0.0.2",
		},
		{
			name:     "alpha prerelease mixed with release candidate",
			versions: []string{"v0.0.1", "v1.3.0-rc.1", "v1.4.0-alpha", "v1.5.0-alpha+meta1234"},
			max:      "v1.4.0-alpha",
		},
		{
			name:     "release candidates same",
			versions: []string{"v0.0.1", "v1.3.0", "v1.3.0-rc.1", "v1.3.0-alpha"},
			max:      "v1.3.0",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := Plugin{
				Name:     "foo",
				Versions: test.versions,
			}
			v := p.FindVersion(SelectVersionAlpha)
			require.Equal(t, test.max, v)
		})
	}
}

func TestSelectVersionExperimental(t *testing.T) {
	for _, test := range []struct {
		name     string
		versions []string
		max      string
	}{
		{
			name:     "basic patch",
			versions: []string{"v0.0.1", "v0.0.2"},
			max:      "v0.0.2",
		},
		{
			name:     "experimental builds and prereleases",
			versions: []string{"v0.0.1", "v1.3.0-rc.1", "v1.4.0-random-nonalpha", "v1.4.1-rc.1+build546"},
			max:      "v1.4.0-random-nonalpha",
		},
		{
			name:     "stable release candidates",
			versions: []string{"v0.0.1", "v1.3.0", "v1.3.0-rc.1", "v1.3.0-pre-alpha-1", "v1.3.0-rc.1+build546"},
			max:      "v1.3.0",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := Plugin{
				Name:     "foo",
				Versions: test.versions,
			}
			v := p.FindVersion(SelectVersionExperimental)
			require.Equal(t, test.max, v)
		})
	}
}

func TestFilterVersions(t *testing.T) {
	for _, test := range []struct {
		name            string
		includeUnstable bool
		versions        []string
		results         []string
	}{
		{
			name:     "ignore bad versions",
			versions: []string{"v0.0.1", "dev", "bad.version.not.semver", "v0.0.2"},
			results:  []string{"v0.0.1", "v0.0.2"},
		},
		{
			name:     "ignore bad versions including unstable",
			versions: []string{"v0.0.1", "v1.3.0-rc.1", "v1.3.0-pre-alpha-1", "carrot"},
			results:  []string{"v0.0.1", "v1.3.0-rc.1", "v1.3.0-pre-alpha-1"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			v := FilterVersions(test.versions)
			require.Equal(t, test.results, v)
		})
	}
}

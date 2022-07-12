// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package propagation

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/pointer"
)

func TestStringValueReplacer(t *testing.T) {
	replacer := stringValueReplacer{
		old: "tkg-system",
		new: "default",
	}

	m0 := map[string]interface{}{
		"ns1": "tkg-system",
		"ns2": pointer.StringPtr("tkg-system"),
		"ns3": map[string]string{
			"ns": "tkg-system",
		},
		"ns4": map[string]*string{
			"ns": pointer.StringPtr("tkg-system"),
		},
	}
	u := &unstructured.Unstructured{Object: m0}

	replacer.Replace(u)

	require.Equal(t, replacer.new, u.Object["ns1"])
	require.Equal(t, replacer.new, *u.Object["ns2"].(*string))
	require.Equal(t, replacer.new, u.Object["ns3"].(map[string]string)["ns"])
	require.Equal(t, replacer.new, *u.Object["ns4"].(map[string]*string)["ns"])
}

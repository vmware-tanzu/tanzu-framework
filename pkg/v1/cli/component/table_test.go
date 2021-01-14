// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTableWriter(t *testing.T) {
	tab := NewTableWriter("a", "b", "c")
	require.NotNil(t, tab)
}

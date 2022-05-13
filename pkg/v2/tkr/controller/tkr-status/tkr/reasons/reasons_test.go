// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package reasons

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReason(t *testing.T) {
	var err HasReason
	err = TKRVersionMismatch("")
	require.Equal(t, "TKRVersionMismatch", err.Reason())
	err = OSImageVersionMismatch("")
	require.Equal(t, "OSImageVersionMismatch", err.Reason())
	err = MissingOSImage("")
	require.Equal(t, "MissingOSImage", err.Reason())
	err = MissingBootstrapPackage("")
	require.Equal(t, "MissingBootstrapPackage", err.Reason())
	err = MissingClusterBootstrapTemplate("")
	require.Equal(t, "MissingClusterBootstrapTemplate", err.Reason())
}

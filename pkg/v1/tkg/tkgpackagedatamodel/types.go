// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// PackagePluginNonCriticalError is used for non critical package plugin errors which should be treated more like warnings
type PackagePluginNonCriticalError struct {
	Reason string
}

func (e *PackagePluginNonCriticalError) Error() string { return e.Reason }

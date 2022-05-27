// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"context"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/sets"
)

type Compatibility interface {
	CompatibleVersions(ctx context.Context) (sets.StringSet, error)
}

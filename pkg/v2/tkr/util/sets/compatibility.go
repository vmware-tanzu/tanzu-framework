// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package sets

import (
	"context"
)

type Compatibility interface {
	CompatibleVersions(ctx context.Context) (StringSet, error)
}

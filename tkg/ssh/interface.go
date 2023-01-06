// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (

)

// Client defines methods to access Oracle Cloud inventory
type Client interface {
	KeysAsString() (string, error)
}

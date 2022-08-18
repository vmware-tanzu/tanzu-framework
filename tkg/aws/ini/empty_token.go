// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ini

// emptyToken is used to satisfy the Token interface
var emptyToken = newToken(TokenNone, []rune{}, NoneType)

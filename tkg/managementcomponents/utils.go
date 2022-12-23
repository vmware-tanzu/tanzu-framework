// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents

import (
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	ErrReconciliationFailed  = "resource reconciliation failed"
	ErrReconciliationTimeout = "resource reconciliation timeout"
)

var DefaultRetry = wait.Backoff{
	Steps:    10,
	Duration: 10 * time.Second,
	Factor:   1.0,
	Jitter:   0,
}

func IsReconciliationError(err error) bool {
	if strings.Contains(err.Error(), ErrReconciliationFailed) ||
		strings.Contains(err.Error(), ErrReconciliationTimeout) {
		return true
	}
	return false
}

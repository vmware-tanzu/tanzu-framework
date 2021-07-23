// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

// PackageProgress includes channels for sending messages
type PackageProgress struct {
	// Use buffered chan so that sending goroutine doesn't block
	ProgressMsg chan string
	// Err chan for reporting errors
	Err chan error
	// Empty struct for signaling that goroutine is finished
	Done chan struct{}
}

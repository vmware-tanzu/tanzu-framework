// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package upgrade contains upgrade related api structs
package upgrade

// Phase defines upgrade phase of a cluster
type Phase string

// Upgrade phase constants
const (
	Init         Phase = ""
	ControlPlane Phase = "ControlPlane"
	WorkerPlane  Phase = "WorkerPlane"
	Success      Phase = "Success"
)

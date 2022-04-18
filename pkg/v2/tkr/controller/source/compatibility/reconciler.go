// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package compatibility blah blah.
// TODO write doc
package compatibility

import (
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	Log    logr.Logger
	Client client.Client
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return nil
}

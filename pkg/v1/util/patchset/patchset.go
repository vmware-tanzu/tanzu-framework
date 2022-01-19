// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package patchset provides the patchSet utility type.
package patchset

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PatchSet is used to snapshot multiple objects and then patch them all at once.
type PatchSet interface {
	// Add an object (which is a pointer to a struct) to the PatchSet. The object's snapshot is taken.
	// Now, any changes made to the object will result in a patch being produced and applied at during Apply call.
	Add(client.Object)

	// Objects returns all objects that have been added to the PatchSet.
	Objects() map[types.UID]client.Object

	// Apply calculates patches to all added objects and applies them all at once.
	Apply(context.Context) error
}

func New(c client.Client) PatchSet {
	return &patchSet{
		client:   c,
		patchers: map[types.UID]*patcher{},
	}
}

type patchSet struct {
	client   client.Client
	patchers map[types.UID]*patcher
}

type patcher struct {
	obj    client.Object
	helper *patch.Helper
}

func (ps *patchSet) Add(obj client.Object) {
	uid := obj.GetUID()
	if _, exists := ps.patchers[uid]; exists {
		return
	}

	helper, _ := patch.NewHelper(obj, ps.client)
	ps.patchers[uid] = &patcher{
		obj:    obj,
		helper: helper,
	}
}

func (ps patchSet) Objects() map[types.UID]client.Object {
	result := make(map[types.UID]client.Object, len(ps.patchers))
	for k, v := range ps.patchers {
		result[k] = v.obj
	}
	return result
}

func (ps patchSet) Apply(ctx context.Context) error {
	errs := make([]error, 0, len(ps.patchers))
	for _, patcher := range ps.patchers {
		if err := patcher.helper.Patch(ctx, patcher.obj); err != nil {
			if !patcher.obj.GetDeletionTimestamp().IsZero() && isNotFound(err) {
				continue
			}
			errs = append(errs, err)
		}
	}
	return kerrors.NewAggregate(errs)
}

func isNotFound(err error) bool {
	return kerrors.FilterOut(err, apierrors.IsNotFound) == nil
}

// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package patchset provides the patchSet utility type.
package patchset

import (
	"context"
	"reflect"
	"sync"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
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
	sync.RWMutex
	client   client.Client
	patchers map[types.UID]*patcher
}

type patcher struct {
	obj       client.Object
	beforeObj client.Object
}

func (h *patcher) patch(ctx context.Context, c client.Client, obj client.Object) error {
	for _, f := range []func() error{
		func() error {
			return c.Patch(ctx, obj, client.MergeFromWithOptions(h.beforeObj, client.MergeFromWithOptimisticLock{}))
		},
		func() error {
			err := c.Status().Patch(ctx, obj, client.MergeFromWithOptions(h.beforeObj, client.MergeFromWithOptimisticLock{}))
			return kerrors.FilterOut(err, apierrors.IsNotFound) // status resource may not exist
		},
	} {
		if err := f(); err != nil {
			return err
		}
		h.beforeObj.SetResourceVersion(obj.GetResourceVersion()) // get the new resourceVersion after a successful patch
	}
	return nil
}

func (ps *patchSet) Add(obj client.Object) {
	ps.Lock()
	defer ps.Unlock()

	uid := obj.GetUID()
	if _, exists := ps.patchers[uid]; exists {
		return
	}

	ps.patchers[uid] = &patcher{
		obj:       obj,
		beforeObj: obj.DeepCopyObject().(client.Object),
	}
}

func (ps *patchSet) Objects() map[types.UID]client.Object {
	ps.RLock()
	defer ps.RUnlock()

	result := make(map[types.UID]client.Object, len(ps.patchers))
	for k, v := range ps.patchers {
		result[k] = v.obj
	}
	return result
}

func (ps *patchSet) Apply(ctx context.Context) error {
	ps.Lock()
	defer ps.Unlock()

	errs := make([]error, 0, len(ps.patchers))
	for _, patcher := range ps.patchers {
		if reflect.DeepEqual(patcher.obj, patcher.beforeObj) {
			continue
		}
		if err := patcher.patch(ctx, ps.client, patcher.obj); err != nil {
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

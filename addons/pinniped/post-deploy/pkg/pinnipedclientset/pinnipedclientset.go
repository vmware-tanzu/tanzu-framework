// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package pinnipedclientset provides implementations for clientsets for the Pinniped API.
//
// These hand-written Pinniped clientset implementations are required because the Pinniped instance
// running on TKG uses a non-default API group, and therefore we must use a dynamic (non-generated)
// client to talk to it.
package pinnipedclientset

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/constants"
)

// translateAPIGroup translates the provided pinniped baseAPIGroup to a new API group based on the
// provided newAPIGroupSuffix.
//
// We assume that all baseAPIGroup's will end in "pinniped.dev", and therefore we can safely replace
// the reference to "pinniped.dev" with the provided newAPIGroupSuffix. If the provided baseAPIGroup
// does not end in "pinniped.dev", then this function will return an empty string.
func translateAPIGroup(baseAPIGroup, newAPIGroupSuffix string) string {
	// Note - most of this logic is copied from:
	//   https://github.com/vmware-tanzu/pinniped/blob/v0.10.0/internal/groupsuffix/groupsuffix.go#L160-L165
	if !strings.HasSuffix(baseAPIGroup, "."+constants.PinnipedDefaultAPIGroupSuffix) {
		zap.S().Warn("cannot convert pinniped API group to TKG API group", "pinnipedAPIGroup", baseAPIGroup)
		return ""
	}
	return strings.TrimSuffix(baseAPIGroup, constants.PinnipedDefaultAPIGroupSuffix) + newAPIGroupSuffix
}

func create(ctx context.Context, client dynamic.ResourceInterface, obj metav1.Object, opts metav1.CreateOptions, newObj metav1.Object, objKind string) error {
	return createOrUpdate(ctx, client, obj, opts, metav1.UpdateOptions{}, newObj, objKind, true)
}

func update(ctx context.Context, client dynamic.ResourceInterface, obj metav1.Object, opts metav1.UpdateOptions, newObj metav1.Object, objKind string) error {
	return createOrUpdate(ctx, client, obj, metav1.CreateOptions{}, opts, newObj, objKind, false)
}

func createOrUpdate(ctx context.Context, client dynamic.ResourceInterface, obj metav1.Object, createOpts metav1.CreateOptions, updateOpts metav1.UpdateOptions, newObj metav1.Object, objKind string, create bool) error {
	unstructuredObjData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return fmt.Errorf("cannot convert %s %s/%s to unstructured: %w", objKind, obj.GetNamespace(), obj.GetName(), err)
	}

	var unstructuredObj *unstructured.Unstructured
	if create {
		unstructuredObj, err = client.Create(ctx, &unstructured.Unstructured{Object: unstructuredObjData}, createOpts)
	} else {
		unstructuredObj, err = client.Update(ctx, &unstructured.Unstructured{Object: unstructuredObjData}, updateOpts)
	}
	if err != nil {
		return err
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, newObj); err != nil {
		return fmt.Errorf("cannot convert unstructured object to %s: %w", objKind, err)
	}

	return nil
}

func get(ctx context.Context, client dynamic.ResourceInterface, name string, opts metav1.GetOptions, newObj metav1.Object, objKind string) error {
	unstructuredObj, err := client.Get(ctx, name, opts)
	if err != nil {
		return err
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, newObj); err != nil {
		return fmt.Errorf("cannot convert unstructured object to %s: %w", objKind, err)
	}

	return nil
}

// nolint:gocritic // DeleteOptions is usually passed by value, so keep the same convention here.
func deleete(ctx context.Context, client dynamic.ResourceInterface, name string, opts metav1.DeleteOptions, objKind string) error {
	err := client.Delete(ctx, name, opts)
	if err != nil {
		return fmt.Errorf("cannot delete %s ?/%s: %w", objKind, name, err)
	}

	return nil
}

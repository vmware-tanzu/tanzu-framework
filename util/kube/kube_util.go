// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kube has utility functions for interacting with kubernetes cluster using client-go library
package kube

import (
	"bytes"
	"context"
	"io"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func CreateResourcesFromManifest(manifestBytes []byte, cfg *rest.Config, dynamicClient dynamic.Interface) error {
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(manifestBytes), 100)
	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return err
	}
	for {
		resource, unstructuredObj, err := getResource(decoder, mapper, dynamicClient)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		_, err = resource.Create(context.Background(), unstructuredObj, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func getResource(decoder *yamlutil.YAMLOrJSONDecoder, mapper meta.RESTMapper, dynamicClient dynamic.Interface) (
	dynamic.ResourceInterface, *unstructured.Unstructured, error) { // nolint:whitespace
	var rawObj runtime.RawExtension
	if err := decoder.Decode(&rawObj); err != nil {
		return nil, nil, err
	}

	obj, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, nil, err
	}

	unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, nil, err
	}

	var resource dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if unstructuredObj.GetNamespace() == "" {
			unstructuredObj.SetNamespace("default")
		}
		resource = dynamicClient.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
	} else {
		resource = dynamicClient.Resource(mapping.Resource)
	}
	return resource, unstructuredObj, nil
}

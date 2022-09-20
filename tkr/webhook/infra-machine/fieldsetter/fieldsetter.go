// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package fieldsetter defines methods to set the fields of <Infra>Machine resource
package fieldsetter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	osImageRefAnnotationKey = "run.tanzu.vmware.com/os-image-ref"
)

// FieldSetter mutates <infra>Machine
type FieldSetter struct {
	decoder      *admission.Decoder
	Log          logr.Logger
	FieldPathMap map[string]string
}

// Handle will handle the infra admission request
func (fs *FieldSetter) Handle(ctx context.Context, req admission.Request) admission.Response { // nolint:gocritic // suppress linter error: hugeParam: req is heavy (400 bytes); consider passing by pointer (gocritic)
	fs.Log.Info("Received the infra machine admission request")
	infraMachine := &unstructured.Unstructured{}
	err := fs.decoder.Decode(req, infraMachine)
	if err != nil {
		fs.Log.Error(err, "Failed to decode infra machine admission request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	annotationValues, err := getOSImageRefAnnotation(infraMachine)
	if err != nil {
		fs.Log.Error(err, "Failed to get 'run.tanzu.vmware.com/os-image-ref' annotation")
		return admission.Errored(http.StatusBadRequest, err)
	}
	// if 'run.tanzu.vmware.com/os-image-ref' annotation is not set, nothing to update
	if annotationValues == nil {
		return admission.ValidationResponse(true, "")
	}

	err = fs.setFields(infraMachine, annotationValues)
	if err != nil {
		fs.Log.Error(err, "Failed to set the fields using the 'run.tanzu.vmware.com/os-image-ref' annotation")
		return admission.Errored(http.StatusBadRequest, err)
	}

	marshalledInfraMachine, err := json.Marshal(infraMachine)
	if err != nil {
		fs.Log.Error(err, "Failed to marshal infraMachine object")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalledInfraMachine)
}

func getOSImageRefAnnotation(infraMachine *unstructured.Unstructured) (map[string]interface{}, error) {
	annotations := infraMachine.GetAnnotations()
	if annotations == nil {
		return nil, nil
	}

	osImageRefAnnotationValue, exists := annotations[osImageRefAnnotationKey]
	if !exists {
		return nil, nil
	}
	annotationValues := make(map[string]interface{}, 1)
	err := yaml.Unmarshal([]byte(osImageRefAnnotationValue), &annotationValues)
	if err != nil {
		return nil, errors.New("failed to unmarshal 'run.tanzu.vmware.com/os-image-ref' annotation")
	}
	return annotationValues, nil
}

// SetFields sets the <Infra>Machine fields using the annotation values
func (fs *FieldSetter) setFields(o *unstructured.Unstructured, annotationValues map[string]interface{}) error {
	for field, value := range annotationValues {
		path, exists := fs.FieldPathMap[field]
		if !exists {
			fs.Log.Info(fmt.Sprintf("os Image reference annotation value's field %q doesn't match with any entry in field path map configured", field))
			continue
		}
		fieldPath := strings.Split(path, ".")
		// check if the field path exists
		if _, exists, _ := unstructured.NestedFieldNoCopy(o.UnstructuredContent(), fieldPath...); !exists {
			fs.Log.Info(fmt.Sprintf("Field path %q doesn't exists in the request object", path))
			return fmt.Errorf("field path %q doesn't exists in the request object", path)
		}

		err := unstructured.SetNestedField(o.UnstructuredContent(), value, fieldPath...)
		if err != nil {
			return errors.Wrapf(err, "failed to set the %q value to %v", fieldPath, value)
		}
	}
	return nil
}

// InjectDecoder injects the decoder. A decoder will be automatically injected.
func (fs *FieldSetter) InjectDecoder(d *admission.Decoder) error {
	fs.decoder = d
	return nil
}

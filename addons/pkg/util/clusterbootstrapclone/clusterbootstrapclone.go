// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterbootstrapclone

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/pointer"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

type Helper struct {
	Ctx                         context.Context
	K8sClient                   client.Client
	AggregateAPIResourcesClient client.Client
	DynamicClient               dynamic.Interface
	GVRHelper                   util.GVRHelper
	Logger                      logr.Logger
}

// NewHelper instantiates a new helper instance
func NewHelper(ctx context.Context, k8sClient client.Client, aggregateAPIResourcesClient client.Client,
	dynamicClient dynamic.Interface,
	gvrHelper util.GVRHelper,
	logger logr.Logger) *Helper {

	return &Helper{
		Ctx:                         ctx,
		K8sClient:                   k8sClient,
		AggregateAPIResourcesClient: aggregateAPIResourcesClient,
		DynamicClient:               dynamicClient,
		GVRHelper:                   gvrHelper,
		Logger:                      logger,
	}
}

// CompleteCBPackageRefNamesFromTKR goes through the ClusterBootstrapPackages and makes the RefName complete.
// For example, when the clusterBootstrapPackage.RefName passed in is "calico*", this function updates it to be
// a fully qualified name(<packageShortName>.<domain>.<version>). The fully qualified name is fetched from TKR.
func (h *Helper) CompleteCBPackageRefNamesFromTKR(tkr *runtanzuv1alpha3.TanzuKubernetesRelease, clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap) error {
	var suffix = "*"
	clusterBootstrapPackages := []*runtanzuv1alpha3.ClusterBootstrapPackage{
		clusterBootstrap.Spec.CNI,
		clusterBootstrap.Spec.CPI,
		clusterBootstrap.Spec.CSI,
		clusterBootstrap.Spec.Kapp,
	}
	clusterBootstrapPackages = append(clusterBootstrapPackages, clusterBootstrap.Spec.AdditionalPackages...)

	for _, clusterBootstrapPackage := range clusterBootstrapPackages {
		if clusterBootstrapPackage != nil && strings.HasSuffix(clusterBootstrapPackage.RefName, suffix) {
			// The partial filled RefName will end with *. E.g., refName: calico*
			prefixToMatch := strings.TrimSuffix(clusterBootstrapPackage.RefName, suffix)
			fullyQualifiedCBPackageRefName, err := getFullyQualifiedCBPackageRefName(tkr, prefixToMatch)
			if err != nil {
				// Error means no match found. We do not expect this to happen.
				// For atomicity, we return error immediately if there is a no match.
				return err
			}
			clusterBootstrapPackage.RefName = fullyQualifiedCBPackageRefName
		}
	}
	return nil
}

func getFullyQualifiedCBPackageRefName(tkr *runtanzuv1alpha3.TanzuKubernetesRelease, prefix string) (string, error) {
	var fullyQualifiedCBPackageRefName string
	var found bool
	for _, tkrBootstrapPackage := range tkr.Spec.BootstrapPackages {
		if strings.HasPrefix(tkrBootstrapPackage.Name, prefix) {
			if found {
				// It is possible that current prefix matches multiple bootstrapPackages. For example, c* might match
				// capabilities bootstrapPackage and calico bootstrapPackage. We would like to return error in this case
				// to let caller revise the prefix in order to narrow down to exactly one match.
				return fullyQualifiedCBPackageRefName, fmt.Errorf("multiple bootstrapPackage names matche the prefix %s within the TanzuKubernetesRelease %s", prefix, tkr.Name)
			}
			fullyQualifiedCBPackageRefName = tkrBootstrapPackage.Name
			found = true
		}
	}
	if !found {
		return fullyQualifiedCBPackageRefName, fmt.Errorf("no bootstrapPackage name matches the prefix %s within the BootstrapPackages [%v] of TanzuKubernetesRelease %s", prefix, tkr.Spec.BootstrapPackages, tkr.Name)
	}
	return fullyQualifiedCBPackageRefName, nil
}

// AddMissingSpecFieldsFromTemplate scans clusterBootstrap's fields. For fields which are not specified, it adds defaults from
// clusterBootstrapTemplate
func (h *Helper) AddMissingSpecFieldsFromTemplate(clusterBootstrapTemplate *runtanzuv1alpha3.ClusterBootstrapTemplate,
	clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, keysToSkip map[string]interface{}) error {

	h.Logger.Info(fmt.Sprintf("adding the missing spec fields of ClusterBootstrap %s/%s from ClusterBootstrapTemplate %s/%s",
		clusterBootstrap.Namespace, clusterBootstrap.Name, clusterBootstrapTemplate.Namespace, clusterBootstrapTemplate.Name))

	converter := runtime.DefaultUnstructuredConverter
	var copyFrom map[string]interface{}
	var target map[string]interface{}
	var err error
	// DeepCopy() here is to make sure to handle the pointer fields properly. We do not want any changes in clusterBootstrapTemplate
	// have side effects on the ClusterBootstrap object
	if copyFrom, err = converter.ToUnstructured(clusterBootstrapTemplate.Spec.DeepCopy()); err != nil {
		return err
	}
	if target, err = converter.ToUnstructured(clusterBootstrap.Spec.DeepCopy()); err != nil {
		return err
	}
	if err := addMissingFields(copyFrom, target, keysToSkip); err != nil {
		return err
	}
	updatedTemplateSpec := &runtanzuv1alpha3.ClusterBootstrapTemplateSpec{}
	if err := converter.FromUnstructured(target, updatedTemplateSpec); err != nil {
		return err
	}
	clusterBootstrap.Spec = updatedTemplateSpec
	h.Logger.Info(fmt.Sprintf("added the missing spec fields of ClusterBootstrap %s/%s from ClusterBootstrapTemplate %s/%s",
		clusterBootstrap.Namespace, clusterBootstrap.Name, clusterBootstrapTemplate.Namespace, clusterBootstrapTemplate.Name))
	return nil
}

func addMissingFields(copyFrom, destination, keysToSkip map[string]interface{}) error {
	for keyInFrom, valueInFrom := range copyFrom {
		if keysToSkip != nil {
			// There might be some keys we do not want to add to destination.
			if _, exist := keysToSkip[keyInFrom]; exist {
				continue
			}
		}
		valueInTarget, exist := destination[keyInFrom]
		if !exist || valueInTarget == nil {
			// If keyInFrom does not exist in destination or valueInTarget is nil, we need to copy valueInFrom and add to
			// destination.
			if valueInFrom != nil {
				valueInFromType := reflect.TypeOf(valueInFrom)
				if valueInFromType.Kind() == reflect.Map {
					// If the valueInFrom is a map, we copy it directly.
					copiedVal, _, copyErr := unstructured.NestedFieldCopy(valueInFrom.(map[string]interface{}))
					if copyErr != nil {
						return copyErr
					}
					// unstructured.NestedFieldCopy() copies all fields, but we need to delete some fields which are meant
					// to be skipped.
					if removeErr := removeFields(copiedVal.(map[string]interface{}), keysToSkip); removeErr != nil {
						return removeErr
					}
					destination[keyInFrom] = copiedVal
				} else if valueInFromType.Kind() == reflect.Slice {
					// If the valueInFrom is a slice, e.g., additionalPackages from a ClusterBootstrapTemplate, we iterate
					// on each item of the slice and make a copy of it.
					var copiedSlice []interface{}
					for _, sliceItemInFrom := range valueInFrom.([]interface{}) {
						if reflect.TypeOf(sliceItemInFrom).Kind() == reflect.Map {
							// Instantiate an empty map and invoke addMissingFields() to add missing fields into it. Logically,
							// this is doing a copy of the slice item and remove the keys which are meant to be skipped.
							copiedMap := make(map[string]interface{})
							if copyErr := addMissingFields(sliceItemInFrom.(map[string]interface{}), copiedMap, keysToSkip); copyErr != nil {
								return copyErr
							}
							copiedSlice = append(copiedSlice, copiedMap)
						} else {
							// TODO: The following line is doing a shallow copy, we might want to revisit in the future.
							// https://github.com/vmware-tanzu/tanzu-framework/issues/2785
							copiedSlice = append(copiedSlice, sliceItemInFrom)
						}
					}
					destination[keyInFrom] = copiedSlice
				} else {
					// TODO: Handle the primitive pointer, e.g., *int. Ideally we need to make a copy of the pointer
					// instead of reassigning directly. It is a shallow copy with current approach, copyFrom and destination
					// share the same underlying data. We are good at the moment because this function is internally
					// used for handling ClusterBootstrapTemplateSpec which does not have any primitive pointers in its
					// API definition.
					// https://github.com/vmware-tanzu/tanzu-framework/issues/2785
					destination[keyInFrom] = valueInFrom
				}
			}
		} else {
			// If keyInFrom exists in destination, recursively look inside the nested fields.
			if valueInFrom != nil && reflect.TypeOf(valueInFrom).Kind() == reflect.Map &&
				valueInTarget != nil && reflect.TypeOf(valueInTarget).Kind() == reflect.Map &&
				// ClusterBootstrapPackage.ValuesFrom.Inline is a map[string]interface{} contains the data values of a
				// specific ClusterBootstrapPackage. When the key "inline" exist in the destination we should not attempt
				// to invoke the addMissingFields(), otherwise the processed ClusterBootstrapPackage.ValuesFrom.Inline will
				// contain a combined data value from both copyFrom and destination.
				// E.g., ClusterBootstrap.Spec.CSI.ValuesFrom.Inline has {"a": "b"},
				// and ClusterBootstrapTemplate.Spec.CSI.ValuesFrom.Inline has {"foo":"bar"}.
				// If we do not have the following check the final processed ClusterBootstrap will have {"a": "b", "foo": "bar"}
				// which is not expected
				keyInFrom != "inline" {
				if err := addMissingFields(valueInFrom.(map[string]interface{}), valueInTarget.(map[string]interface{}), keysToSkip); err != nil {
					return err
				}
			} else if valueInFrom != nil && reflect.TypeOf(valueInFrom).Kind() == reflect.Slice &&
				valueInTarget != nil && reflect.TypeOf(valueInTarget).Kind() == reflect.Slice &&
				keyInFrom == "additionalPackages" {
				// The following code explicitly handles ClusterBootstrap.Spec.AdditionalPackages(Golang slice). Not all
				// Golang slices need to be handled in the same way.
				/*
					Example of the use case:
						ClusterBootstrapTemplate has the following in AdditionalPackages:
						  - {refName: a, valuesFrom: {...}}
						  - {refName: b, valuesFrom: {...}}
						  - {refName: c, valuesFrom: {...}}

						ClusterBootstrap has the following in AdditionalPackages:
						  - {refName: a}
						  - {refName: c, valuesFrom: {...}}
						  - {refName: d}

						After addMissingFields(cbTemplate, cb, nil), we expect to get the following result in ClusterBootstrap:
						  - {refName: a, valuesFrom: {...}}
						  - {refName: b, valuesFrom: {...}}
						  - {refName: c, valuesFrom: {...}}
						  - {refName: d}
				*/

				sliceInFrom := valueInFrom.([]interface{})
				sliceInTarget := valueInTarget.([]interface{})
				var updatedSlice []interface{}
				for _, sliceItemInFrom := range sliceInFrom {
					foundMatch := false
					itemMapInFrom := sliceItemInFrom.(map[string]interface{})
					for _, sliceItemInTarget := range sliceInTarget {
						itemMapInTarget := sliceItemInTarget.(map[string]interface{})
						// If two slices items match(using refName to compare), we want to add missing fields from one to another.
						// e.g., {refName: a, valuesFrom: {...}} and {refName: a} matches
						if itemMapInFrom["refName"].(string) == itemMapInTarget["refName"].(string) {
							foundMatch = true
							copiedVal, _, copyErr := unstructured.NestedFieldCopy(itemMapInFrom)
							if copyErr != nil {
								return copyErr
							}
							// unstructured.NestedFieldCopy() copies all fields, but we need to delete some fields which are meant
							// to be skipped.
							if removeErr := removeFields(copiedVal.(map[string]interface{}), keysToSkip); removeErr != nil {
								return removeErr
							}
							if err := addMissingFields(copiedVal.(map[string]interface{}), itemMapInTarget, keysToSkip); err != nil {
								return err
							}
							updatedSlice = append(updatedSlice, itemMapInTarget)
							break // we expect only one match
						}
					}
					// If there is no match, we simply add the new slice item to the final slice
					if !foundMatch {
						copiedVal, _, copyErr := unstructured.NestedFieldCopy(itemMapInFrom)
						if copyErr != nil {
							return copyErr
						}
						// unstructured.NestedFieldCopy() copies all fields, but we need to delete some fields which are meant
						// to be skipped.
						if removeErr := removeFields(copiedVal.(map[string]interface{}), keysToSkip); removeErr != nil {
							return removeErr
						}
						updatedSlice = append(updatedSlice, copiedVal)
					}
				}
				destination[keyInFrom] = updatedSlice
			}
		}
	}
	return nil
}

// removeFields deletes the map entries by comparing the keys with keysToRemove
func removeFields(m, keysToRemove map[string]interface{}) error {
	if keysToRemove == nil || len(keysToRemove) == 0 {
		return nil
	}
	for key, value := range m {
		_, exist := keysToRemove[key]
		if exist {
			// Delete the field and continue to scan next key
			delete(m, key)
			continue
		}
		if value != nil && reflect.TypeOf(value).Kind() == reflect.Map {
			// If value is a map, recursively look into it
			if err := removeFields(value.(map[string]interface{}), keysToRemove); err != nil {
				return err
			}
		}
	}
	return nil
}

// HandleExistingClusterBootstrap does the following:
//  1. Check the annotation from the existing ClusterBootstrap CR. Add missing fields to the ClusterBootstrap CR if needed
//  2. Clone the referenced object of ClusterBootstrapPackage.ValuesFrom into the cluster namespace. If the referenced
//     object has embedded TypedLocalObjectReference(e.g., CPI has VSphereCPIConfig as valuesFrom, VSphereCPIConfig has
//     a Secret reference under its Spec), this function clones the embedded object into the cluster namespace as well.
//  3. Add OwnerReferences and Labels to the cloned objects.
func (h *Helper) HandleExistingClusterBootstrap(clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, cluster *clusterapiv1beta1.Cluster, tkrName, systemNamespace string) (*runtanzuv1alpha3.ClusterBootstrap, error) {
	packages := append([]*runtanzuv1alpha3.ClusterBootstrapPackage{
		clusterBootstrap.Spec.CNI,
		clusterBootstrap.Spec.CPI,
		clusterBootstrap.Spec.CSI,
		clusterBootstrap.Spec.Kapp,
	}, clusterBootstrap.Spec.AdditionalPackages...)

	if clusterBootstrap.Annotations != nil {
		if annotationValue, annotationExist := clusterBootstrap.Annotations[constants.AddCBMissingFieldsAnnotationKey]; annotationExist {
			tkrName = annotationValue
			clusterBootstrapTemplateName := annotationValue // TanzuKubernetesRelease and ClusterBootstrapTemplate share the same name
			clusterBootstrapTemplate := &runtanzuv1alpha3.ClusterBootstrapTemplate{}
			// Get the ClusterBootstrapTemplate mentioned in the annotation under system namespace
			if err := h.K8sClient.Get(h.Ctx, client.ObjectKey{Namespace: systemNamespace, Name: clusterBootstrapTemplateName}, clusterBootstrapTemplate); err != nil {
				return nil, err
			}

			// Update the packages slice to make it only contains the ones we want the rest of the code to do the cloning.
			// For cloning, we only care about the ClusterBootstrap packages which have no valuesFrom originally. The missing
			// valuesFrom field will be added by AddMissingSpecFieldsFromTemplate() in few lines below, and we want to clone
			// those packages from system namespace.
			var nilValuesFromPackages []*runtanzuv1alpha3.ClusterBootstrapPackage
			for _, cbPackage := range packages {
				if cbPackage != nil && cbPackage.ValuesFrom == nil {
					nilValuesFromPackages = append(nilValuesFromPackages, cbPackage)
				}
			}

			var nonEmptyInlinePackages []*runtanzuv1alpha3.ClusterBootstrapPackage
			for _, cbPackage := range packages {
				if cbPackage != nil && cbPackage.ValuesFrom != nil && cbPackage.ValuesFrom.Inline != nil {
					nonEmptyInlinePackages = append(nonEmptyInlinePackages, cbPackage)
				}
			}

			if err := h.AddMissingSpecFieldsFromTemplate(clusterBootstrapTemplate, clusterBootstrap, nil); err != nil {
				h.Logger.Error(err, fmt.Sprintf("unable to add missing spec fields of ClusterBootstrap %s/%s from ClusterBootstrapTemplate %s/%s",
					clusterBootstrap.Namespace, clusterBootstrap.Name, clusterBootstrapTemplate.Namespace, clusterBootstrapTemplate.Name))
				return nil, err
			}

			// AddMissingSpecFieldsFromTemplate() updates clusterBootstrap.Spec, it holds the updated pointers now.
			// Declaring updatedCBPackages is to compare with original nilValuesFromPackages and figure out which packages
			// need to be cloned.
			updatedCBPackages := append([]*runtanzuv1alpha3.ClusterBootstrapPackage{
				clusterBootstrap.Spec.CNI,
				clusterBootstrap.Spec.CPI,
				clusterBootstrap.Spec.CSI,
				clusterBootstrap.Spec.Kapp,
			}, clusterBootstrap.Spec.AdditionalPackages...)

			var packagesToBeCloned []*runtanzuv1alpha3.ClusterBootstrapPackage
			for _, nilValuesFromPackage := range nilValuesFromPackages {
				for _, updatedCBPackage := range updatedCBPackages {
					// If the nilValuesFromPackage has been filled by AddMissingSpecFieldsFromTemplate(), it is the package
					// we want to record and do cloning.
					if updatedCBPackage != nil && updatedCBPackage.ValuesFrom != nil && updatedCBPackage.RefName == nilValuesFromPackage.RefName {
						packagesToBeCloned = append(packagesToBeCloned, updatedCBPackage)
					}
				}
			}
			h.Logger.Info("Updating packagesToBeCloned with inline packages")
			packagesToBeCloned = append(packagesToBeCloned, nonEmptyInlinePackages...)
			packages = packagesToBeCloned
		}
	}

	secrets, providers, err := h.CloneReferencedObjectsFromCBPackages(cluster, packages, systemNamespace)
	if err != nil {
		h.Logger.Error(err, "unable to clone secrets or providers")
		return nil, err
	}

	clusterBootstrap.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         clusterapiv1beta1.GroupVersion.String(),
			Kind:               cluster.Kind,
			Name:               cluster.Name,
			UID:                cluster.UID,
			Controller:         pointer.BoolPtr(true),
			BlockOwnerDeletion: pointer.BoolPtr(true),
		},
	}
	if err := h.K8sClient.Update(h.Ctx, clusterBootstrap); err != nil {
		return nil, err
	}
	clusterBootstrap.Status.ResolvedTKR = tkrName
	if err := h.K8sClient.Status().Update(h.Ctx, clusterBootstrap); err != nil {
		return nil, err
	}

	// ensure ownerRef of clusterBootstrap on created secrets and providers, this can only be done after
	// clusterBootstrap is created
	if err := h.EnsureOwnerRef(clusterBootstrap, secrets, providers); err != nil {
		h.Logger.Error(err, fmt.Sprintf("unable to ensure ClusterBootstrap %s/%s as a ownerRef on created secrets and providers", clusterBootstrap.Namespace, clusterBootstrap.Name))
		return nil, err
	}

	return clusterBootstrap, nil
}

// CreateClusterBootstrapFromTemplate does the following:
//  1. Create a new ClusterBootstrap CR under the cluster namespace
//  2. Clone the referenced object of ClusterBootstrapPackage.ValuesFrom into the cluster namespace. If the referenced
//     object has embedded TypedLocalObjectReference(e.g., CPI has VSphereCPIConfig as valuesFrom, VSphereCPIConfig has
//     a Secret reference under its Spec), this function clones the embedded object into the cluster namespace as well.
//  3. Add OwnerReferences and Labels to the cloned objects.
func (h *Helper) CreateClusterBootstrapFromTemplate(
	clusterBootstrapTemplate *runtanzuv1alpha3.ClusterBootstrapTemplate,
	cluster *clusterapiv1beta1.Cluster,
	tkrName string) (*runtanzuv1alpha3.ClusterBootstrap, error) {

	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	clusterBootstrap.Name = cluster.Name
	clusterBootstrap.Namespace = cluster.Namespace
	clusterBootstrap.Spec = clusterBootstrapTemplate.Spec.DeepCopy()

	packages := append([]*runtanzuv1alpha3.ClusterBootstrapPackage{
		clusterBootstrap.Spec.CNI,
		clusterBootstrap.Spec.CPI,
		clusterBootstrap.Spec.CSI,
		clusterBootstrap.Spec.Kapp,
	}, clusterBootstrap.Spec.AdditionalPackages...)

	secrets, providers, err := h.CloneReferencedObjectsFromCBPackages(cluster, packages, clusterBootstrapTemplate.Namespace)
	if err != nil {
		h.Logger.Error(err, "unable to clone secrets or providers")
		return nil, err
	}

	clusterBootstrap.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         clusterapiv1beta1.GroupVersion.String(),
			Kind:               cluster.Kind,
			Name:               cluster.Name,
			UID:                cluster.UID,
			Controller:         pointer.BoolPtr(true),
			BlockOwnerDeletion: pointer.BoolPtr(true),
		},
	}

	if err := h.K8sClient.Create(h.Ctx, clusterBootstrap); err != nil {
		return nil, err
	}

	// ensure ownerRef of clusterBootstrap on created secrets and providers, this can only be done after
	// clusterBootstrap is created
	if err := h.EnsureOwnerRef(clusterBootstrap, secrets, providers); err != nil {
		h.Logger.Error(err, fmt.Sprintf("unable to ensure ClusterBootstrap %s/%s as a ownerRef on created secrets and providers", clusterBootstrap.Namespace, clusterBootstrap.Name))
		return nil, err
	}

	clusterBootstrap.Status.ResolvedTKR = tkrName
	if err := h.K8sClient.Status().Update(h.Ctx, clusterBootstrap); err != nil {
		h.Logger.Error(err, fmt.Sprintf("unable to update the status of ClusterBootstrap %s/%s", clusterBootstrap.Namespace, clusterBootstrap.Name))
		return nil, err
	}
	h.Logger.Info("created clusterBootstrap", "clusterBootstrap", clusterBootstrap)

	return clusterBootstrap, nil
}

// CloneReferencedObjectsFromCBPackages clones all referenced objects of a list of ClusterBootstrapPackage.ValuesFrom from
// sourceNamespace into the cluster namespace.
func (h *Helper) CloneReferencedObjectsFromCBPackages(
	cluster *clusterapiv1beta1.Cluster,
	cbPackages []*runtanzuv1alpha3.ClusterBootstrapPackage,
	sourceNamespace string) ([]*corev1.Secret, []*unstructured.Unstructured, error) {

	var createdProviders []*unstructured.Unstructured
	var createdSecrets []*corev1.Secret

	for _, cbPackage := range cbPackages {
		if cbPackage == nil {
			continue
		}

		// cbPackage.RefName is usually the same as Carvel Package.Metadata.Name which is <package.spec.refName>.<spec.version>
		// carvelPkgRefName is Carvel Package.Spec.RefName
		carvelPkgRefName, _, err := util.GetPackageMetadata(h.Ctx, h.AggregateAPIResourcesClient, cbPackage.RefName, cluster.Namespace)
		if carvelPkgRefName == "" || err != nil {
			// Package.Spec.RefName and Package.Spec.Version are required fields for Package CR. We do not expect them to be
			// empty and error should not happen when fetching them from a Package CR.
			h.Logger.Error(err, fmt.Sprintf("unable to fetch Package.Spec.RefName or Package.Spec.Version from Package %s/%s",
				cluster.Namespace, cbPackage.RefName))
			return nil, nil, err
		}

		secret, provider, err := h.cloneReferencedObjectsFromCBPackage(cluster, cbPackage, carvelPkgRefName, sourceNamespace)
		if err != nil {
			return nil, nil, err
		}
		if secret != nil {
			createdSecrets = append(createdSecrets, secret)
		}
		if provider != nil {
			createdProviders = append(createdProviders, provider)
		}
	}
	return createdSecrets, createdProviders, nil
}

// cloneReferencedObjectsFromCBPackage is an internal function clones the referenced objects of a single ClusterBootstrapPackage.ValuesFrom
// from sourceNamespace into the cluster namespace.
func (h *Helper) cloneReferencedObjectsFromCBPackage(
	cluster *clusterapiv1beta1.Cluster,
	clusterBootstrapPkg *runtanzuv1alpha3.ClusterBootstrapPackage,
	carvelPkgRefName string,
	sourceNamespace string) (*corev1.Secret, *unstructured.Unstructured, error) {

	if clusterBootstrapPkg.ValuesFrom == nil {
		// Nothing to be cloned
		return nil, nil, nil
	}

	if clusterBootstrapPkg.ValuesFrom.Inline != nil {
		secret, err := h.CreateSecretFromInline(cluster, clusterBootstrapPkg, carvelPkgRefName)
		if err != nil {
			return nil, nil, err
		}
		return secret, nil, nil
	}

	if clusterBootstrapPkg.ValuesFrom.SecretRef != "" {
		secret, err := h.cloneSecretRef(cluster, clusterBootstrapPkg, carvelPkgRefName, sourceNamespace)
		if err != nil {
			return nil, nil, err
		}
		return secret, nil, nil
	}

	if clusterBootstrapPkg.ValuesFrom.ProviderRef != nil {
		provider, err := h.cloneProviderRef(cluster, clusterBootstrapPkg, carvelPkgRefName, sourceNamespace)
		if err != nil {
			return nil, nil, err
		}
		return nil, provider, nil
	}
	return nil, nil, nil
}

// CreateSecretFromInline is an internal function creates a Secret resource from ClusterBootstrapPackage.ValuesFrom.Inline into the cluster namespace
func (h *Helper) CreateSecretFromInline(
	cluster *clusterapiv1beta1.Cluster,
	cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage,
	carvelPkgRefName string) (*corev1.Secret, error) {

	inlineSecret := &corev1.Secret{}
	inlineSecret.Name = util.GeneratePackageSecretName(cluster.Name, carvelPkgRefName)
	// The secret will be created or patched under tkg-system namespace on remote cluster
	inlineSecret.Namespace = cluster.Namespace

	return h.CreateOrPatchInlineSecret(cluster, cbPkg, inlineSecret)
}

// CreateOrPatchInlineSecret is a function that creates or patches a Secret resource from ClusterBootstrapPackage.ValuesFrom.Inline into the cluster namespace
func (h *Helper) CreateOrPatchInlineSecret(
	cluster *clusterapiv1beta1.Cluster,
	cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage,
	inlineSecret *corev1.Secret) (*corev1.Secret, error) {

	inlineSecret.Type = corev1.SecretTypeOpaque
	opResult, createOrPatchErr := controllerutil.CreateOrPatch(h.Ctx, h.K8sClient, inlineSecret, func() error {
		inlineSecret.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			},
		}
		if inlineSecret.StringData == nil {
			inlineSecret.StringData = make(map[string]string)
		}
		inlineConfigYamlBytes, err := yaml.Marshal(cbPkg.ValuesFrom.Inline)
		if err != nil {
			h.Logger.Error(err, "unable to marshal inline ValuesFrom")
			return err
		}
		inlineSecret.StringData[constants.TKGDataValueFileName] = string(inlineConfigYamlBytes)

		// Add cluster and package labels to cloned secrets
		if inlineSecret.Labels == nil {
			inlineSecret.Labels = make(map[string]string)
		}
		inlineSecret.Labels[addontypes.PackageNameLabel] = util.ParseStringForLabel(cbPkg.RefName)
		inlineSecret.Labels[addontypes.ClusterNameLabel] = cluster.Name
		// Set secret.Type to ClusterBootstrapManagedSecret to enable us to Watch these secrets
		inlineSecret.Type = constants.ClusterBootstrapManagedSecret
		return nil
	})
	if createOrPatchErr != nil {
		return nil, createOrPatchErr
	}
	h.Logger.Info(fmt.Sprintf("Secret %s/%s for inline ValuesFrom %s", inlineSecret.Namespace, inlineSecret.Name, opResult))
	return inlineSecret, nil
}

// cloneSecretRef is an internal function clones the referenced objects of a single ClusterBootstrapPackage.ValuesFrom.SecretRef
// from sourceNamespace into the cluster namespace.
func (h *Helper) cloneSecretRef(
	cluster *clusterapiv1beta1.Cluster,
	cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage,
	carvelPkgRefName string,
	sourceNamespace string) (*corev1.Secret, error) {

	if cbPkg.ValuesFrom.SecretRef == "" {
		return nil, nil
	}

	secret := &corev1.Secret{}
	key := client.ObjectKey{Namespace: sourceNamespace, Name: cbPkg.ValuesFrom.SecretRef}
	if err := h.K8sClient.Get(h.Ctx, key, secret); err != nil {
		h.Logger.Error(err, "unable to fetch secret", "objectkey", key)
		return nil, err
	}

	newSecret := secret.DeepCopy()
	newSecret.ObjectMeta.Reset()
	newSecret.Name = util.GeneratePackageSecretName(cluster.Name, carvelPkgRefName)
	newSecret.Namespace = cluster.Namespace

	opResult, createOrPatchErr := controllerutil.CreateOrPatch(h.Ctx, h.K8sClient, newSecret, func() error {
		newSecret.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			},
		}

		// Add cluster and package labels to cloned secrets
		if newSecret.Labels == nil {
			newSecret.Labels = map[string]string{}
		}
		newSecret.Labels[addontypes.PackageNameLabel] = util.ParseStringForLabel(cbPkg.RefName)
		newSecret.Labels[addontypes.ClusterNameLabel] = cluster.Name
		// Set secret.Type to ClusterBootstrapManagedSecret to enable clusterbootstrap_controller to Watch these secrets
		newSecret.Type = constants.ClusterBootstrapManagedSecret
		return nil
	})
	if createOrPatchErr != nil {
		return nil, createOrPatchErr
	}
	h.Logger.Info(fmt.Sprintf("Secret %s/%s %s", newSecret.Namespace, newSecret.Name, opResult))
	cbPkg.ValuesFrom.SecretRef = newSecret.Name

	return newSecret, nil
}

// cloneProviderRef is an internal function clones the referenced objects of a single ClusterBootstrapPackage.ValuesFrom.ProviderRef
// from sourceNamespace into the cluster namespace.
func (h *Helper) cloneProviderRef(
	cluster *clusterapiv1beta1.Cluster,
	cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage,
	carvelPkgRefName string,
	sourceNamespace string) (*unstructured.Unstructured, error) {

	if cbPkg.ValuesFrom == nil {
		return nil, nil
	}

	var newProvider *unstructured.Unstructured
	var createdOrUpdatedProvider *unstructured.Unstructured
	gvr, err := h.GVRHelper.GetGVR(schema.GroupKind{Group: *cbPkg.ValuesFrom.ProviderRef.APIGroup, Kind: cbPkg.ValuesFrom.ProviderRef.Kind})
	if err != nil {
		h.Logger.Error(err, "failed to getGVR")
		return nil, err
	}
	provider, err := h.DynamicClient.Resource(*gvr).Namespace(sourceNamespace).Get(h.Ctx, cbPkg.ValuesFrom.ProviderRef.Name, metav1.GetOptions{})
	if err != nil {
		h.Logger.Error(err, fmt.Sprintf("unable to fetch provider %s/%s", sourceNamespace, cbPkg.ValuesFrom.ProviderRef.Name), "gvr", gvr)
		return nil, err
	}
	newProvider = provider.DeepCopy()
	unstructured.RemoveNestedField(newProvider.Object, "metadata")
	newProvider.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: clusterapiv1beta1.GroupVersion.String(),
			Kind:       cluster.Kind,
			Name:       cluster.Name,
			UID:        cluster.UID,
		},
	})
	// Add cluster and package labels to cloned providers
	providerLabels := newProvider.GetLabels()
	if providerLabels == nil {
		newProvider.SetLabels(map[string]string{
			// A package refName could contain characters that K8S does not like as a label value.
			// For example, kapp-controller.tanzu.vmware.com.0.30.0+vmware.1-tkg.1 is a
			// valid package refName, but it contains "+" that K8S complains. We parse the refName by replacing
			// + to ---.
			addontypes.PackageNameLabel: util.ParseStringForLabel(cbPkg.RefName),
			addontypes.ClusterNameLabel: cluster.Name,
		})
	} else {
		providerLabels[addontypes.PackageNameLabel] = util.ParseStringForLabel(cbPkg.RefName)
		providerLabels[addontypes.ClusterNameLabel] = cluster.Name
		newProvider.SetLabels(providerLabels)
	}

	newProvider.SetName(util.GeneratePackageSecretName(cluster.Name, carvelPkgRefName))
	newProvider.SetNamespace(cluster.Namespace)
	h.Logger.Info(fmt.Sprintf("cloning provider %s/%s to namespace %s", sourceNamespace, newProvider.GetName(), cluster.Namespace), "gvr", gvr)
	createdOrUpdatedProvider, err = h.DynamicClient.Resource(*gvr).Namespace(cluster.Namespace).Create(h.Ctx, newProvider, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			h.Logger.Info(fmt.Sprintf("provider %s/%s already exist, patching its Labels and OwnerReferences fields", newProvider.GetNamespace(), newProvider.GetName()))
			// Instantiate an empty unstructured and only set ownerReferences and Labels for patching
			patchObj := unstructured.Unstructured{}
			patchObj.SetLabels(newProvider.GetLabels())
			patchObj.SetOwnerReferences(newProvider.GetOwnerReferences())
			patchData, err := patchObj.MarshalJSON()
			if err != nil {
				h.Logger.Error(err, fmt.Sprintf("unable to patch provider %s/%s", newProvider.GetNamespace(), newProvider.GetName()), "gvr", gvr)
				return nil, err
			}
			createdOrUpdatedProvider, err = h.DynamicClient.Resource(*gvr).Namespace(cluster.Namespace).Patch(h.Ctx, newProvider.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
			if err != nil {
				h.Logger.Info(fmt.Sprintf("unable to update provider %s/%s", newProvider.GetNamespace(), newProvider.GetName()), "gvr", gvr)
				return nil, err
			}
		} else {
			h.Logger.Error(err, fmt.Sprintf("unable to clone provider %s/%s", newProvider.GetNamespace(), newProvider.GetName()), "gvr", gvr)
			return nil, err
		}
	}

	cbPkg.ValuesFrom.ProviderRef.Name = createdOrUpdatedProvider.GetName()
	h.Logger.Info(fmt.Sprintf("cloned provider %s/%s to namespace %s", createdOrUpdatedProvider.GetNamespace(), createdOrUpdatedProvider.GetName(), cluster.Namespace), "gvr", gvr)

	if err := h.cloneEmbeddedLocalObjectRef(cluster, provider); err != nil {
		return nil, err
	}

	return createdOrUpdatedProvider, nil
}

// cloneEmbeddedLocalObjectRef is an internal function attempts to clone the embedded local object references from provider's namespace to cluster's
// namespace. An example of embedded local object reference is the secret reference under CPIConfig.
func (h *Helper) cloneEmbeddedLocalObjectRef(cluster *clusterapiv1beta1.Cluster, provider *unstructured.Unstructured) error {
	groupKindNamesMap := util.ExtractTypedLocalObjectRef(provider.UnstructuredContent(), constants.LocalObjectRefSuffix)
	if len(groupKindNamesMap) == 0 {
		return nil
	}

	providerGVK := provider.GroupVersionKind()
	h.Logger.Info(fmt.Sprintf("cloning the embedded local object references within provider: %s with name: %s from"+
		" %s namespace to %s namespace", provider.GroupVersionKind().String(), provider.GetName(), provider.GetNamespace(), cluster.Namespace))
	for groupKind, resourceNames := range groupKindNamesMap {
		gvr, err := h.GVRHelper.GetGVR(groupKind)
		if err != nil {
			// error has been logged within getGVR()
			return err
		}
		for _, resourceName := range resourceNames {
			h.Logger.Info(fmt.Sprintf("cloning the GVR %s with name %s from %s namespace to %s namespace",
				gvr.String(), resourceName, provider.GetNamespace(), cluster.Namespace))
			fetchedObj, err := h.DynamicClient.Resource(*gvr).Namespace(provider.GetNamespace()).Get(h.Ctx, resourceName, metav1.GetOptions{})
			if err != nil {
				h.Logger.Error(err, fmt.Sprintf("unable to get provider: %s with name: %s under namespace: %s",
					providerGVK.String(), provider.GetName(), provider.GetNamespace()))
				return err
			}

			copiedObj := fetchedObj.DeepCopy()
			// Remove resourceVersion to make sure dynamicClient create could work. We should not set resourceVersion on object creation
			unstructured.RemoveNestedField(copiedObj.Object, "metadata", "resourceVersion")
			copiedObj.SetNamespace(cluster.Namespace)
			ownerReferences := copiedObj.GetOwnerReferences()
			ownerReferences = clusterapiutil.EnsureOwnerRef(ownerReferences, metav1.OwnerReference{
				APIVersion: provider.GetAPIVersion(),
				Kind:       provider.GetKind(),
				Name:       provider.GetName(),
				UID:        provider.GetUID(),
			})
			ownerReferences = clusterapiutil.EnsureOwnerRef(ownerReferences, metav1.OwnerReference{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       "Cluster",
				Name:       cluster.GetName(),
				UID:        cluster.GetUID(),
			})
			copiedObj.SetOwnerReferences(ownerReferences)
			_, err = h.DynamicClient.Resource(*gvr).Namespace(cluster.Namespace).Create(h.Ctx, copiedObj, metav1.CreateOptions{})
			if err != nil {
				if apierrors.IsAlreadyExists(err) {
					// Only patch the ownerReferences
					patchObj := unstructured.Unstructured{}
					patchObj.SetOwnerReferences(ownerReferences)
					var jsonData []byte
					if jsonData, err = patchObj.MarshalJSON(); err != nil {
						return err
					}
					_, err = h.DynamicClient.Resource(*gvr).Namespace(cluster.Namespace).Patch(h.Ctx, copiedObj.GetName(), types.MergePatchType, jsonData, metav1.PatchOptions{})
					if err != nil {
						h.Logger.Error(err, fmt.Sprintf("unable to clone the GVR %s with name %s from namespace %s to"+
							" target namespace %s", gvr.String(), resourceName, provider.GetNamespace(), cluster.Namespace))
						return err
					}
				} else {
					h.Logger.Error(err, fmt.Sprintf("unable to clone the GVR %s with name %s from namespace %s to"+
						" target namespace %s", gvr.String(), resourceName, provider.GetNamespace(), cluster.Namespace))
					return err
				}
			}
		}
	}
	h.Logger.Info(fmt.Sprintf("cloned the embedded local object references within provider: %s with name: %s under"+
		" namespace: %s to target namespace %s", providerGVK.String(), provider.GetName(), provider.GetNamespace(), cluster.Namespace))
	return nil
}

// EnsureOwnerRef will ensure the provided OwnerReference onto the secrets and provider objects
func (h *Helper) EnsureOwnerRef(clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, secrets []*corev1.Secret, providers []*unstructured.Unstructured) error {
	ownerRef := metav1.OwnerReference{
		APIVersion:         runtanzuv1alpha3.GroupVersion.String(),
		Kind:               "ClusterBootstrap", // kind is empty after create
		Name:               clusterBootstrap.Name,
		UID:                clusterBootstrap.UID,
		Controller:         pointer.BoolPtr(true),
		BlockOwnerDeletion: pointer.BoolPtr(true),
	}

	for _, secret := range secrets {
		ownerRefsMutateFn := func() error {
			secret.OwnerReferences = clusterapiutil.EnsureOwnerRef(secret.OwnerReferences, ownerRef)
			return nil
		}
		_, err := controllerutil.CreateOrPatch(h.Ctx, h.K8sClient, secret, ownerRefsMutateFn)
		if err != nil {
			h.Logger.Error(err, fmt.Sprintf("unable to create or patch the secret %s/%s with ownerRef", secret.Namespace, secret.Name))
			return err
		}
	}
	for _, provider := range providers {
		gvr, err := h.GVRHelper.GetGVR(provider.GroupVersionKind().GroupKind())
		if err != nil {
			h.Logger.Error(err, fmt.Sprintf("unable to get GVR of provider %s/%s", provider.GetNamespace(), provider.GetName()))
			return err
		}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// We need to get and update, otherwise there could have concurrency issue: ["the object has been modified; please
			// apply your changes to the latest version and try again"]
			newProvider, errGetProvider := h.DynamicClient.Resource(*gvr).Namespace(provider.GetNamespace()).Get(h.Ctx, provider.GetName(), metav1.GetOptions{})
			if errGetProvider != nil {
				h.Logger.Error(errGetProvider, fmt.Sprintf("unable to get provider %s/%s", provider.GetNamespace(), provider.GetName()))
				return errGetProvider
			}
			newProvider = newProvider.DeepCopy()
			newProvider.SetOwnerReferences(clusterapiutil.EnsureOwnerRef(provider.GetOwnerReferences(), ownerRef))
			_, errUpdateProvider := h.DynamicClient.Resource(*gvr).Namespace(newProvider.GetNamespace()).Update(h.Ctx, newProvider, metav1.UpdateOptions{})
			if errUpdateProvider != nil {
				h.Logger.Error(errUpdateProvider, fmt.Sprintf("unable to update provider %s/%s", provider.GetNamespace(), provider.GetName()))
				return errUpdateProvider
			}
			return nil
		})
		if err != nil {
			h.Logger.Error(err, fmt.Sprintf("unable to update the OwnerRefrences for provider %s/%s", provider.GetNamespace(), provider.GetName()))
			return err
		}
	}
	return nil
}

// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// GetTKRByNameV1Alpha1 gets v1Alpha1 TKR object given a TKR name
func GetTKRByNameV1Alpha1(ctx context.Context, c client.Client, tkrName string) (*runtanzuv1alpha1.TanzuKubernetesRelease, error) {
	tkrV1Alpha1 := &runtanzuv1alpha1.TanzuKubernetesRelease{}

	if tkrName == "" {
		return nil, nil
	}

	tkrNamespaceName := client.ObjectKey{Name: tkrName}

	if err := c.Get(ctx, tkrNamespaceName, tkrV1Alpha1); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return tkrV1Alpha1, nil
}

// GetTKRByNameV1Alpha3 gets v1Alpha3 TKR object given a TKR name
func GetTKRByNameV1Alpha3(ctx context.Context, c client.Client, cluster *clusterapiv1beta1.Cluster, tkrName string) (*runtanzuv1alpha3.TanzuKubernetesRelease, error) {
	tkrV1Alpha3 := &runtanzuv1alpha3.TanzuKubernetesRelease{}

	if tkrName == "" {
		return nil, nil
	}

	tkrNamespaceName := client.ObjectKey{Name: tkrName}

	if err := c.Get(ctx, tkrNamespaceName, tkrV1Alpha3); err != nil {
		if apierrors.IsNotFound(err) {
			tkrV1Alpha3, err = getTKRFromAnnotation(cluster.Annotations)
			if tkrV1Alpha3 != nil {
				return tkrV1Alpha3, nil
			}

			return nil, nil
		}
		return nil, err
	}
	return tkrV1Alpha3, nil
}

// GetBootstrapPackageNameFromTKR tries to find the prefix of the provided package RefName in the bootstrap packages of the TKR v1Alpha3 object associated
// with the cluster. Upon finding the corresponding bootstrap package name, it returns it as the bumped version of the package
func GetBootstrapPackageNameFromTKR(ctx context.Context, clt client.Client, pkgRefName string, cluster *clusterapiv1beta1.Cluster) (string, string, error) {
	pkgNamePrefix := pkgRefName
	pkgNameTokens := strings.Split(pkgRefName, ".")
	if len(pkgNameTokens) >= 1 {
		pkgNamePrefix = pkgNameTokens[0]
	}

	// it is expected to have a label corresponding to the TKR name in the cluster object
	tkrName := GetClusterLabel(cluster.Labels, constants.TKRLabelClassyClusters)
	if tkrName == "" {
		return "", pkgNamePrefix, fmt.Errorf("no '%s' label found in the cluster object", constants.TKRLabelClassyClusters)
	}

	// get TKR object associated with the cluster
	tkr, err := GetTKRByNameV1Alpha3(ctx, clt, cluster, tkrName)
	if err != nil || tkr == nil {
		return "", pkgNamePrefix, fmt.Errorf("unable to fetch TKR object '%s'", tkrName)
	}

	tkrBootstrapPackages := tkr.Spec.BootstrapPackages
	if len(tkrBootstrapPackages) == 0 {
		return "", pkgNamePrefix, errors.New("unable to find any bootstrap packages in the TKR object")
	}

	for _, bootstrapPackage := range tkrBootstrapPackages {
		if strings.HasPrefix(bootstrapPackage.Name, pkgNamePrefix) {
			return bootstrapPackage.Name, pkgNamePrefix, nil
		}
	}

	return "", pkgNamePrefix, fmt.Errorf("no bootstrap package prefixed with '%s' is found in the TKR object", pkgNamePrefix)
}

func getTKRFromAnnotation(tkc map[string]string) (*runtanzuv1alpha3.TanzuKubernetesRelease, error) {
	controlPlaneTKR := &runtanzuv1alpha3.TanzuKubernetesRelease{}

	tkrSpec, ok := tkc[constants.TKRAnnotationKey]
	if !ok {
		return nil, nil
	}

	unzipOut, err := gunzipAndBase64Decode(tkrSpec)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(unzipOut, controlPlaneTKR)
	if err != nil {
		return nil, err
	}

	return controlPlaneTKR, nil
}

// GunzipAndBase64Decode extracts a gzip archive to a byte array
func gunzipAndBase64Decode(input string) ([]byte, error) {
	decodedData, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}

	r, err := gzip.NewReader(bytes.NewReader(decodedData))
	if err != nil {
		return nil, err
	}

	return io.ReadAll(r)
}

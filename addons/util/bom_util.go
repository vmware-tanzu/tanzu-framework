package util

import (
	"context"
	"github.com/vmware-tanzu-private/core/addons/constants"
	bomv1alpha1 "github.com/vmware-tanzu-private/core/apis/bom/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/bom"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetBOMByTKRName returns the bom associated with the TKR
func GetBOMByTKRName(ctx context.Context, c client.Client, tkrName string) (*bomv1alpha1.BomConfig, error) {
	configMapList := &corev1.ConfigMapList{}
	var bomConfigMap *corev1.ConfigMap
	if err := c.List(ctx, configMapList, client.InNamespace(constants.TKG_BOM_NAMESPACE), client.MatchingLabels{constants.TKR_LABEL: tkrName}); err != nil {
		return nil, err
	}

	if len(configMapList.Items) <= 0 {
		return nil, nil
	}

	bomConfigMap = &configMapList.Items[0]
	bomData, ok := bomConfigMap.Data["bomContent"]
	if !ok || bomData == "" {
		return nil, nil
	}

	bomConfig, err := bom.UnmarshalBOM([]byte(bomData))
	if err != nil {
		return nil, err
	}

	return bomConfig, nil
}

// GetTKRNameFromBOMConfigMap returns tkr name given a bom configmap
func GetTKRNameFromBOMConfigMap(bomConfigMap *corev1.ConfigMap) string {
	return bomConfigMap.Labels[constants.TKR_LABEL]
}

// GetAddonConfigFromBom gets addon config from BOM matching addon name
func GetAddonConfigFromBom(addonName string, bomConfig *bomv1alpha1.BomConfig) *bomv1alpha1.BomAddon {
	if addonName == "" {
		return nil
	}

	bomAddon, ok := bomConfig.Addons[addonName]
	if !ok {
		return nil
	}

	return &bomAddon
}

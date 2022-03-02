package utils

import (
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/constants"
	corev1 "k8s.io/api/core/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// IsAddonType returns true if the secret is type `tkg.tanzu.vmware.com/addon`
func IsAddonType(secret *corev1.Secret) bool {
	return secret.Type == constants.TKGAddonType
}

// HasAddonLabel returns true if the `tkg.tanzu.vmware.com/addon` label matches the parameter we pass in
func HasAddonLabel(secret *corev1.Secret, label string) bool {
	return secret.Labels[constants.TKGAddonLabel] == label
}
// IsManagementCluster returns true if the cluster has the "cluster-role.tkg.tanzu.vmware.com/management" label
func IsManagementCluster(cluster clusterapiv1beta1.Cluster) bool {
	_, labelExists := cluster.GetLabels()[constants.TKGManagementLabel]
	return labelExists
}
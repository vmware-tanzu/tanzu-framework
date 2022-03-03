package utils

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/constants"
	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	corev1 "k8s.io/api/core/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
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

// GetInfraProvider get infrastructure kind from cluster spec
func GetInfraProvider(cluster clusterapiv1beta1.Cluster) (string, error) {
	var infraProvider string

	infrastructureRef := cluster.Spec.InfrastructureRef
	if infrastructureRef == nil {
		return "", fmt.Errorf("cluster.Spec.InfrastructureRef is not set for cluster '%s", cluster.Name)
	}

	switch infrastructureRef.Kind {
	case tkgconstants.InfrastructureRefVSphere:
		infraProvider = tkgconstants.InfrastructureProviderVSphere
	case tkgconstants.InfrastructureRefAWS:
		infraProvider = tkgconstants.InfrastructureProviderAWS
	case tkgconstants.InfrastructureRefAzure:
		infraProvider = tkgconstants.InfrastructureProviderAzure
	case constants.InfrastructureRefDocker:
		infraProvider = tkgconstants.InfrastructureProviderDocker
	default:
		return "", fmt.Errorf("unknown cluster.Spec.InfrastructureRef.Kind is set for cluster '%s", cluster.Name)
	}

	return infraProvider, nil
}

// ClusterHasLabel checks if the cluster has the given label... copy from predicates/tkr.go.... could we use that directly?
func ClusterHasLabel(label string, logger logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return processIfClusterHasLabel(label, e.ObjectNew, logger.WithValues("predicate", "updateEvent"))
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return processIfClusterHasLabel(label, e.Object, logger.WithValues("predicate", "createEvent"))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return processIfClusterHasLabel(label, e.Object, logger.WithValues("predicate", "deleteEvent"))
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return processIfClusterHasLabel(label, e.Object, logger.WithValues("predicate", "genericEvent"))
		},
	}
}

// processIfClusterHasLabel determines if the input object is a cluster with a non-empty
// value for the specified label. For other input object types, it returns true
func processIfClusterHasLabel(label string, obj client.Object, logger logr.Logger) bool {
	kind := obj.GetObjectKind().GroupVersionKind().Kind

	if kind != constants.ClusterKind {
		return true
	}

	labels := obj.GetLabels()
	if labels != nil {
		if l, ok := labels[label]; ok && l != "" {
			return true
		}
	}

	log := logger.WithValues("namespace", obj.GetNamespace(), strings.ToLower(kind), obj.GetName())
	log.V(6).Info("Cluster resource does not have label", "label", label)
	return false
}

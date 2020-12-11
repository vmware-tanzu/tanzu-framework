package predicates

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/vmware-tanzu-private/core/addons/constants"
	"k8s.io/apimachinery/pkg/runtime"
	controlplanev1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// processKubeadmControlPlaneUpdate detects if kcp update event needs to be processed or not
// kcp event is process if there is a kcp.spec.version change detected between old and new objects
func processKubeadmControlPlaneUpdate(old runtime.Object, new runtime.Object, log logr.Logger) bool {
	var (
		kcpOld *controlplanev1alpha3.KubeadmControlPlane
		kcpNew *controlplanev1alpha3.KubeadmControlPlane
	)

	switch obj := old.(type) {
	case *controlplanev1alpha3.KubeadmControlPlane:
		kcpOld = obj
	default:
		// Defaults to true so we don't filter out other objects as the
		// filters are global
		log.Info("Expected object type of kcp. Got object type", "actualType", fmt.Sprintf("%T", obj))
		return true
	}

	switch obj := new.(type) {
	case *controlplanev1alpha3.KubeadmControlPlane:
		kcpNew = obj
	default:
		// Defaults to true so we don't filter out other objects as the
		// filters are global
		log.Info("Expected object type of kcp. Got object type", "actualType", fmt.Sprintf("%T", obj))
		return true
	}

	// If there is a version change (upgrade or downgrade) then return true for reconciling kcp
	if kcpNew.Spec.Version != kcpOld.Spec.Version {
		return true
	}

	log.V(7).Info("KCP version change not detected",
		constants.NAMESPACE_LOG_KEY, kcpNew.Namespace, constants.NAME_LOG_KEY, kcpNew.Name,
		"old-version", kcpOld.Spec.Version, "new-version", kcpNew.Spec.Version)

	return false
}

// KubeadmControlPlane returns a predicate.Predicate that detects kubeadm
// control plane version changes i.e. upgrade of a cluster
func KubeadmControlPlane(log logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		UpdateFunc:  func(e event.UpdateEvent) bool { return processKubeadmControlPlaneUpdate(e.ObjectOld, e.ObjectNew, log) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
}

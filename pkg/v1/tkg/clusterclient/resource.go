// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterclient

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// PostVerifyrFunc is a function which should be used as closure
type PostVerifyrFunc func(obj crtclient.Object) error

// PostVerifyListrFunc is a function which should be used as closure
type PostVerifyListrFunc func(obj crtclient.ObjectList) error

func (c *client) ListResources(resourceReference interface{}, option ...crtclient.ListOption) error {
	obj, err := getRuntimeObjectList(resourceReference)
	if err != nil {
		return err
	}
	if err := c.clientSet.List(context.TODO(), obj, option...); err != nil {
		return errors.Wrapf(err, "failed to list %v", reflect.TypeOf(resourceReference))
	}
	return nil
}

func (c *client) DeleteResource(resourceReference interface{}) error {
	obj, err := getRuntimeObject(resourceReference)
	if err != nil {
		return err
	}
	if err := c.clientSet.Delete(ctx, obj); err != nil {
		return errors.Wrapf(err, "failed to delete %v", reflect.TypeOf(resourceReference))
	}
	return nil
}

func (c *client) CreateResource(resourceReference interface{}, resourceName, namespace string, opts ...crtclient.CreateOption) error {
	obj, err := getRuntimeObject(resourceReference)
	if err != nil {
		return err
	}

	if err := c.clientSet.Create(ctx, obj, opts...); err != nil {
		return errors.Wrapf(err, "error while creating object for %q %s/%s",
			obj.GetObjectKind(), namespace, resourceName)
	}
	return nil
}

func (c *client) UpdateResourceWithPolling(resourceReference interface{}, resourceName, namespace string, pollOptions *PollOptions, opts ...crtclient.UpdateOption) error {
	if pollOptions != nil {
		log.V(6).Infof("Updating resource %s of type %s ...", resourceName, reflect.TypeOf(resourceReference))
		_, err := c.poller.PollImmediateWithGetter(pollOptions.Interval, pollOptions.Timeout, func() (interface{}, error) {
			return nil, c.UpdateResource(resourceReference, resourceName, namespace, opts...)
		})
		return err
	}

	return c.UpdateResource(resourceReference, resourceName, namespace, opts...)
}

func (c *client) UpdateResource(resourceReference interface{}, resourceName, namespace string, opts ...crtclient.UpdateOption) error {
	obj, err := getRuntimeObject(resourceReference)
	if err != nil {
		return err
	}

	if err := c.clientSet.Update(ctx, obj, opts...); err != nil {
		return errors.Wrapf(err, "error while creating object for %q %s/%s",
			obj.GetObjectKind(), namespace, resourceName)
	}
	return nil
}

func (c *client) PatchResource(resourceReference interface{}, resourceName, namespace, patchJSONString string, patchType types.PatchType, pollOptions *PollOptions) error {
	// if pollOptions are provided use the polling and wait for the result/error/timeout
	// else use normal get
	if pollOptions != nil {
		log.V(6).Infof("Applying patch to resource %s of type %s ...", resourceName, reflect.TypeOf(resourceReference))
		_, err := c.poller.PollImmediateWithGetter(pollOptions.Interval, pollOptions.Timeout, func() (interface{}, error) {
			return nil, c.patchResource(resourceReference, resourceName, namespace, patchJSONString, patchType)
		})
		return err
	}

	return c.patchResource(resourceReference, resourceName, namespace, patchJSONString, patchType)
}

func (c *client) patchResource(resourceReference interface{}, resourceName, namespace, patchJSONString string, patchType types.PatchType) error {
	patch := crtclient.RawPatch(patchType, []byte(patchJSONString))

	obj, err := getRuntimeObject(resourceReference)
	if err != nil {
		return err
	}
	clusterObjKey := crtclient.ObjectKey{
		Namespace: namespace,
		Name:      resourceName,
	}

	if err := c.clientSet.Get(ctx, clusterObjKey, obj); err != nil {
		return errors.Wrapf(err, "error reading %q %s/%s",
			obj.GetObjectKind(), namespace, resourceName)
	}

	if err := c.clientSet.Patch(ctx, obj, patch); err != nil {
		return errors.Wrapf(err, "error while applying patch for %q %s/%s",
			obj.GetObjectKind(), namespace, resourceName)
	}
	return nil
}

func (c *client) get(objectName, namespace string, o interface{}, postVerify PostVerifyrFunc) error {
	obj, err := getRuntimeObject(o)
	if err != nil {
		return err
	}
	objKey := crtclient.ObjectKey{Name: objectName, Namespace: namespace}
	if err := c.clientSet.Get(ctx, objKey, obj); err != nil {
		return err
	}
	if postVerify != nil {
		return postVerify(obj)
	}
	return nil
}

func (c *client) list(clusterName, namespace string, o interface{}, postVerify PostVerifyListrFunc) error {
	obj, err := getRuntimeObjectList(o)
	if err != nil {
		return err
	}

	selectors := []crtclient.ListOption{
		crtclient.InNamespace(namespace),
		crtclient.MatchingLabels(map[string]string{capi.ClusterLabelName: clusterName}),
	}

	if err := c.clientSet.List(ctx, obj, selectors...); err != nil {
		return err
	}

	if postVerify != nil {
		return postVerify(obj)
	}
	return nil
}

func getRuntimeObject(o interface{}) (crtclient.Object, error) {
	obj, ok := o.(crtclient.Object)
	if !ok {
		return nil, errors.New("invalid object type")
	}
	return obj, nil
}

func getRuntimeObjectList(o interface{}) (crtclient.ObjectList, error) {
	obj, ok := o.(crtclient.ObjectList)
	if !ok {
		return nil, errors.New("invalid object type")
	}
	return obj, nil
}

// VerifyClusterInitialized verifies the cluster is initialized or not (this is required before reading the kubeconfig secret)
func VerifyClusterInitialized(obj crtclient.Object) error {
	switch cluster := obj.(type) {
	case *capi.Cluster:
		errList := []error{}
		if !conditions.IsTrue(cluster, capi.ControlPlaneReadyCondition) {
			reason := conditions.GetReason(cluster, capi.ReadyCondition)
			errList = append(errList, fmt.Errorf("cluster control plane is still being initialized: %s", reason))
		}

		// Nb. We are verifying infrastructure ready at this stage because it provides an early signal that the infrastructure provided is
		// properly working, but this is not strictly required for getting the kubeconfig secret
		if !conditions.IsTrue(cluster, capi.InfrastructureReadyCondition) {
			reason := conditions.GetReason(cluster, capi.ReadyCondition)
			errList = append(errList, fmt.Errorf("cluster infrastructure is still being provisioned: %s", reason))
		}
		return kerrors.NewAggregate(errList)
	default:
		return errors.Errorf("invalid type: %s during VerifyClusterInitialized", reflect.TypeOf(cluster))
	}
}

// VerifyClusterReady verifies the cluster is ready or not (this is required before starting the move operation)
func VerifyClusterReady(obj crtclient.Object) error {
	// Nb. Currently there is no difference between VerifyClusterReady and VerifyClusterInitialized unless WorkersReady condition
	// would be added to cluster `Ready` condition aggregation.
	return VerifyClusterInitialized(obj)
}

// VerifyMachinesReady verifies the machine are ready or not (this is required before starting the move operation)
func VerifyMachinesReady(obj crtclient.ObjectList) error {
	switch machines := obj.(type) {
	case *capi.MachineList:
		errList := []error{}
		// Checking all the machine have a NodeRef
		// Nb. NodeRef is considered a better signal than InfrastructureReady, because it ensures the node in the workload cluster is up and running.
		for i := range machines.Items {
			if machines.Items[i].Status.NodeRef == nil {
				errList = append(errList, errors.Errorf("machine %s is still being provisioned", machines.Items[i].Name))
			}
		}
		return kerrors.NewAggregate(errList)
	default:
		return errors.Errorf("invalid type: %s during VerifyMachinesReady", reflect.TypeOf(machines))
	}
}

// VerifyKubeadmControlPlaneReplicas verifies the KubeadmControlPlane has all the required replicas (this is required before starting the move operation)
func VerifyKubeadmControlPlaneReplicas(obj crtclient.ObjectList) error {
	switch kcps := obj.(type) {
	case *controlplanev1.KubeadmControlPlaneList:
		errList := []error{}
		for i := range kcps.Items {
			var desiredReplica int32 = 1
			if kcps.Items[i].Spec.Replicas != nil {
				desiredReplica = *kcps.Items[i].Spec.Replicas
			}
			if desiredReplica != kcps.Items[i].Status.ReadyReplicas {
				errList = append(errList, errors.Errorf("control-plane is still creating replicas, DesiredReplicas=%v Replicas=%v ReadyReplicas=%v UpdatedReplicas=%v",
					desiredReplica, kcps.Items[i].Status.Replicas, kcps.Items[i].Status.ReadyReplicas, kcps.Items[i].Status.UpdatedReplicas))
			}
		}
		return kerrors.NewAggregate(errList)
	default:
		return errors.Errorf("invalid type: %s during VerifyKubeadmControlPlaneReplicas", reflect.TypeOf(kcps))
	}
}

// VerifyMachineDeploymentsReplicas verifies the MachineDeployment has all the required replicas (this is required before starting the move operation)
func VerifyMachineDeploymentsReplicas(obj crtclient.ObjectList) error {
	switch deployments := obj.(type) {
	case *capi.MachineDeploymentList:
		errList := []error{}
		for i := range deployments.Items {
			var desiredReplica int32 = 1
			if deployments.Items[i].Spec.Replicas != nil {
				desiredReplica = *deployments.Items[i].Spec.Replicas
			}
			if desiredReplica != deployments.Items[i].Status.ReadyReplicas {
				errList = append(errList, errors.Errorf("worker nodes are still being created for MachineDeployment '%s', DesiredReplicas=%v Replicas=%v ReadyReplicas=%v UpdatedReplicas=%v",
					deployments.Items[i].Name, desiredReplica, deployments.Items[i].Status.Replicas, deployments.Items[i].Status.ReadyReplicas, deployments.Items[i].Status.UpdatedReplicas))
			}
		}
		return kerrors.NewAggregate(errList)
	default:
		return errors.Errorf("invalid type: %s during VerifyMachineDeploymentsReplicas", reflect.TypeOf(deployments))
	}
}

// VerifyDeploymentAvailable verifies the deployment has at least one replica running under it or not
func VerifyDeploymentAvailable(obj crtclient.Object) error {
	switch deployment := obj.(type) {
	case *appsv1.Deployment:
		if deployment.Status.AvailableReplicas < 1 {
			return errors.Errorf("pods are not yet running for deployment '%s' in namespace '%s'", deployment.Name, deployment.Namespace)
		}
	default:
		return errors.Errorf("invalid type: %s during VerifyDeploymentAvailable", reflect.TypeOf(deployment))
	}
	return nil
}

// VerifyAutoscalerDeploymentAvailable verifies autoscaler deployment's availability
func VerifyAutoscalerDeploymentAvailable(obj crtclient.Object) error {
	switch deployment := obj.(type) {
	case *appsv1.Deployment:
		if *deployment.Spec.Replicas != deployment.Status.AvailableReplicas || *deployment.Spec.Replicas != deployment.Status.UpdatedReplicas || *deployment.Spec.Replicas != deployment.Status.Replicas {
			return errors.Errorf("pods are not yet running for deployment '%s' in namespace '%s'", deployment.Name, deployment.Namespace)
		}
	default:
		return errors.Errorf("invalid type: %s during VerifyAutoscalerDeploymentAvailable", reflect.TypeOf(deployment))
	}
	return nil
}

// VerifyCRSAppliedSuccessfully verifies that all CRS objects are applied successfully after cluster creation
func VerifyCRSAppliedSuccessfully(obj crtclient.ObjectList) error {
	switch crsList := obj.(type) {
	case *addonsv1.ClusterResourceSetList:
		errList := []error{}
		for i := range crsList.Items {
			if !conditions.IsTrue(crsList.Items[i].DeepCopy(), addonsv1.ResourcesAppliedCondition) {
				errList = append(errList, errors.Errorf("ClusterResourceSet %s is not yet applied", crsList.Items[i].Name))
			}
		}
		return kerrors.NewAggregate(errList)
	default:
		return errors.Errorf("invalid type: %s during VerifyCRSAppliedSuccessfully", reflect.TypeOf(crsList))
	}
}

// VerifyAVIResourceCleanupFinished verifies that avi objects clean up finished.
func VerifyAVIResourceCleanupFinished(obj crtclient.Object) error {
	switch statefulSet := obj.(type) {
	case *appsv1.StatefulSet:
		if statefulSet.Annotations != nil && statefulSet.Annotations[constants.AkoCleanUpAnnotationKey] == constants.AkoCleanUpFinishedStatus {
			return nil
		}
		return errors.Errorf("AVI Resource clean up in progress")
	default:
		return errors.Errorf("invalid type: %s during VerifyAVIResourceCleanupFinished", reflect.TypeOf(statefulSet))
	}
}

// VerifyPackageInstallReconciledSuccessfully verifies that packageInstall reconcile successfully
func VerifyPackageInstallReconciledSuccessfully(obj crtclient.Object) error {
	switch packageInstall := obj.(type) {
	case *kappipkg.PackageInstall:

		for _, cond := range packageInstall.Status.Conditions {
			switch cond.Type {
			case kappctrl.ReconcileSucceeded:
				return nil
			case kappctrl.ReconcileFailed:
				return fmt.Errorf("package reconciliation failed. package: %s, reason: %s", packageInstall.Name, packageInstall.Status.UsefulErrorMessage)
			}
		}

		return errors.Errorf("waiting for '%s' Package to be installed", packageInstall.Name)
	default:
		return errors.Errorf("invalid type: %s during VerifyPackageInstallReconcilledSuccessfully", reflect.TypeOf(packageInstall))
	}
}

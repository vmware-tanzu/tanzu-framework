// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterclient

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	betav1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	capav1alpha4 "sigs.k8s.io/cluster-api-provider-aws/api/v1alpha4"
	capzv1alpha4 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha4"
	capvv1alpha4 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha4"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capi "sigs.k8s.io/cluster-api/api/v1alpha4"
	capdv1alpha4 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1alpha4"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha4"
	bootstrapv1alpha3 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha4"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1alpha4"
	"sigs.k8s.io/cluster-api/util/conditions"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/api/tmc/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// PostVerifyrFunc is a function which should be used as closure
type PostVerifyrFunc func(obj crtclient.Object) error

// PostVerifyListrFunc is a function which should be used as closure
type PostVerifyListrFunc func(obj crtclient.ObjectList) error

func (c *client) ListResources(resourceReference interface{}, option ...crtclient.ListOption) error {
	obj, err := c.getRuntimeObjectList(resourceReference)
	if err != nil {
		return err
	}
	if err := c.clientSet.List(context.TODO(), obj, option...); err != nil {
		return errors.Wrapf(err, "failed to list %v", reflect.TypeOf(resourceReference))
	}
	return nil
}

func (c *client) DeleteResource(resourceReference interface{}) error {
	obj, err := c.getRuntimeObject(resourceReference)
	if err != nil {
		return err
	}
	if err := c.clientSet.Delete(ctx, obj); err != nil {
		return errors.Wrapf(err, "failed to delete %v", reflect.TypeOf(resourceReference))
	}
	return nil
}

func (c *client) CreateResource(resourceReference interface{}, resourceName, namespace string, opts ...crtclient.CreateOption) error {
	obj, err := c.getRuntimeObject(resourceReference)
	if err != nil {
		return err
	}

	if err := c.clientSet.Create(ctx, obj, opts...); err != nil {
		return errors.Wrapf(err, "error while creating object for %q %s/%s",
			obj.GetObjectKind(), namespace, resourceName)
	}
	return nil
}

func (c *client) UpdateResource(resourceReference interface{}, resourceName, namespace string, opts ...crtclient.UpdateOption) error {
	obj, err := c.getRuntimeObject(resourceReference)
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

	obj, err := c.getRuntimeObject(resourceReference)
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
	obj, err := c.getRuntimeObject(o)
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
	obj, err := c.getRuntimeObjectList(o)
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

func (c *client) getRuntimeObject(o interface{}) (crtclient.Object, error) { //nolint:gocyclo,funlen
	switch obj := o.(type) {
	case *corev1.Namespace:
		return obj, nil
	case *corev1.Secret:
		return obj, nil
	case *capi.Cluster:
		return obj, nil
	case *capiv1alpha3.Cluster:
		return obj, nil
	case *appsv1.Deployment:
		return obj, nil
	case *appsv1.StatefulSet:
		return obj, nil
	case *capi.Machine:
		return obj, nil
	case *capi.MachineHealthCheck:
		return obj, nil
	case *capi.MachineDeployment:
		return obj, nil
	case *capdv1alpha4.DockerMachineTemplate:
		return obj, nil
	case *capiv1alpha3.Machine:
		return obj, nil
	case *capiv1alpha3.MachineHealthCheck:
		return obj, nil
	case *capiv1alpha3.MachineDeployment:
		return obj, nil
	case *controlplanev1.KubeadmControlPlane:
		return obj, nil
	case *unstructured.Unstructured:
		return obj, nil
	case *betav1.CronJob:
		return obj, nil
	case *capvv1alpha4.VSphereCluster:
		return obj, nil
	case *capvv1alpha4.VSphereMachineTemplate:
		return obj, nil
	case *capav1alpha4.AWSMachineTemplate:
		return obj, nil
	case *capav1alpha4.AWSCluster:
		return obj, nil
	case *appsv1.DaemonSet:
		return obj, nil
	case *corev1.ConfigMap:
		return obj, nil
	case *v1alpha1.Extension:
		return obj, nil
	case *extensionsV1.CustomResourceDefinition:
		return obj, nil
	case *capzv1alpha4.AzureMachineTemplate:
		return obj, nil
	case *capzv1alpha4.AzureCluster:
		return obj, nil
	case *corev1.ServiceAccount:
		return obj, nil
	case *rbacv1.ClusterRole:
		return obj, nil
	case *rbacv1.ClusterRoleBinding:
		return obj, nil
	case *addonsv1.ClusterResourceSet:
		return obj, nil
	case *runv1alpha1.TanzuKubernetesRelease:
		return obj, nil
	case *bootstrapv1alpha3.KubeadmConfigTemplate:
		return obj, nil
	case *kappipkg.PackageInstall:
		return obj, nil
	default:
		return nil, errors.New("invalid object type")
	}
}

func (c *client) getRuntimeObjectList(o interface{}) (crtclient.ObjectList, error) { //nolint:gocyclo,funlen
	switch obj := o.(type) {
	case *corev1.SecretList:
		return obj, nil
	case *clusterctlv1.ProviderList:
		return obj, nil
	case *capi.ClusterList:
		return obj, nil
	case *capi.MachineHealthCheckList:
		return obj, nil
	case *capi.MachineList:
		return obj, nil
	case *capi.MachineDeploymentList:
		return obj, nil
	case *capiv1alpha3.ClusterList:
		return obj, nil
	case *capiv1alpha3.MachineHealthCheckList:
		return obj, nil
	case *capiv1alpha3.MachineList:
		return obj, nil
	case *capiv1alpha3.MachineDeploymentList:
		return obj, nil
	case *controlplanev1.KubeadmControlPlaneList:
		return obj, nil
	case *betav1.CronJobList:
		return obj, nil
	case *capvv1alpha4.VSphereClusterList:
		return obj, nil
	case *v1alpha1.ExtensionList:
		return obj, nil
	case *addonsv1.ClusterResourceSetList:
		return obj, nil
	case *runv1alpha1.TanzuKubernetesReleaseList:
		return obj, nil
	default:
		return nil, errors.New("invalid object type")
	}
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
		for _, condition := range statefulSet.Status.Conditions {
			if condition.Type == constants.AkoCleanupCondition && condition.Status == corev1.ConditionFalse {
				return nil
			}
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
				return fmt.Errorf("package reconciliation failed: %s", packageInstall.Status.UsefulErrorMessage)
			}
		}

		return errors.Errorf("waiting for '%s' Package to be installed", packageInstall.Name)
	default:
		return errors.Errorf("invalid type: %s during VerifyPackageInstallReconcilledSuccessfully", reflect.TypeOf(packageInstall))
	}
}

// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package source

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/image"
	"github.com/pkg/errors"
	runv1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/constants"
	mgrcontext "github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/context"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/registry"
	types "github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/types"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TanzuKubernetesReleaseReconciler reconciles a TanzuKubernetesRelease object
type reconciler struct {
	ctx                        context.Context
	client                     client.Client
	log                        logr.Logger
	scheme                     *runtime.Scheme
	options                    mgrcontext.TanzuKubernetesReleaseDiscoverOptions
	bomImage                   string
	compatibilityMetadataImage string
	registry                   registry.Registry
	registryOps                ctlimg.RegistryOpts
}

// Reconcile performs the reconciliation step
func (r *reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.log.WithValues("tanzukubernetesrelease", req.NamespacedName)
	return ctrl.Result{}, nil
}

// AddToManager adds this package's controller to the provided manager.
func AddToManager(ctx *mgrcontext.ControllerManagerContext, mgr ctrl.Manager) error {
	r := newReconciler(ctx)
	err := mgr.Add(r.(*reconciler))
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&runv1.TanzuKubernetesRelease{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=tanzukubernetesreleases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=tanzukubernetesreleases/status,verbs=get;update;patch

func (r *reconciler) diffRelease(ctx context.Context) (newReleases, existingReleases []runv1.TanzuKubernetesRelease, err error) {
	// get TKRs that already exist
	tkrList := &runv1.TanzuKubernetesReleaseList{}
	err = r.client.List(ctx, tkrList)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list current TKRs")
	}
	existingReleases = tkrList.Items

	tkrMap := make(map[string]bool)
	for _, tkr := range existingReleases {
		tkrMap[tkr.ObjectMeta.Name] = true
	}

	// get all Configmap under the tkr-system namespace
	cmList := &corev1.ConfigMapList{}
	if err := r.client.List(ctx, cmList, &client.ListOptions{Namespace: constants.TKRNamespace}); err != nil {
		return nil, nil, errors.Wrap(err, "failed to get BOM ConfigMaps")
	}
	for _, cm := range cmList.Items {
		// process ConfigMap with the tkr label
		if tkrName, ok := cm.ObjectMeta.Labels[constants.BomConfigMapTKRLabel]; ok {
			if _, ok := tkrMap[tkrName]; ok {
				continue
			}

			bomContent, ok := cm.BinaryData[constants.BomConfigMapContentKey]
			if !ok {
				return nil, nil, errors.New("failed to get the BOM file content from the BOM ConfigMap")
			}

			// generate a TKR if the BOM ConfigMap does not have a corresponding one
			newTkr, err := NewTkrFromBom(tkrName, bomContent)
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to generate TKR from Bom configmap")
			}
			newReleases = append(newReleases, newTkr)
		}
	}

	return newReleases, existingReleases, nil
}

func (r *reconciler) createBOMConfigMap(ctx context.Context, tag string) error {
	bomContent, err := r.registry.GetFile(r.bomImage, tag, "")
	if err != nil {
		r.log.Error(err, "failed to get the BOM file from image", "name", fmt.Sprintf("%s:%s", r.bomImage, tag))
		return nil
	}

	bom, err := types.NewBom(bomContent)
	if err != nil {
		r.log.Error(err, "failed to parse content from image", "name", fmt.Sprintf("%s:%s", r.bomImage, tag))
		return nil
	}

	releaseName, err := bom.GetReleaseVersion()
	if err != nil || releaseName == "" {
		r.log.Error(err, "failed to get the release version from the BOM", "name", fmt.Sprintf("%s:%s", r.bomImage, tag))
		return nil
	}

	strs := strings.Split(releaseName, "+")
	name := strings.Join(strs, "---")

	// label the ConfigMap with image tag and tkr name
	labels := make(map[string]string)
	labels[constants.BomConfigMapTKRLabel] = name

	annotations := make(map[string]string)
	annotations[constants.BomConfigMapImageTagAnnotation] = tag

	binaryData := make(map[string][]byte)
	binaryData[constants.BomConfigMapContentKey] = bomContent

	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   constants.TKRNamespace,
			Labels:      labels,
			Annotations: annotations,
		},
		BinaryData: binaryData,
	}

	return r.client.Create(ctx, &cm)

}

func (r *reconciler) reconcileBOMConfigMap(ctx context.Context) (err error) {
	imageTags, err := r.registry.ListImageTags(r.bomImage)
	if err != nil {
		return errors.Wrap(err, "failed to list current available BOM image tags")
	}
	tagMap := make(map[string]bool)
	for _, tag := range imageTags {
		tagMap[tag] = false
	}

	cmList := &corev1.ConfigMapList{}
	if err := r.client.List(ctx, cmList, &client.ListOptions{Namespace: constants.TKRNamespace}); err != nil {
		return errors.Wrap(err, "failed to get BOM ConfigMaps")
	}

	for _, cm := range cmList.Items {
		if imageTag, ok := cm.ObjectMeta.Annotations[constants.BomConfigMapImageTagAnnotation]; ok {
			if _, ok := tagMap[imageTag]; ok {
				tagMap[imageTag] = true
			}
		}
	}
	for tag, exist := range tagMap {
		if !exist {
			err = r.createBOMConfigMap(ctx, tag)
			if err != nil {
				return errors.Wrapf(err, "failed to create BOM ConfigMap for image %s", fmt.Sprintf("%s:%s", r.bomImage, tag))
			}
		}
	}

	return nil
}

func (r *reconciler) createTKR(ctx context.Context, tkrs []runv1.TanzuKubernetesRelease) (created []runv1.TanzuKubernetesRelease, err error) {
	for _, tkr := range tkrs {

		r.log.Info("Creating release", "name", tkr.Name)
		err := r.client.Create(ctx, &tkr)
		if err != nil {
			return created, errors.Wrapf(err, "failed to create tkr %s", tkr.Name)
		}

		err = r.client.Status().Update(ctx, &tkr)
		if err != nil {
			return created, errors.Wrapf(err, "failed to update status sub resource for TKR %s", tkr.Name)
		}
		created = append(created, tkr)
	}

	return created, nil
}

func changeTKRCondition(tkr *runv1.TanzuKubernetesRelease, conditionType string, status corev1.ConditionStatus, message string) {

	newCondition := &clusterv1.Condition{
		Type:    clusterv1.ConditionType(conditionType),
		Status:  status,
		Message: message,
	}
	conditions.Set(tkr, newCondition)
}

func (r *reconciler) UpdateTKRUpgradeAvailableCondition(tkrs []runv1.TanzuKubernetesRelease) {

	for i, from := range tkrs {
		upgradeTo := []string{}
		for _, to := range tkrs {
			if upgradeQualified(&from, &to) {
				upgradeTo = append(upgradeTo, to.ObjectMeta.Name)
			}
		}
		if len(upgradeTo) != 0 {
			msg := fmt.Sprintf("TKR(s) with later version is available: %s", strings.Join(upgradeTo, ","))
			changeTKRCondition(&tkrs[i], runv1.ConditionUpgradeAvailable, corev1.ConditionTrue, msg)
		} else {
			changeTKRCondition(&tkrs[i], runv1.ConditionUpgradeAvailable, corev1.ConditionFalse, "")
		}
	}

}

func (r *reconciler) UpdateTKRCompatibleCondition(ctx context.Context, tkrs []runv1.TanzuKubernetesRelease) error {
	// TODO: reconcile compatible status based on compatibility metadata

	mgmtClusterVersion, err := r.GetManagementClusterVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get the management cluster info")
	}
	tags, err := r.registry.ListImageTags(r.compatibilityMetadataImage)
	if err != nil || len(tags) == 0 {
		return errors.Wrap(err, "failed to list compatibility metadata image tags")
	}

	tagNum := []int{}
	for _, tag := range tags {
		ver, err := strconv.Atoi(tag[1:])
		if err == nil {
			tagNum = append(tagNum, ver)
		}
	}

	sort.Ints(tagNum)

	metadataContent := []byte{}
	var metadata types.CompatibilityMetadata

	for i := len(tagNum) - 1; i >= 0; i-- {
		tagName := fmt.Sprintf("v%d", tagNum[i])
		metadataContent, err = r.registry.GetFile(r.compatibilityMetadataImage, tagName, "")
		if err == nil {
			if err = yaml.Unmarshal(metadataContent, &metadata); err == nil {
				break
			}
			r.log.Error(err, "failed to unmarshal tkr compatibility metadata file", "image", fmt.Sprintf("%s:%s", r.compatibilityMetadataImage, tagName))
		} else {
			r.log.Error(err, "failed to retrieve tkr compatibility metadata image content", "image", fmt.Sprintf("%s:%s", r.compatibilityMetadataImage, tagName))
		}
	}

	if len(metadataContent) == 0 {
		return errors.New("failed to get tkr compatibility metadata")
	}

	compatibileReleases := []string{}
	for _, mgmtVersion := range metadata.ManagementClusterVersions {
		if strings.HasPrefix(mgmtClusterVersion, mgmtVersion.TKGVersion) {
			compatibileReleases = mgmtVersion.SupportedKubernetesVersions
		}
	}

	compatibleSet := make(map[string]bool)
	for _, r := range compatibileReleases {

		compatibleSet[r] = true
	}

	for i, tkr := range tkrs {
		if _, ok := compatibleSet[tkr.Spec.Version]; ok {
			changeTKRCondition(&tkrs[i], runv1.ConditionCompatible, corev1.ConditionTrue, "")
		} else {
			changeTKRCondition(&tkrs[i], runv1.ConditionCompatible, corev1.ConditionFalse, "")
		}
	}

	return nil
}

func (r *reconciler) ReconcileConditions(ctx context.Context, added, existing []runv1.TanzuKubernetesRelease) error {
	allTKRs := append(added, existing...)

	err := r.UpdateTKRCompatibleCondition(ctx, allTKRs)
	if err != nil {
		return errors.Wrap(err, "failed to update Compatible condition for TKRs")
	}

	r.UpdateTKRUpgradeAvailableCondition(allTKRs)

	for _, tkr := range allTKRs {
		if err = r.client.Status().Update(ctx, &tkr); err != nil {
			return errors.Wrapf(err, "failed to update status sub resrouce for TKR %s", tkr.ObjectMeta.Name)
		}
	}

	return nil
}

func (r *reconciler) SyncRelease(ctx context.Context) (added []runv1.TanzuKubernetesRelease, existing []runv1.TanzuKubernetesRelease, err error) {
	// create BOM ConfigMaps for new images
	if err := r.reconcileBOMConfigMap(ctx); err != nil {
		return added, existing, errors.Wrap(err, "failed to reconcile the BOM ConfigMap")
	}

	newTkrs, existingTkrs, err := r.diffRelease(ctx)
	if err != nil {
		return added, existing, errors.Wrap(err, "failed to sync up TKRs with BOM ConfigMap")
	}

	createdTkrs, err := r.createTKR(ctx, newTkrs)
	if err != nil {
		return added, existing, errors.Wrap(err, "failed to create TKRs")
	}

	return createdTkrs, existingTkrs, nil
}

func (r *reconciler) ReconcileRelease(ctx context.Context) (err error) {

	added, existing, err := r.SyncRelease(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to sync up TKRs with the BOM repository")
	}

	err = r.ReconcileConditions(ctx, added, existing)
	if err != nil {
		return errors.Wrap(err, "failed to reconcile TKR conditions")
	}

	return nil
}

func (r *reconciler) checkInitialSync(ctx context.Context) error {
	r.log.Info("performing checks on initial TKR discovery")
	tkrList := &runv1.TanzuKubernetesReleaseList{}
	err := r.client.List(ctx, tkrList)
	if err != nil {
		return errors.Wrap(err, "failed to list current TKRs")
	}

	tags, err := r.registry.ListImageTags(r.bomImage)
	if err != nil {
		return errors.Wrap(err, "failed to list BOM image tags")
	}
	if len(tags) != len(tkrList.Items) {
		return errors.New("number of inital TKRs and BOM image tags do not match")
	}
	return nil
}

func (r *reconciler) initialReconcile(ticker *time.Ticker, initSyncDone chan bool, stopChan <-chan struct{}, initialDiscoveryRetry int) {
	for {
		select {
		case <-stopChan:
			r.log.Info("Stop performing initial TKR discovery")
			ticker.Stop()
			close(initSyncDone)
			return
		case <-ticker.C:

			var errReconcile error
			var errCheck error

			if errReconcile = r.ReconcileRelease(context.Background()); errReconcile == nil {
				// The number of TKRs should be the same as the number of bom image tags for initial sync-up
				errCheck = r.checkInitialSync(context.Background())
			}

			if errReconcile != nil {
				r.log.Info("Failed to complete initial TKR discovery", "error", errReconcile.Error())
				if isManagementClusterNotReadyError(errReconcile) {
					continue
				}
			} else if errCheck != nil {
				r.log.Info("Failed to complete initial TKR discovery", "error", errCheck.Error())
			}

			if (errReconcile == nil && errCheck == nil) || initialDiscoveryRetry == 0 {
				ticker.Stop()
				close(initSyncDone)
				return
			}
			r.log.Info("Failed to complete initial TKR discovery, retrying")
			initialDiscoveryRetry--
		}
	}
}

func (r *reconciler) tkrDiscovery(ticker *time.Ticker, done chan bool, stopChan <-chan struct{}) {
	for {
		select {
		case <-stopChan:
			r.log.Info("Stop performing TKR discovery")
			ticker.Stop()
			close(done)
			return
		case <-ticker.C:
			err := r.ReconcileRelease(context.Background())
			if err != nil {
				r.log.Info("failed to reconcile TKRs, retrying", "error", err.Error())
			}
		}
	}
}

func (r *reconciler) Start(stopChan <-chan struct{}) error {
	r.log.Info("Starting TanzuKubernetesReleaase Reconciler")

	r.log.Info("Performing configuration setup")
	err := r.Configure()
	if err != nil {
		return errors.Wrap(err, "failed to configure the controller")
	}

	registryCertPath, err := getRegistryCertFile()
	if err == nil {
		if _, err = os.Stat(registryCertPath); err == nil {
			r.registryOps.CACertPaths = []string{registryCertPath}
		}
	}

	r.registry = registry.New(r.registryOps)

	r.log.Info("Performing an initial release discovery")
	initSyncDone := make(chan bool)
	done := make(chan bool)

	ticker := time.NewTicker(r.options.InitialDiscoveryFrequency)
	initialDiscoveryRetry := InitialDiscoveryRetry
	go r.initialReconcile(ticker, initSyncDone, stopChan, initialDiscoveryRetry)

	<-initSyncDone
	r.log.Info("Initial TKR discovery completed")

	ticker = time.NewTicker(r.options.ContinuousDiscoveryFrequency)
	go r.tkrDiscovery(ticker, done, stopChan)

	<-done
	r.log.Info("Stopping TanzuKubernetesReleaase Reconciler")
	return nil
}

func newReconciler(ctx *mgrcontext.ControllerManagerContext) reconcile.Reconciler {

	regOpts := ctlimg.RegistryOpts{
		VerifyCerts: ctx.VerifyRegistryCert,
		Anon:        true,
	}
	return &reconciler{
		ctx:                        ctx.Context,
		client:                     ctx.Client,
		log:                        ctx.Logger,
		scheme:                     ctx.Scheme,
		options:                    ctx.TKRDiscoveryOption,
		registryOps:                regOpts,
		bomImage:                   ctx.BOMImagePath,
		compatibilityMetadataImage: ctx.BOMMetadataImagePath,
	}
}

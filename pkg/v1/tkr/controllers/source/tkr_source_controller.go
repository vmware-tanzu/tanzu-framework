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
	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/registry"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types2 "k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
	mgrcontext "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/context"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/registry"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/util/patchset"
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
	registryOps                ctlimg.Opts
}

// Reconcile performs the reconciliation step
func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result reconcile.Result, retErr error) {
	ps := patchset.New(r.client)
	defer func() {
		// apply patches unless an error is being returned
		if retErr != nil {
			return
		}
		if err := ps.Apply(ctx); err != nil {
			if err = kerrors.FilterOut(err, apierrors.IsConflict); err == nil {
				// retry if someone updated an object we wanted to patch
				result = ctrl.Result{Requeue: true}
			}
			retErr = errors.Wrap(err, "applying patches to TKRs")
		}
	}()

	configMap := &corev1.ConfigMap{}
	if err := r.client.Get(ctx, types2.NamespacedName{Namespace: req.Namespace, Name: req.Name}, configMap); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil // do nothing if the ConfigMap does not exist
		}
		return ctrl.Result{}, err
	}
	r.log.Info("reconciling CM", "ns", req.Namespace, "name", req.Name)
	if configMap.Name == constants.BOMMetadataConfigMapName {
		return ctrl.Result{}, r.updateConditions(ctx, ps)
	}

	return ctrl.Result{}, r.reconcileConfigMap(ctx, configMap)
}

func (r *reconciler) reconcileConfigMap(ctx context.Context, configMap *corev1.ConfigMap) error {
	tkr, err := tkrFromConfigMap(configMap)
	if err != nil {
		r.log.Error(err, "Could not create TKR from ConfigMap", "ConfigMap", configMap.Name)
		return nil // no need to retry: if the ConfigMap changes, we'll get called
	}
	if tkr == nil {
		return nil // no need to retry: no TKR in this ConfigMap
	}

	existingTKR := &runv1.TanzuKubernetesRelease{}
	if err := r.client.Get(ctx, client.ObjectKey{Name: tkr.Name}, existingTKR); err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.client.Create(ctx, tkr); err != nil {
				if apierrors.IsAlreadyExists(err) {
					return nil // the TKR already exists, we're done.
				}
				return errors.Wrapf(err, "could not create TKR: ConfigMap.name='%s'", configMap.Name)
			}
			return nil
		}
		return errors.Wrapf(err, "could not get TKR: ConfigMap.name='%s'", configMap.Name)
	}

	return nil
}

func tkrFromConfigMap(configMap *corev1.ConfigMap) (*runv1.TanzuKubernetesRelease, error) {
	tkrName, labelOK := configMap.ObjectMeta.Labels[constants.BomConfigMapTKRLabel]
	if !labelOK {
		return nil, nil // not interested in ConfigMaps without this label
	}

	bomContent, ok := configMap.BinaryData[constants.BomConfigMapContentKey]
	if !ok {
		return nil, errors.New("failed to get the BOM file content from the BOM ConfigMap")
	}

	newTkr, err := NewTkrFromBom(tkrName, bomContent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate TKR from BOM ConfigMap")
	}

	return &newTkr, nil
}

func (r *reconciler) updateConditions(ctx context.Context, ps patchset.PatchSet) error {
	tkrList := &runv1.TanzuKubernetesReleaseList{}
	if err := r.client.List(ctx, tkrList); err != nil {
		return errors.Wrap(err, "could not list TKRs")
	}

	for i := range tkrList.Items {
		ps.Add(&tkrList.Items[i])
	}

	if err := r.UpdateTKRCompatibleCondition(ctx, tkrList.Items); err != nil {
		return errors.Wrap(err, "failed to update Compatible condition for TKRs")
	}

	r.UpdateTKRUpdatesAvailableCondition(tkrList.Items)

	return nil
}

// AddToManager adds this package's controller to the provided manager.
func AddToManager(ctx *mgrcontext.ControllerManagerContext, mgr ctrl.Manager) error {
	r := newReconciler(ctx)
	if err := mgr.Add(r); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}, builder.WithPredicates(eventFilter(func(eventMeta metav1.Object) bool {
			return eventMeta.GetNamespace() == constants.TKRNamespace && eventMeta.GetName() != constants.TKRControllerLeaderElectionCM
		}))). // we're watching ConfigMaps and producing TKRs
		Watches(&source.Kind{Type: &runv1.TanzuKubernetesRelease{}}, handler.EnqueueRequestsFromMapFunc(watchTKRs)).
		Watches(&source.Kind{Type: &clusterv1.Cluster{}}, handler.EnqueueRequestsFromMapFunc(watchMgmtCluster)).
		Named("tkr-source-controller").
		Complete(r)
}

// watchTKRs returns reconcile.Request for bom-metadata ConfigMap reacting to TKR updates,
// which should affect TKR Compatible and UpdatesAvailable conditions.
func watchTKRs(client.Object) []reconcile.Request {
	return []reconcile.Request{{NamespacedName: client.ObjectKey{
		Namespace: constants.TKRNamespace,
		Name:      constants.BOMMetadataConfigMapName,
	}}}
}

// watchMgmtCluster returns reconcile.Request for bom-metadata ConfigMap reacting to mgmt Cluster updates,
// which should affect TKR Compatible and UpdatesAvailable conditions.
func watchMgmtCluster(o client.Object) []reconcile.Request {
	labels := o.GetLabels()
	if labels == nil {
		return nil
	}
	if _, exists := labels[constants.ManagememtClusterRoleLabel]; exists {
		return []reconcile.Request{{NamespacedName: client.ObjectKey{
			Namespace: constants.TKRNamespace,
			Name:      constants.BOMMetadataConfigMapName,
		}}}
	}
	return nil
}

func eventFilter(p func(eventMeta metav1.Object) bool) *predicate.Funcs {
	return &predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return p(createEvent.Object)
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return p(deleteEvent.Object)
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return p(updateEvent.ObjectOld)
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return p(genericEvent.Object)
		},
	}
}

// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=tanzukubernetesreleases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=tanzukubernetesreleases/status,verbs=get;update;patch

func (r *reconciler) createBOMConfigMap(ctx context.Context, tag string) error {
	r.log.Info("Fetching BOM", "image", r.bomImage, "tag", tag)
	bomContent, err := r.registry.GetFile(fmt.Sprintf("%s:%s", r.bomImage, tag), "")
	if err != nil {
		return errors.Wrapf(err, "failed to get the BOM file from image %s:%s", r.bomImage, tag)
	}

	bom, err := types.NewBom(bomContent)
	if err != nil {
		return errors.Wrapf(err, "failed to parse content from image %s:%s", r.bomImage, tag)
	}

	releaseName, err := bom.GetReleaseVersion()
	if err != nil || releaseName == "" {
		return errors.Wrapf(err, "failed to get the release version from BOM image %s:%s", r.bomImage, tag)
	}

	name := strings.ReplaceAll(releaseName, "+", "---")

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

	if err := r.client.Create(ctx, &cm); err != nil && !apierrors.IsAlreadyExists(err) {
		return errors.Wrapf(err, "could not create ConfigMap: name='%s'", cm.Name)
	}
	return nil
}

func (r *reconciler) reconcileBOMMetadataCM(ctx context.Context) error {
	metadata, err := r.fetchCompatibilityMetadata()
	if err != nil {
		return err
	}

	metadataContent, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.TKRNamespace,
			Name:      constants.BOMMetadataConfigMapName,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.client, cm, func() error {
		cm.BinaryData = map[string][]byte{
			constants.BOMMetadataCompatibilityKey: metadataContent,
		}
		return nil
	})

	return errors.Wrap(err, "error creating or updating BOM metadata ConfigMap")
}

func (r *reconciler) reconcileBOMConfigMap(ctx context.Context) error {
	r.log.Info("Listing BOM image tags", "image", r.bomImage)
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

	for i := range cmList.Items {
		if imageTag, ok := cmList.Items[i].ObjectMeta.Annotations[constants.BomConfigMapImageTagAnnotation]; ok {
			if _, ok := tagMap[imageTag]; ok {
				tagMap[imageTag] = true
			}
		}
	}
	var errs errorSlice
	for tag, exist := range tagMap {
		if !exist {
			if err := r.createBOMConfigMap(ctx, tag); err != nil {
				errs = append(errs, errors.Wrapf(err, "failed to create BOM ConfigMap for image %s", fmt.Sprintf("%s:%s", r.bomImage, tag)))
			}
		}
	}
	if len(errs) != 0 {
		return errs
	}

	r.log.Info("Done reconciling BOM images", "image", r.bomImage)
	return nil
}

func changeTKRCondition(tkr *runv1.TanzuKubernetesRelease, conditionType string, status corev1.ConditionStatus, message string) {
	newCondition := &clusterv1.Condition{
		Type:    clusterv1.ConditionType(conditionType),
		Status:  status,
		Message: message,
	}
	conditions.Set(tkr, newCondition)
}

func (r *reconciler) UpdateTKRUpdatesAvailableCondition(tkrs []runv1.TanzuKubernetesRelease) {
	for i := range tkrs {
		upgradeTo := []string{}
		for j := range tkrs {
			if upgradeQualified(&tkrs[i], &tkrs[j]) {
				upgradeTo = append(upgradeTo, tkrs[j].Spec.Version)
			}
		}
		if len(upgradeTo) != 0 {
			msg := fmt.Sprintf("[%s]", strings.Join(upgradeTo, " "))
			changeTKRCondition(&tkrs[i], runv1.ConditionUpdatesAvailable, corev1.ConditionTrue, msg)
		} else {
			changeTKRCondition(&tkrs[i], runv1.ConditionUpdatesAvailable, corev1.ConditionFalse, "")
		}

		if hasDeprecateUpgradeAvailableCondition(tkrs[i].Status.Conditions) {
			if len(upgradeTo) != 0 {
				msg := fmt.Sprintf("Deprecated, TKR(s) with later version is available: %s", strings.Join(upgradeTo, ","))
				changeTKRCondition(&tkrs[i], runv1.ConditionUpgradeAvailable, corev1.ConditionTrue, msg)
			} else {
				changeTKRCondition(&tkrs[i], runv1.ConditionUpgradeAvailable, corev1.ConditionFalse, "Deprecated")
			}
		}
	}
}

func (r *reconciler) UpdateTKRCompatibleCondition(ctx context.Context, tkrs []runv1.TanzuKubernetesRelease) error {
	// TODO: reconcile compatible status based on compatibility metadata

	compatibleSet := make(map[string]struct{})
	defer func() { // update conditions no matter what
		for i := range tkrs {
			if _, ok := compatibleSet[tkrs[i].Spec.Version]; ok {
				changeTKRCondition(&tkrs[i], runv1.ConditionCompatible, corev1.ConditionTrue, "")
			} else {
				changeTKRCondition(&tkrs[i], runv1.ConditionCompatible, corev1.ConditionFalse, "")
			}
		}
	}()

	mgmtClusterVersion, err := r.GetManagementClusterVersion(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get the management cluster info")
	}

	metadata, err := r.compatibilityMetadata(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to get BOM compatibility metadata")
	}

	compatibleReleases := []string{}
	for _, mgmtVersion := range metadata.ManagementClusterVersions {
		// Fix before TKG v1.10: what if mgmtClusterVersion is "v1.10" and mgmtVersion.TKGVersion is "v1.1"?
		// See https://github.com/vmware-tanzu/tanzu-framework/issues/452
		if strings.HasPrefix(mgmtClusterVersion, mgmtVersion.TKGVersion) {
			compatibleReleases = mgmtVersion.SupportedKubernetesVersions
		}
	}

	for _, r := range compatibleReleases {
		compatibleSet[r] = struct{}{}
	}

	return nil
}

func (r *reconciler) fetchCompatibilityMetadata() (*types.CompatibilityMetadata, error) {
	r.log.Info("Listing BOM metadata image tags", "image", r.compatibilityMetadataImage)
	tags, err := r.registry.ListImageTags(r.compatibilityMetadataImage)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list compatibility metadata image tags")
	}
	if len(tags) == 0 {
		return nil, errors.New("no compatibility metadata image tags found")
	}

	tagNum := []int{}
	for _, tag := range tags {
		ver, err := strconv.Atoi(strings.TrimPrefix(tag, "v"))
		if err == nil {
			tagNum = append(tagNum, ver)
		}
	}

	sort.Ints(tagNum)

	metadataContent := []byte{}
	var metadata types.CompatibilityMetadata

	for i := len(tagNum) - 1; i >= 0; i-- {
		tagName := fmt.Sprintf("v%d", tagNum[i])
		r.log.Info("Fetching BOM metadata image", "image", r.compatibilityMetadataImage, "tag", tagName)
		metadataContent, err = r.registry.GetFile(fmt.Sprintf("%s:%s", r.compatibilityMetadataImage, tagName), "")
		if err == nil {
			if err = yaml.Unmarshal(metadataContent, &metadata); err == nil {
				break
			}
			r.log.Error(err, "Failed to unmarshal TKR compatibility metadata file", "image", fmt.Sprintf("%s:%s", r.compatibilityMetadataImage, tagName))
		} else {
			r.log.Error(err, "Failed to retrieve TKR compatibility metadata image content", "image", fmt.Sprintf("%s:%s", r.compatibilityMetadataImage, tagName))
		}
	}

	if len(metadataContent) == 0 {
		return nil, errors.New("failed to fetch TKR compatibility metadata")
	}

	return &metadata, nil
}

func (r *reconciler) compatibilityMetadata(ctx context.Context) (*types.CompatibilityMetadata, error) {
	cm := &corev1.ConfigMap{}
	cmObjectKey := client.ObjectKey{Namespace: constants.TKRNamespace, Name: constants.BOMMetadataConfigMapName}
	if err := r.client.Get(ctx, cmObjectKey, cm); err != nil {
		return nil, err
	}

	metadataContent, ok := cm.BinaryData[constants.BOMMetadataCompatibilityKey]
	if !ok {
		return nil, errors.New("compatibility key not found in bom-metadata ConfigMap")
	}

	var metadata types.CompatibilityMetadata
	if err := yaml.Unmarshal(metadataContent, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (r *reconciler) SyncRelease(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errors.New("canceled")
	default:
	}

	// create/update bom-metadata ConfigMap
	if err := r.reconcileBOMMetadataCM(ctx); err != nil {
		// not returning: even if we fail to get BOM metadata, we still want to reconcile BOM ConfigMaps
		r.log.Error(err, "Failed to reconcile BOM metadata ConfigMap")
	}
	// create BOM ConfigMaps for new images
	err := r.reconcileBOMConfigMap(ctx)
	return errors.Wrap(err, "failed to reconcile BOM ConfigMaps")
}

func (r *reconciler) initialReconcile(ctx context.Context, frequency time.Duration, retries int) {
	for {
		if err := r.SyncRelease(ctx); err != nil {
			r.log.Error(err, "Failed to complete initial TKR discovery")
			retries--
			if retries <= 0 {
				return
			}

			r.log.Info("Failed to complete initial TKR discovery, retrying")
			select {
			case <-ctx.Done():
				r.log.Info("Stop performing initial TKR discovery")
				return
			case <-time.After(frequency):
				continue
			}
		}
		return
	}
}

func (r *reconciler) tkrDiscovery(ctx context.Context, frequency time.Duration) {
	for {
		if err := r.SyncRelease(ctx); err != nil {
			r.log.Error(err, "Failed to reconcile TKRs, retrying")
		}
		select {
		case <-ctx.Done():
			r.log.Info("Stop performing TKR discovery")
			return
		case <-time.After(frequency):
		}
	}
}

func (r *reconciler) Start(ctx context.Context) error {
	var err error
	r.log.Info("Starting TanzuKubernetesReleaase Reconciler")

	r.log.Info("Performing configuration setup")
	err = r.Configure()
	if err != nil {
		return errors.Wrap(err, "failed to configure the controller")
	}

	// Add custom CA cert paths only if VerifyCerts is enabled
	if r.registryOps.VerifyCerts {
		registryCertPath, err := getRegistryCertFile()
		if err == nil {
			if _, err = os.Stat(registryCertPath); err == nil {
				r.registryOps.CACertPaths = []string{registryCertPath}
			}
		}
	}

	r.registry, err = registry.New(&r.registryOps)
	if err != nil {
		return err
	}

	r.log.Info("Performing an initial release discovery")
	r.initialReconcile(ctx, r.options.InitialDiscoveryFrequency, InitialDiscoveryRetry)

	r.log.Info("Initial TKR discovery completed")

	r.tkrDiscovery(ctx, r.options.ContinuousDiscoveryFrequency)

	r.log.Info("Stopping Tanzu Kubernetes release Reconciler")
	return nil
}

func newReconciler(ctx *mgrcontext.ControllerManagerContext) *reconciler {
	regOpts := ctlimg.Opts{
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

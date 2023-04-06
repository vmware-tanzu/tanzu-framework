// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package featuregateclient

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/config"
)

// FeatureGateClient defines methods to interact with FeatureGate resources
type FeatureGateClient struct {
	crClient client.Client
}

// NewFeatureGateClient returns an instance of FeatureGateClient.
func NewFeatureGateClient(options ...Option) (*FeatureGateClient, error) {
	featureGateClient := &FeatureGateClient{}
	// Apply options
	for _, option := range options {
		featureGateClient = option(featureGateClient)
	}
	if featureGateClient.crClient == nil {
		c, err := getFeatureGateClient()
		if err != nil {
			return nil, err
		}
		featureGateClient.crClient = c
	}
	return featureGateClient, nil
}

// Option is FeatureGateClient Option definition
type Option func(*FeatureGateClient) *FeatureGateClient

// WithClient function is for setting the passed in client when creating FeatureGateClient
func WithClient(cl client.Client) Option {
	return func(featureGateClient *FeatureGateClient) *FeatureGateClient {
		featureGateClient.crClient = cl
		return featureGateClient
	}
}

// getFeatureGateClient returns a new FeatureGate client
func getFeatureGateClient() (client.Client, error) {
	var err error

	scheme := runtime.NewScheme()
	if err := corev1alpha2.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	restConfig, err := getCurrentClusterConfig()
	if err != nil {
		return nil, err
	}
	crClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("could not create cluster client: %w", err)
	}
	return crClient, nil
}

// GetFeatureGate fetches the specified FeatureGate resource.
func (f *FeatureGateClient) GetFeatureGate(ctx context.Context, featureGateName string) (*corev1alpha2.FeatureGate, error) {
	gate := &corev1alpha2.FeatureGate{}
	err := f.crClient.Get(ctx, client.ObjectKey{Name: featureGateName}, gate)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrTypeNotFound
		}
		return nil, fmt.Errorf("could not get featuregate %s: %w", featureGateName, err)
	}
	return gate, nil
}

// GetFeatureGateList fetches all featuregates on the cluster.
func (f *FeatureGateClient) GetFeatureGateList(ctx context.Context) (*corev1alpha2.FeatureGateList, error) {
	gates := &corev1alpha2.FeatureGateList{}
	err := f.crClient.List(ctx, gates)
	if err != nil {
		return nil, fmt.Errorf("could not get featuregates on cluster: %w", err)
	}
	return gates, nil
}

// GetFeature fetches the specified Feature resource.
func (f *FeatureGateClient) GetFeature(ctx context.Context, featureName string) (*corev1alpha2.Feature, error) {
	feature := &corev1alpha2.Feature{}
	err := f.crClient.Get(ctx, client.ObjectKey{Name: featureName}, feature)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrTypeNotFound
		}
		return nil, err
	}
	return feature, nil
}

// GetFeatureList fetches all features on the cluster.
func (f *FeatureGateClient) GetFeatureList(ctx context.Context) (*corev1alpha2.FeatureList, error) {
	features := &corev1alpha2.FeatureList{}
	err := f.crClient.List(ctx, features)
	if err != nil {
		return nil, fmt.Errorf("could not get features on cluster: %w", err)
	}
	return features, nil
}

// ActivateFeature activates a Feature if it passes validation and warranty checks.
// Warning: Before sending `true` via the warrantyVoidAllowed function argument, ensure
// explicit user awareness and approval if activating a Feature will cause the support
// warranty to be void. Once warranty is void, it is is permanent for the environment.
func (f *FeatureGateClient) ActivateFeature(ctx context.Context, featureName string, warrantyVoidAllowed bool) error {
	// A Feature must exist in the cluster to be activated.
	feature, err := f.GetFeature(ctx, featureName)
	if err != nil {
		return fmt.Errorf("could not get Feature %s: %w", featureName, err)
	}

	gates, err := f.GetFeatureGateList(ctx)
	if err != nil {
		return fmt.Errorf("could not get FeatureGateList: %w", err)
	}

	gateName, featRef := FeatureRefFromGateList(gates, featureName)

	if featRef.Activate {
		fmt.Printf("The Feature Reference %s is already set to be activated in FeatureGate %s.\n", featureName, gateName)
		return nil
	}

	if err := validateFeatureActivationToggle(gates, feature); err != nil {
		return err
	}

	gate, err := f.GetFeatureGate(ctx, gateName)
	if err != nil {
		return err
	}

	ok, err := setVoidWarrantyChecksPass(featRef, feature, warrantyVoidAllowed)
	if err != nil {
		return err
	}
	if ok {
		if err := f.setVoidWarranty(ctx, gate, feature.Name); err != nil {
			return err
		}
	}

	return f.setActivated(ctx, gate, featureName)
}

// FeatureRefFromGateList finds the requested Feature from a list of featuregates. If found,
// the name of the FeatureGate and the FeatureReference is returned.
func FeatureRefFromGateList(gates *corev1alpha2.FeatureGateList, featureName string) (string, corev1alpha2.FeatureReference) {
	for i := range gates.Items {
		for _, featRef := range gates.Items[i].Spec.Features {
			if featureName == featRef.Name {
				return gates.Items[i].Name, featRef
			}
		}
	}
	return "", corev1alpha2.FeatureReference{}
}

// setVoidWarrantyChecksPass will check if voiding the support warranty will happen for a Feature and if so,
// that it is allowed. The following will pass the set void warranty check, indicating that the warranty
// can be voided. All conditions must be met to pass the check.
//   - The stability policy dictates that deviating from the activation default will void warranty.
//   - Warranty will be voided (it has not already been voided).
//   - The new activation setting request is different than the default. Another way to say this is that
//     the old activation setting is the same as the default (policy.DefaultActivation == ref.Activate).
//   - The user gave permission to void the warranty.
//
// Not passing conditions, however, need one or more of the following to not pass:
//   - Warranty is not going to be voided.
//   - Warranty will be voided, but user does not give permission to do so.
//   - The new activation setting is the same as the default. Another way to say this is that the old
//     activation setting is different than the default (policy.DefaultActivation != ref.Activate)
func setVoidWarrantyChecksPass(ref corev1alpha2.FeatureReference, feature *corev1alpha2.Feature, warrantyVoidAllowed bool) (bool, error) {
	stability := feature.Spec.Stability
	policy := corev1alpha2.GetPolicyForStabilityLevel(stability)

	// Check if toggling activation state will void the warranty if not already voided.
	if policy.VoidsWarranty && !ref.PermanentlyVoidAllSupportGuarantees && policy.DefaultActivation == ref.Activate {
		// Ensure it is acceptable to the user to void the support warranty of a Feature.
		if warrantyVoidAllowed {
			return true, nil
		}
		return false, fmt.Errorf("warranty will be voided with new activation set point, but user has not given express permission to void the warranty: %w", ErrTypeForbidden)
	}

	// The requested Feature activation set point is already the same as default and will not void warranty,
	// or the warranty is already permanently voided. Nothing needs to change with the warranty state.
	return false, nil
}

func (f *FeatureGateClient) setVoidWarranty(ctx context.Context, gate *corev1alpha2.FeatureGate, featureName string) error {
	for i, featureRef := range gate.Spec.Features {
		if featureRef.Name == featureName {
			gate.Spec.Features[i].PermanentlyVoidAllSupportGuarantees = true
			return f.crClient.Update(ctx, gate)
		}
	}
	return fmt.Errorf("could not void warranty for Feature %s as it was not found in any FeatureGate", ErrTypeNotFound)
}

// setActivated sets the Feature to activate in FeatureGate
func (f *FeatureGateClient) setActivated(ctx context.Context, gate *corev1alpha2.FeatureGate, featureName string) error {
	for i := range gate.Spec.Features {
		if gate.Spec.Features[i].Name == featureName {
			gate.Spec.Features[i].Activate = true
			return f.crClient.Update(ctx, gate)
		}
	}
	return fmt.Errorf("could not activate Feature %s as it was not found in FeatureGate %s: %w", featureName, gate.Name, ErrTypeNotFound)
}

// DeactivateFeature deactivates a Feature. Along with the error, it returns the name of the FeatureGate
// that gates the Feature.
func (f *FeatureGateClient) DeactivateFeature(ctx context.Context, featureName string) (string, error) {
	// A Feature must exist in the cluster to be deactivated.
	feature, err := f.GetFeature(ctx, featureName)
	if err != nil {
		return "", fmt.Errorf("could not get Feature %s: %w", featureName, err)
	}

	gates, err := f.GetFeatureGateList(ctx)
	if err != nil {
		return "", fmt.Errorf("could not get FeatureGateList: %w", err)
	}

	gateName, featRef := FeatureRefFromGateList(gates, featureName)

	if gateName != "" && !featRef.Activate {
		fmt.Printf("The Feature Reference %s is already set to be deactivated in FeatureGate %s.\n", featureName, gateName)
		return gateName, nil
	}

	if err := validateFeatureActivationToggle(gates, feature); err != nil {
		return gateName, err
	}

	gate, err := f.GetFeatureGate(ctx, gateName)
	if err != nil {
		return gateName, err
	}

	return gateName, f.setDeactivated(ctx, gate, featureName)
}

// setDeactivated sets the Feature to 'deactivate' in the FeatureGate resource.
func (f *FeatureGateClient) setDeactivated(ctx context.Context, gate *corev1alpha2.FeatureGate, featureName string) error {
	for i, featureRef := range gate.Spec.Features {
		if featureRef.Name == featureName {
			gate.Spec.Features[i].Activate = false
			return f.crClient.Update(ctx, gate)
		}
	}
	return nil
}

// getCurrentClusterConfig gets the config of current logged in cluster
func getCurrentClusterConfig() (*rest.Config, error) {
	server, err := config.GetCurrentServer()
	if err != nil {
		return nil, err
	}
	restConfig, err := getRestConfigWithContext(server.ManagementClusterOpts.Context, server.ManagementClusterOpts.Path)
	if err != nil {
		return nil, fmt.Errorf("could not get rest config: %w", err)
	}
	return restConfig, nil
}

// getRestConfigWithContext returns config using the passed context.
func getRestConfigWithContext(ctx, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: ctx,
		}).ClientConfig()
}

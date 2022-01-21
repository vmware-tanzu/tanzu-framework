// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"fmt"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

// FeatureGateClient defines methods to interact with FeatureGate resources
type FeatureGateClient struct {
	c client.Client
}

// NewFeatureGateClient returns an instance of FeatureGateClient.
func NewFeatureGateClient(options ...Option) (*FeatureGateClient, error) {
	featureGateClient := &FeatureGateClient{}
	// Apply options
	for _, option := range options {
		featureGateClient = option(featureGateClient)
	}
	if featureGateClient.c == nil {
		c, err := getFeatureGateClient()
		if err != nil {
			return nil, err
		}
		featureGateClient.c = c
	}
	return featureGateClient, nil
}

// Option is FeatureGateClient Option definition
type Option func(*FeatureGateClient) *FeatureGateClient

// WithClient function is for setting the passed in client when creating FeatureGateClient
func WithClient(cl client.Client) Option {
	return func(featureGateClient *FeatureGateClient) *FeatureGateClient {
		featureGateClient.c = cl
		return featureGateClient
	}
}

// getFeatureGateClient returns a new FeatureGate client
func getFeatureGateClient() (client.Client, error) {
	var restConfig *rest.Config
	var err error

	scheme := runtime.NewScheme()
	if err := configv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if restConfig, err = config.GetCurrentClusterConfig(); err != nil {
		return nil, err
	}

	crClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("unable to create cluster client: %w", err)
	}
	return crClient, nil
}

// GetFeatureGate fetches the FeatureGate
func (f *FeatureGateClient) GetFeatureGate(ctx context.Context, featureGateName string) (*configv1alpha1.FeatureGate, error) {
	gate := &configv1alpha1.FeatureGate{}

	if err := f.c.Get(ctx, client.ObjectKey{
		Name: featureGateName,
	}, gate); err != nil {
		return nil, fmt.Errorf("couldn't get featuregate %s: %w", featureGateName, err)
	}
	return gate, nil
}

// GetFeature fetches the Feature
func (f *FeatureGateClient) GetFeature(ctx context.Context, featureName string) (*configv1alpha1.Feature, error) {
	feature := &configv1alpha1.Feature{}

	if err := f.c.Get(ctx, client.ObjectKey{Name: featureName}, feature); err != nil {
		return nil, err
	}
	return feature, nil
}

// ActivateFeature activates a Feature
func (f *FeatureGateClient) ActivateFeature(ctx context.Context, featureName, featureGateName string) error {
	feature, err := f.GetFeature(ctx, featureName)
	if err != nil {
		return fmt.Errorf("couldn't get feature %s: %w", featureName, err)
	}
	if !feature.Spec.Discoverable {
		return fmt.Errorf("feature not found %s", featureName)
	} else if feature.Spec.Immutable {
		return fmt.Errorf("cannot activate an immutable feature %s", featureName)
	}
	gate, err := f.GetFeatureGate(ctx, featureGateName)
	if err != nil {
		return err
	}
	return f.setActivated(ctx, gate, featureName)
}

// setActivated sets the Feature to activate in FeatureGate
func (f *FeatureGateClient) setActivated(ctx context.Context, gate *configv1alpha1.FeatureGate, featureName string) error {
	for i, featureRef := range gate.Spec.Features {
		if featureRef.Name == featureName {
			gate.Spec.Features[i].Activate = true
			return f.c.Update(ctx, gate)
		}
	}

	ref := configv1alpha1.FeatureReference{
		Name:     featureName,
		Activate: true,
	}
	gate.Spec.Features = append(gate.Spec.Features, ref)
	if err := f.c.Update(ctx, gate); err != nil {
		return fmt.Errorf("couldn't update featurgate %s: %w", gate.Name, err)
	}
	return nil
}

// DeactivateFeature deactivates a Feature
func (f *FeatureGateClient) DeactivateFeature(ctx context.Context, featureName, featureGateName string) error {
	feature, err := f.GetFeature(ctx, featureName)
	if err != nil {
		return fmt.Errorf("couldn't get feature %s: %w", featureName, err)
	}
	if !feature.Spec.Discoverable {
		return fmt.Errorf("feature not found %s", featureName)
	} else if feature.Spec.Immutable {
		return fmt.Errorf("cannot deactivate an immutable feature %s", featureName)
	}
	gate, err := f.GetFeatureGate(ctx, featureGateName)
	if err != nil {
		return err
	}
	return f.setDeactivated(ctx, gate, featureName)
}

// setDeactivated sets the Feature to deactivate in FeatureGate
func (f *FeatureGateClient) setDeactivated(ctx context.Context, gate *configv1alpha1.FeatureGate, featureName string) error {
	for i, featureRef := range gate.Spec.Features {
		if featureRef.Name == featureName {
			if !featureRef.Activate {
				return nil
			}
			gate.Spec.Features[i].Activate = false
			return f.c.Update(ctx, gate)
		}
	}

	ref := configv1alpha1.FeatureReference{
		Name:     featureName,
		Activate: false,
	}
	gate.Spec.Features = append(gate.Spec.Features, ref)
	if err := f.c.Update(ctx, gate); err != nil {
		return fmt.Errorf("couldn't update featurgate %s: %w", gate.Name, err)
	}
	return nil
}

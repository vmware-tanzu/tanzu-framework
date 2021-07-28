package main

import (
	"context"
	"os"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/capabilities/discovery/constants"
)

// Operator -> FeatureGate for interacting with gate - whats activated/deactived/unavialble
// Developer -> Kubectl get for any installed, more details, creation of new features, validation, etc

var logLevel int32
var logFile, outputFormat, kubeConfig, kubeContext string
var activated, deactivated, unavailable, extended bool

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "featuregate",
	Description: "operate on features via the tkg feature-gate",
	Version:     "v0.0.1",
	Group:       cliv1alpha1.ManageCmdGroup,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.Cmd.PersistentFlags().Int32VarP(&logLevel, "verbose", "", 0, "Number for the log level verbosity(0-9)")
	p.Cmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Log file path")

	p.AddCommands(
		FeatureListCmd,
		InstallGateCmd,
		DescribeGateCmd,
		FeatureActivateCmd,
		FeatureDeactivateCmd,
	)
	FeatureListCmd.Flags().BoolVarP(&extended, "extended", "e", false, "Include extended output. Higher latency as it requires more API calls.")
	FeatureListCmd.Flags().BoolVarP(&activated, "activated", "a", false, "List only activated Features")
	FeatureListCmd.Flags().BoolVarP(&unavailable, "unavailable", "v", false, "List only  Features specified in the gate but missing from cluster")
	FeatureListCmd.Flags().BoolVarP(&deactivated, "deactivated", "d", false, "List only deactivated Features")
	FeatureListCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")

	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

var InstallGateCmd = &cobra.Command{
	Use:   "install",
	Short: "Install generic FeatureGate on the cluster",
	Args:  cobra.NoArgs,
	RunE:  featureInstall,
}

var DescribeGateCmd = &cobra.Command{
	Use:   "overrides",
	Short: "describe the system FeatureGate overrides",
	Args:  cobra.NoArgs,
	RunE:  featureGateDescribe,
}

var FeatureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Features",
	Args:  cobra.NoArgs,
	Example: `
    # List a clusters Features
    tanzu featuregate list (default)
    tanzu featuregate list --extended
    tanzu featuregate list --activated
    tanzu featuregate list --available
    tanzu featuregate list --deactivated`,
	RunE: featureList,
}

var FeatureActivateCmd = &cobra.Command{
	Use:   "activate <feature>",
	Short: "Activate Features",
	Args:  cobra.ExactArgs(1),
	Example: `
    # List a clusters Features
    tanzu feature activate myfeature`,
	RunE: featureActivate,
}

var FeatureDeactivateCmd = &cobra.Command{
	Use:   "deactivate <feature>",
	Short: "Deactivate Features",
	Args:  cobra.ExactArgs(1),
	Example: `
    # List a clusters Features
    tanzu feature deactivate myfeature`,
	RunE: featureDeactivate,
}

func featureInstall(cmd *cobra.Command, args []string) error {
	runner, err := NewFeatureRunner()
	if err != nil {
		return err
	}

	if _, err := runner.createSystemGate(); err != nil {
		return err
	}
	cmd.Println("TKG FeatureGate Installed")
	return nil
}

func featureActivate(cmd *cobra.Command, args []string) error {
	runner, err := NewFeatureRunner()
	if err != nil {
		return err
	}

	if err := runner.ActivateFeature(args[0]); err != nil {
		return err
	}
	cmd.Println("Feature Activated")
	return nil
}

func featureDeactivate(cmd *cobra.Command, args []string) error {
	runner, err := NewFeatureRunner()
	if err != nil {
		return err
	}

	if err := runner.DeactivateFeature(args[0]); err != nil {
		return err
	}
	cmd.Println("Feature Deactivated")
	return nil
}

type featureInfo struct {
	name        string
	maturity    string
	description string
	activated   bool
	available   bool
	immutable   bool
}

// TODO: join features to display more details
func featureList(cmd *cobra.Command, _ []string) error {
	runner, err := NewFeatureRunner()
	if err != nil {
		return err
	}

	systemGate, err := runner.GetTKGSystemFeatureGate()
	if crtclient.IgnoreNotFound(err) != nil {
		return err
	}
	if err != nil {
		cmd.Println("tkg-system gate not found, processing Feature overrides")
	}

	features := []*featureInfo{}
	switch {
	case activated:
		for _, a := range systemGate.Status.ActivatedFeatures {
			features = append(features, &featureInfo{
				name:      a,
				activated: true,
				available: true,
			})
		}
	case deactivated:
		for _, a := range systemGate.Status.DeactivatedFeatures {
			features = append(features, &featureInfo{
				name:      a,
				activated: false,
				available: true,
			})
		}
	case unavailable:
		for _, a := range systemGate.Status.UnavailableFeatures {
			features = append(features, &featureInfo{
				name:      a,
				activated: false,
				available: false,
			})
		}
	default:
		for _, a := range systemGate.Status.ActivatedFeatures {
			features = append(features, &featureInfo{
				name:      a,
				activated: true,
				available: true,
			})
		}
		for _, a := range systemGate.Status.DeactivatedFeatures {
			features = append(features, &featureInfo{
				name:      a,
				activated: false,
				available: true,
			})
		}
		for _, a := range systemGate.Status.UnavailableFeatures {
			features = append(features, &featureInfo{
				name:      a,
				activated: false,
				available: false,
			})
		}
	}
	if extended {
		err := runner.joinFeatures(features)
		if err != nil {
			return err
		}

		// Many API calls to gather and join on Features
		return listExtended(cmd, systemGate, features)
	}

	// Single API call to FeatureGate
	return listBasic(cmd, systemGate, features)
}

func listExtended(cmd *cobra.Command, systemGate *configv1alpha1.FeatureGate, features []*featureInfo) error {
	var t component.OutputWriterSpinner
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		"Retrieving Features from system gate...", true, "NAME", "ACTIVATION STATE", "AVAILABLE", "MATURITY", "DESCRIPTION", "IMMUTABLE")
	if err != nil {
		return err
	}

	// Determine if the state is due to a Gate which is overriding it, and which gate.
	for _, info := range features {
		t.AddRow(
			info.name,
			info.activated,
			info.available,
			info.maturity,
			info.description,
			info.immutable)
	}
	t.RenderWithSpinner()

	return nil
}

func listBasic(cmd *cobra.Command, systemGate *configv1alpha1.FeatureGate, features []*featureInfo) error {
	var t component.OutputWriterSpinner
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		"Retrieving Features from system gate...", true, "NAME", "ACTIVATION STATE", "AVAILABLE")
	if err != nil {
		return err
	}

	// Determine if the state is due to a Gate which is overriding it, and which gate.
	for _, info := range features {
		t.AddRow(
			info.name,
			info.activated,
			info.available)
	}
	t.RenderWithSpinner()

	return nil
}

func featureGateDescribe(cmd *cobra.Command, _ []string) error {
	runner, err := NewFeatureRunner()
	if err != nil {
		return err
	}

	systemGate, err := runner.GetTKGSystemFeatureGate()
	if apierrors.IsNotFound(err) {
		cmd.Printf("System FeatureGate not installed - run featuregate install")
		return err
	}
	var t component.OutputWriterSpinner
	t, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		"Retrieving System FeatureGate...", true, "FEATURE NAME", "OVERRIDE")
	if err != nil {
		return err
	}

	for _, f := range systemGate.Spec.Features {
		t.AddRow(
			f.Name,
			f.Activate)
	}
	t.RenderWithSpinner()

	return nil
}

func getCRClient() (crtclient.Client, error) {
	var restConfig *rest.Config
	var err error

	scheme := runtime.NewScheme()
	if err := configv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if kubeConfig == "" {
		if restConfig, err = k8sconfig.GetConfig(); err != nil {
			return nil, err
		}
	} else {
		config, err := clientcmd.LoadFromFile(kubeConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load from %s", kubeConfig)
		}
		rawConfig, err := clientcmd.Write(*config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to write config")
		}
		if restConfig, err = clientcmd.RESTConfigFromKubeConfig(rawConfig); err != nil {
			return nil, errors.Wrap(err, "Unable to set up rest config")
		}
	}
	mapper, err := apiutil.NewDynamicRESTMapper(restConfig, apiutil.WithLazyDiscovery)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to set up rest mapper")
	}
	crtClient, err := crtclient.New(restConfig, crtclient.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create cluster client")
	}
	return crtClient, nil
}

func NewFeatureRunner() (*featureGateRunner, error) {
	crClient, err := getCRClient()
	if err != nil {
		return nil, err
	}

	return &featureGateRunner{crtClient: crClient}, nil
}

type featureGateRunner struct {
	crtClient crtclient.Client
}

func (r *featureGateRunner) joinFeatures(features []*featureInfo) error {
	for _, info := range features {
		if !info.available {
			continue
		}
		f, err := r.GetFeature(info.name)
		if err != nil {
			return err
		}
		info.maturity = f.Spec.Maturity
		info.description = f.Spec.Description
		info.immutable = f.Spec.Immutable
	}
	return nil
}

func (f *featureGateRunner) ActivateFeature(featureName string) error {
	gate, err := f.GetTKGSystemFeatureGate()
	if crtclient.IgnoreNotFound(err) != nil {
		return err
	}
	if err != nil {
		gate, err = f.createSystemGate()
		if err != nil {
			return err
		}

	}

	return f.SetActivated(gate, featureName)
}

func (f *featureGateRunner) DeactivateFeature(featureName string) error {
	gate, err := f.GetTKGSystemFeatureGate()
	if crtclient.IgnoreNotFound(err) != nil {
		return err
	}
	if err != nil {
		gate, err = f.createSystemGate()
		if err != nil {
			return err
		}

	}

	return f.SetDeactivated(gate, featureName)
}

func (f *featureGateRunner) createSystemGate() (*configv1alpha1.FeatureGate, error) {
	if err := f.ensureTKGNamespace(); crtclient.IgnoreNotFound(err) != nil {
		return nil, err
	}

	gate := &configv1alpha1.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tkgconstants.TKGSystemFeatureGateNamespace,
			Name:      tkgconstants.TKGSystemFeatureGate,
		},
	}
	//	Create(ctx context.Context, obj runtime.Object, opts ...CreateOption) error
	err := f.crtClient.Create(context.Background(), gate)
	if apierrors.IsAlreadyExists(err) {
		return gate, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "Creating TKG system gate")

	}
	return gate, nil
}

func (f *featureGateRunner) ensureTKGNamespace() error {
	ns := &corev1.Namespace{}
	err := f.crtClient.Get(context.Background(), crtclient.ObjectKey{
		Name: tkgconstants.TKGSystemFeatureGateNamespace,
	}, ns)
	if apierrors.IsNotFound(err) {
		return errors.Wrapf(err, "TKG system namespace %s does not exist", tkgconstants.TKGSystemFeatureGateNamespace)
	}
	if err != nil {
		return errors.Wrap(err, "Checking TKG system namespace")

	}

	return nil
}

func (f *featureGateRunner) SetDeactivated(gate *configv1alpha1.FeatureGate, featureName string) error {
	for i, featureRef := range gate.Spec.Features {
		if featureRef.Name == featureName {
			if featureRef.Activate == false {
				return nil
			}
			gate.Spec.Features[i].Activate = false

			return f.crtClient.Update(context.Background(), gate)
		}
	}

	ref := configv1alpha1.FeatureReference{
		Name:     featureName,
		Activate: false,
	}
	gate.Spec.Features = append(gate.Spec.Features, ref)
	return f.crtClient.Update(context.Background(), gate)
}

func (f *featureGateRunner) SetActivated(gate *configv1alpha1.FeatureGate, featureName string) error {
	for i, featureRef := range gate.Spec.Features {
		if featureRef.Name == featureName {
			if featureRef.Activate == true {
				return nil
			}
			gate.Spec.Features[i].Activate = true

			return f.crtClient.Update(context.Background(), gate)
		}
	}

	ref := configv1alpha1.FeatureReference{
		Name:     featureName,
		Activate: true,
	}
	gate.Spec.Features = append(gate.Spec.Features, ref)
	return f.crtClient.Update(context.Background(), gate)
}

func (f *featureGateRunner) GetTKGSystemFeatureGate() (*configv1alpha1.FeatureGate, error) {
	gate := &configv1alpha1.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tkgconstants.TKGSystemFeatureGateNamespace,
			Name:      tkgconstants.TKGSystemFeatureGate,
		},
	}

	err := f.crtClient.Get(context.Background(), crtclient.ObjectKey{
		Namespace: tkgconstants.TKGSystemFeatureGateNamespace,
		Name:      tkgconstants.TKGSystemFeatureGate,
	}, gate)

	if err != nil {
		return gate, err
	}
	return gate, nil
}

func (f *featureGateRunner) GetFeature(featureName string) (*configv1alpha1.Feature, error) {
	feature := &configv1alpha1.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: featureName,
		},
	}

	err := f.crtClient.Get(context.Background(), crtclient.ObjectKey{
		Name: featureName,
	}, feature)

	if err != nil {
		return feature, err
	}
	return feature, nil
}

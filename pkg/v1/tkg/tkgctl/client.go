// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	clusterctllogger "sigs.k8s.io/cluster-api/cmd/clusterctl/log"

	providergetterclient "github.com/vmware-tanzu/tanzu-framework/providers/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clientcreator"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/providerinterface"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigupdater"
	"github.com/vmware-tanzu/tanzu-framework/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
	providergetterclient "github.com/vmware-tanzu/tanzu-framework/providers/client"
)

type tkgctl struct {
	kubeconfig               string
	kubecontext              string
	configDir                string
	appConfig                types.AppConfig
	tkgClient                client.Client
	tkgConfigProvidersClient tkgconfigproviders.Client
	tkgBomClient             tkgconfigbom.Client
	tkgConfigUpdaterClient   tkgconfigupdater.Client
	tkgConfigPathsClient     tkgconfigpaths.Client
	providerGetter           providerinterface.ProviderInterface
	tkgConfigReaderWriter    tkgconfigreaderwriter.TKGConfigReaderWriter
	featureGateHelper        FeatureGateHelper
}

// LoggingOptions options to configure logging with tkgctl client
type LoggingOptions struct {
	// File log file name where you want logs to be written
	File string
	// Quietly if set logs will not be written to stdout/stderr
	Quietly bool
	// Verbosity number for the log level verbosity
	Verbosity int32
	// LogChannel if channel is set, writer will forward log messages to this log channel
	// The result of this will be in the format of 'LogData' struct mentioned in `pkg/log/type.go`
	LogChannel chan<- []byte
}

// Options options to create tkgctl client
type Options struct {
	// Management cluster kubeconfig
	KubeConfig string
	// Management cluster kubecontext
	KubeContext string
	// ConfigDir specifies the directory where tkg will store metadata
	ConfigDir string
	// LoggingOptions options to configure logging with tkgctl client
	LogOptions LoggingOptions
	// ProviderGetter provide provider getter interface if you want to use custom providers with tkg
	// if not set, default providers bundled with tkg library will be used
	ProviderGetter providerinterface.ProviderInterface
	// CustomizerOptions allows consumers to provider customizers for underlying client definitions
	CustomizerOptions types.CustomizerOptions
	// SettingsFile should only be provided if user want to override default TKG settings file path
	// by default `ConfigDir/config.yaml` file will be used to store tkg settings
	SettingsFile string
	// ForceUpdateTKGCompatibilityImage should only be provided if user want to explicitly force update
	// TKG compatibility image which would also update the tkg-bom and tkr-bom files
	ForceUpdateTKGCompatibilityImage bool
}

// New creates new tkgctl client
func New(options Options) (TKGClient, error) { //nolint:gocritic
	var err error

	// configure log options for tkg library
	configureLogging(options.LogOptions)

	if options.ConfigDir == "" {
		return nil, errors.New("config directory cannot be empty. Please provide config directory when creating tkgctl client")
	}

	if options.ProviderGetter == nil {
		options.ProviderGetter = getDefaultProviderGetter()
	}

	if options.CustomizerOptions.RegionManagerFactory == nil {
		options.CustomizerOptions = types.CustomizerOptions{
			RegionManagerFactory: region.NewFactory(),
		}
	}
	appConfig := types.AppConfig{
		TKGConfigDir:      options.ConfigDir,
		ProviderGetter:    options.ProviderGetter,
		CustomizerOptions: options.CustomizerOptions,
		TKGSettingsFile:   options.SettingsFile,
	}

	err = ensureTKGConfigFile(options.ConfigDir, options.ProviderGetter)
	if err != nil {
		return nil, err
	}

	allClients, err := clientcreator.CreateAllClients(appConfig, nil)
	if err != nil {
		return nil, err
	}

	var clusterKubeConfig *types.ClusterKubeConfig
	if options.KubeConfig != "" {
		clusterKubeConfig = &types.ClusterKubeConfig{
			File:    options.KubeConfig,
			Context: options.KubeContext,
		}
	}

	tkgClient, err := client.New(client.Options{
		ClusterCtlClient:         allClients.ClusterCtlClient,
		ReaderWriterConfigClient: allClients.ConfigClient,
		RegionManager:            allClients.RegionManager,
		TKGConfigDir:             options.ConfigDir,
		Timeout:                  constants.DefaultOperationTimeout,
		FeaturesClient:           allClients.FeaturesClient,
		TKGConfigProvidersClient: allClients.TKGConfigProvidersClient,
		TKGBomClient:             allClients.TKGBomClient,
		TKGConfigUpdater:         allClients.TKGConfigUpdaterClient,
		TKGPathsClient:           allClients.TKGConfigPathsClient,
		ClusterKubeConfig:        clusterKubeConfig,
		ClusterClientFactory:     clusterclient.NewClusterClientFactory(),
		FeatureFlagClient:        allClients.FeatureFlagClient,
	})
	if err != nil {
		return nil, err
	}

	// ensure BoM and Providers prerequisite files are extracted if missing
	err = ensureBoMandProvidersPrerequisite(options.ConfigDir, allClients.TKGConfigUpdaterClient, options.ForceUpdateTKGCompatibilityImage)
	if err != nil {
		return nil, errors.Wrap(err, "unable to ensure prerequisites")
	}
	tkgConfigFile, err := allClients.TKGConfigPathsClient.GetTKGConfigPath()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get TKG config file path")
	}
	// re-initialize the TKG config reader writer after the providers are updated
	// as the TKG config file would be updated too.
	err = tkgClient.TKGConfigReaderWriter().Init(tkgConfigFile)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize the TKG config reader writer")
	}

	// Set default BOM name to the config variables to use during template generation
	defaultBoMFileName, err := allClients.TKGBomClient.GetDefaultBoMFileName()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get default BOM file name")
	}
	allClients.ConfigClient.TKGConfigReaderWriter().Set(constants.ConfigVariableDefaultBomFile, defaultBoMFileName)
	clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
	return &tkgctl{
		configDir:                options.ConfigDir,
		kubeconfig:               options.KubeConfig,
		kubecontext:              options.KubeContext,
		appConfig:                appConfig,
		tkgBomClient:             allClients.TKGBomClient,
		tkgConfigUpdaterClient:   allClients.TKGConfigUpdaterClient,
		tkgConfigProvidersClient: allClients.TKGConfigProvidersClient,
		tkgConfigPathsClient:     allClients.TKGConfigPathsClient,
		tkgClient:                tkgClient,
		providerGetter:           options.ProviderGetter,
		tkgConfigReaderWriter:    allClients.ConfigClient.TKGConfigReaderWriter(),
		featureGateHelper:        NewFeatureGateHelper(&clusterClientOptions, options.KubeContext, options.KubeConfig),
	}, nil
}

func ensureTKGConfigFile(configDir string, providerGetter providerinterface.ProviderInterface) error {
	var err error

	lock, err := utils.GetFileLockWithTimeOut(filepath.Join(configDir, constants.LocalTanzuFileLock), utils.DefaultLockTimeout)
	if err != nil {
		return errors.Wrap(err, "cannot acquire lock for ensuring local files")
	}

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Warningf("cannot release lock for ensuring local files, reason: %v", err)
		}
	}()

	_, err = tkgconfigupdater.New(configDir, providerGetter, nil).EnsureTKGConfigFile()
	return err
}

func ensureBoMandProvidersPrerequisite(configDir string, tkgConfigUpdaterClient tkgconfigupdater.Client, forceUpdateBoMFiles bool) error {
	var err error

	lock, err := utils.GetFileLockWithTimeOut(filepath.Join(configDir, constants.LocalTanzuFileLock), utils.DefaultLockTimeout)
	if err != nil {
		return errors.Wrap(err, "cannot acquire lock for ensuring local files")
	}

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Warningf("cannot release lock for ensuring local files, reason: %v", err)
		}
	}()
	// ensure BOM files are extracted if missing
	err = tkgConfigUpdaterClient.EnsureBOMFiles(forceUpdateBoMFiles)
	if err != nil {
		return errors.Wrap(err, "unable to ensure tkg BOM file")
	}
	// ensure that `images` configuration gets updated correctly in tkg settings file
	err = tkgConfigUpdaterClient.EnsureConfigImages()
	if err != nil {
		return err
	}

	// ensure that providers templates are extracted and placed under ConfigDir/providers
	err = tkgConfigUpdaterClient.EnsureProviderTemplates()
	if err != nil {
		return err
	}

	return nil
}

func configureLogging(logOptions LoggingOptions) {
	// Set logfile to the tkg logger
	if logOptions.File != "" {
		log.SetFile(logOptions.File)
	}
	if logOptions.LogChannel != nil {
		log.SetChannel(logOptions.LogChannel)
	}
	log.QuietMode(logOptions.Quietly)
	log.SetVerbosity(logOptions.Verbosity)

	clusterctllogger.SetLogger(log.GetLogr())
}

func getDefaultProviderGetter() providerinterface.ProviderInterface {
	return providergetterclient.New()
}

// restoreAfterSettingTimeout Sets the timeout value to tkgClient and returns a function
// to restore it back to defaultOperationTimeout value. This function should most likely be
// used with defer call which sets the config timeout and resets it back to default value
// after the operation is complete.
// This is done because we want to be consistent with the timeout value we use when single
// instance of tkgctl client is used to invoke multiple functions.
func (t *tkgctl) restoreAfterSettingTimeout(currentTimeout time.Duration) func() {
	t.tkgClient.ConfigureTimeout(currentTimeout)
	return func() {
		t.tkgClient.ConfigureTimeout(constants.DefaultOperationTimeout)
	}
}

func (t *tkgctl) TKGConfigReaderWriter() tkgconfigreaderwriter.TKGConfigReaderWriter {
	return t.tkgConfigReaderWriter
}

// FeatureGateHelper returns feature gate helper to query feature gate
func (t *tkgctl) FeatureGateHelper() FeatureGateHelper {
	return t.featureGateHelper
}

func ensureConfigImages(configDir string, tkgConfigUpdater tkgconfigupdater.Client) error {
	var err error

	lock, err := utils.GetFileLockWithTimeOut(filepath.Join(configDir, constants.LocalTanzuFileLock), utils.DefaultLockTimeout)
	if err != nil {
		return errors.Wrap(err, "cannot acquire lock for ensuring local files")
	}

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Warningf("cannot release lock for ensuring local files, reason: %v", err)
		}
	}()

	return tkgConfigUpdater.EnsureConfigImages()
}

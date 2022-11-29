// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	netutils "k8s.io/utils/net"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/avi"
	"github.com/vmware-tanzu/tanzu-framework/tkg/aws"
	"github.com/vmware-tanzu/tanzu-framework/tkg/azure"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/tkg/vc"
)

// error code constants
const (
	ValidationErrorCode      = 1
	PacificInVC7ErrorCode    = 2
	PacificNotInVC7ErrorCode = 3
)

// AWS config variables
var (
	AWSNewVPCConfigVariables       = []string{constants.ConfigVariableAWSPublicNodeCIDR, constants.ConfigVariableAWSPrivateNodeCIDR, constants.ConfigVariableAWSVPCCIDR}
	AWSProdNewVPCConfigVariables   = []string{constants.ConfigVariableAWSPublicNodeCIDR1, constants.ConfigVariableAWSPrivateNodeCIDR1, constants.ConfigVariableAWSPublicNodeCIDR2, constants.ConfigVariableAWSPrivateNodeCIDR2}
	AWSSubnetIDConfigVariables     = []string{constants.ConfigVariableAWSPrivateSubnetID, constants.ConfigVariableAWSPublicSubnetID}
	AWSProdSubnetIDConfigVariables = []string{constants.ConfigVariableAWSPrivateSubnetID1, constants.ConfigVariableAWSPublicSubnetID1, constants.ConfigVariableAWSPrivateSubnetID2, constants.ConfigVariableAWSPublicSubnetID2}
	VsphereNodeCPUVarName          = []string{constants.ConfigVariableVsphereCPNumCpus, constants.ConfigVariableVsphereWorkerNumCpus}
	VsphereNodeMemVarName          = []string{constants.ConfigVariableVsphereCPMemMib, constants.ConfigVariableVsphereWorkerMemMib}
	VsphereNodeDiskVarName         = []string{constants.ConfigVariableVsphereCPDiskGib, constants.ConfigVariableVsphereWorkerDiskGib}

	AWSPrivateSubnetIDConfigVariables = []string{constants.ConfigVariableAWSPrivateSubnetID, constants.ConfigVariableAWSPrivateSubnetID1, constants.ConfigVariableAWSPrivateSubnetID2}
	AWSPublicSubnetIDConfigVariables  = []string{constants.ConfigVariableAWSPublicSubnetID, constants.ConfigVariableAWSPublicSubnetID1, constants.ConfigVariableAWSPublicSubnetID2}
)

var trueString = "true"

// VsphereResourceConfigKeys vsphere resource types
var VsphereResourceConfigKeys = []string{constants.ConfigVariableVsphereDatacenter, constants.ConfigVariableVsphereNetwork, constants.ConfigVariableVsphereResourcePool, constants.ConfigVariableVsphereDatastore, constants.ConfigVariableVsphereFolder}

// CNITypes supported CNI types
var CNITypes = map[string]bool{"calico": true, "antrea": true, "none": true}

// ValidationError defines error during config validation
type ValidationError struct {
	Message string
	Code    int
}

// NodeSizeOptions contains node size options specified by user
type NodeSizeOptions struct {
	Size             string
	ControlPlaneSize string
	WorkerSize       string
}

// Error returns error message from validation error
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates new validation error object
func NewValidationError(code int, text string) *ValidationError {
	return &ValidationError{
		Message: text,
		Code:    code,
	}
}

// DownloadBomFile downloads BoM file
func (c *TkgClient) DownloadBomFile(tkrName string) error {
	log.V(1).Infof("Downloading bom for TKr %q", tkrName)

	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "cannot get current management cluster context")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while listing tkg clusters")
	}

	namespace := constants.TkrNamespace
	if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
		//TODO: After CLI fully support package based LCM, "constants.TkrNamespace" should be updated to "tkg-system"
		namespace = "tkg-system"
	}

	tkrConfigMap := &corev1.ConfigMap{}
	if err := regionalClusterClient.GetResource(tkrConfigMap, tkrName, namespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			return errors.Errorf("ConfigMap for TKr name %q not available to download bom", tkrName)
		}

		return err
	}

	bomData := tkrConfigMap.BinaryData["bomContent"]

	bomDir, err := c.tkgConfigPathsClient.GetTKGBoMDirectory()
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(bomDir, "tkr-bom-"+tkrName+".yaml"), bomData, 0o600)
}

// ConfigureAndValidateTkrVersion takes tkrVersion, if empty fetches default tkr & k8s version from config
// and validates k8s version format is valid semantic version
func (c *TkgClient) ConfigureAndValidateTkrVersion(tkrVersion string) (string, string, error) {
	var k8sVersion string
	var err error

	// use default BoM file if tkrVersion is not provided
	if tkrVersion == "" {
		tkrBoMConfig, err := c.tkgBomClient.GetDefaultTkrBOMConfiguration()
		if err != nil {
			return "", "", errors.Wrap(err, "unable to get default TKr bom")
		}
		tkrVersion = tkrBoMConfig.Release.Version
		k8sVersion, err = tkgconfigbom.GetK8sVersionFromTkrBoM(tkrBoMConfig)
		if err != nil {
			return "", "", errors.Wrap(err, "unable to get default k8s version from TKr bom")
		}
	} else {
		// BoM downloading should only be required if user are passing tkrName,
		// otherwise we should use default config which is always present on user's machine

		// download bom if not present locally for given TKr
		_, err = c.tkgBomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
		if err != nil {
			_, ok := err.(tkgconfigbom.BomNotPresent)
			if ok { // bom not present locally
				if err := c.DownloadBomFile(utils.GetTkrNameFromTkrVersion(tkrVersion)); err != nil {
					return "", "", err
				}
			} else {
				return "", "", err
			}
		}
		k8sVersion, err = c.tkgBomClient.GetK8sVersionFromTkrVersion(tkrVersion)
		if err != nil {
			return "", "", err
		}
	}

	if !strings.HasPrefix(k8sVersion, "v") {
		return "", "", errors.Errorf("unsupported KubernetesVersion: %s. Kubernetes version should have prefix v", k8sVersion)
	}

	// Set tkrName and k8sVersion to the tkg config
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableKubernetesVersion, k8sVersion)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableTkrName, utils.GetTkrNameFromTkrVersion(tkrVersion))

	return k8sVersion, tkrVersion, nil
}

// ConfigureAzureVMImage configures azure image from BoM or user config file
func (c *TkgClient) ConfigureAzureVMImage(tkrVersion string) error {
	azureVMImage, err := c.tkgConfigProvidersClient.GetAzureVMImageInfo(tkrVersion)
	if err != nil {
		return err
	}

	if azureVMImage == nil {
		osInfo := tkgconfighelper.GetUserProvidedOsOptions(c.TKGConfigReaderWriter())
		return errors.Errorf("unable to find the azure vm image info for TKr version: '%v' and os options: '(%v,%v,%v)'", tkrVersion, osInfo.Name, osInfo.Version, osInfo.Arch)
	}

	// using image ID
	if azureVMImage.ID != "" {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImageID, azureVMImage.ID)
	}

	// using shared gallery image
	if isSharedGalleryImage(azureVMImage) {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImageResourceGroup, azureVMImage.ResourceGroup)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImageName, azureVMImage.Name)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImageSubscriptionID, azureVMImage.SubscriptionID)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImageGallery, azureVMImage.Gallery)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImageVersion, azureVMImage.Version)
	}

	// using market place image
	if isMarketplaceImage(azureVMImage) {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImagePublisher, azureVMImage.Publisher)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImageOffer, azureVMImage.Offer)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImageSku, azureVMImage.Sku)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImageVersion, azureVMImage.Version)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureImageThirdParty, strconv.FormatBool(azureVMImage.ThirdPartyImage))
	}

	if azureVMImage.ID != "" || isSharedGalleryImage(azureVMImage) || isMarketplaceImage(azureVMImage) {
		log.V(6).Infof("consuming Azure Image info: %v", *azureVMImage)

		// configure OS name, version and arch for the selected OS
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSName, azureVMImage.OSInfo.Name)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSVersion, azureVMImage.OSInfo.Version)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSArch, azureVMImage.OSInfo.Arch)
		return nil
	}

	return errors.Errorf("invalid azure image info: %v, for TKr version: %v", *azureVMImage, tkrVersion)
}

// ConfigureAndValidateAzureConfig configures and validates azure configurationn
func (c *TkgClient) ConfigureAndValidateAzureConfig(tkrVersion string, nodeSizes NodeSizeOptions, skipValidation bool, clusterClient clusterclient.Client) error {
	var client azure.Client
	var err error

	c.SetProviderType(AzureProviderName)

	if !skipValidation {
		client, err = c.EncodeAzureCredentialsAndGetClient(clusterClient)
		if err != nil {
			return errors.Wrapf(err, "failed to initialize Azure client")
		}

		err = c.ValidateAzurePublicSSHKey()
		if err != nil {
			return errors.Wrapf(err, "failed to validate %s", constants.ConfigVariableAzureSSHPublicKeyB64)
		}
	}

	if err := c.OverrideAzureNodeSizeWithOptions(client, nodeSizes, skipValidation); err != nil {
		return errors.Wrap(err, "cannot set Azure node size")
	}

	if tkrVersion == "" {
		return errors.New("TKr version is empty")
	}

	if err := c.ConfigureAzureVMImage(tkrVersion); err != nil {
		return err
	}

	return nil
}

// ValidateAzurePublicSSHKey validates AZURE_SSH_PUBLIC_KEY_B64 exists and is base64 encoded
func (c *TkgClient) ValidateAzurePublicSSHKey() error {
	sshKey, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureSSHPublicKeyB64)
	if err != nil {
		return errors.Errorf("config variable %s is missing", constants.ConfigVariableAzureSSHPublicKeyB64)
	}
	_, errD := base64.StdEncoding.DecodeString(sshKey)
	if errD != nil {
		return errors.Errorf("config variable %s was not properly base64 encoded", constants.ConfigVariableAzureSSHPublicKeyB64)
	}
	return nil
}

// ConfigureAndValidateDockerConfig configures and validates docker configuration
func (c *TkgClient) ConfigureAndValidateDockerConfig(tkrVersion string, nodeSizes NodeSizeOptions, skipValidation bool) error {
	c.SetProviderType(DockerProviderName)

	if tkrVersion == "" {
		return errors.New("TKr version is empty")
	}

	bomConfiguration, err := c.tkgBomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to get bom configuration for TKr version %s", tkrVersion)
	}

	kindNodeImage := bomConfiguration.Components["kubernetes-sigs_kind"][0].Images["kindNodeImage"]
	dockerTemplateImage := tkgconfigbom.GetFullImagePath(kindNodeImage, bomConfiguration.ImageConfig.ImageRepository) + ":" + kindNodeImage.Tag
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableDockerMachineTemplateImage, dockerTemplateImage)
	return nil
}

// ConfigureAndValidateAwsConfig configures and validates aws configuration
func (c *TkgClient) ConfigureAndValidateAwsConfig(tkrVersion string, workerMachineCount int64, useExistingVPC bool) error {
	if tkrVersion == "" {
		return errors.New("TKr version is empty")
	}

	bomConfiguration, err := c.tkgBomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to get bom configuration for TKr version %s", tkrVersion)
	}

	awsRegion, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSRegion)
	if err != nil {
		// all missing variables will be handled in a later step where clusterctl generates template and catch all of the missing variables
		return nil
	}

	if az, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSNodeAz); err == nil && !useExistingVPC {
		if !strings.HasPrefix(az, awsRegion) {
			return errors.Errorf("Node availability zone %s is not part of AWS region %s. Please check the AWS_REGION environment variable for possible conflict", az, awsRegion)
		}
	}

	if err := c.configureAMIAndOSForAWS(bomConfiguration, awsRegion); err != nil {
		return errors.Wrap(err, "unable to configure AMI and OS options")
	}

	// the WorkerMachineCount, if valid, can be used in templates using the ${ WORKER_MACHINE_COUNT } variable.
	if workerMachineCount < 0 {
		return errors.Errorf("invalid WorkerMachineCount. Please use a number greater or equal than 0")
	}

	return nil
}

func (c *TkgClient) configureAMIAndOSForAWS(bomConfiguration *tkgconfigbom.BOMConfiguration, awsRegion string) error {
	if awsAmiID, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSAMIID); awsAmiID != "" {
		log.Warningf("The '%s' variable is obsolete. The correct '%s' is obtained from appropriate BoM metadata file or user settings file based on given OS options", constants.ConfigVariableAWSAMIID, constants.ConfigVariableAWSAMIID)
	}

	amiInfo, err := c.tkgConfigProvidersClient.GetAWSAMIInfo(bomConfiguration, awsRegion)
	if err != nil || amiInfo == nil {
		// throw error
		return errors.Errorf("unable to get ami info, error: %v", err.Error())
	}

	log.V(3).Infof("using AMI '%v' for tkr version: '%v', aws-region '%v'", amiInfo.ID, bomConfiguration.Release.Version, awsRegion)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableAWSAMIID, amiInfo.ID)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSName, amiInfo.OSInfo.Name)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSVersion, amiInfo.OSInfo.Version)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSArch, amiInfo.OSInfo.Arch)
	return nil
}

func checkIfRequiredPermissionsPresent(awsClient aws.Client) {
	stacks, err := awsClient.ListCloudFormationStacks()
	if err != nil {
		log.Warningf("unable to verify if the AWS CloudFormation stack %s is available in the AWS account.", aws.DefaultCloudFormationStackName)
		return
	}

	for _, stack := range stacks {
		if stack == aws.DefaultCloudFormationStackName {
			return
		}
	}
	// TODO: should have check on whether IAM permissions are present
	log.Warningf("cannot find AWS CloudFormation stack %s, which is used in the management of IAM groups and policies required by TKG. You might need to create one manually before creating a cluster", aws.DefaultCloudFormationStackName)
}

// ConfigureAndValidateAWSConfig configures and validates aws configuration
func (c *TkgClient) ConfigureAndValidateAWSConfig(tkrVersion string, nodeSizes NodeSizeOptions, skipValidation, isProdConfig bool, workerMachineCount int64, clusterClient clusterclient.Client, isManagementCluster bool) error {
	c.SetProviderType(AWSProviderName)
	awsClient, err := c.EncodeAWSCredentialsAndGetClient(clusterClient)
	if err != nil {
		// We must have credentials present for the instantiation of the management cluster
		if isManagementCluster {
			return err
		}
		log.Warningf("unable to create AWS client. Skipping validations that require an AWS client")
		return c.ConfigureAndValidateAwsConfig(tkrVersion, workerMachineCount, false)
	}

	if !skipValidation {
		checkIfRequiredPermissionsPresent(awsClient)
	}

	if err := c.OverrideAWSNodeSizeWithOptions(nodeSizes, awsClient, skipValidation); err != nil {
		log.Warningf("unable to override node size")
	}

	useExistingVPC, err := c.SetAndValidateDefaultAWSVPCConfiguration(isProdConfig, awsClient, skipValidation)
	if err != nil {
		log.Warningf("unable to validate VPC configuration, %s", err.Error())
	}

	return c.ConfigureAndValidateAwsConfig(tkrVersion, workerMachineCount, useExistingVPC)
}

// TrimVsphereSSHKey trim the comment part of the vsphere ssh key
func (c *TkgClient) TrimVsphereSSHKey() {
	sshKeyPartsLen := 2
	sshKey, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereSSHAuthorizedKey)
	if err == nil {
		sshKeyParts := strings.Fields(sshKey)
		if len(sshKeyParts) >= sshKeyPartsLen {
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereSSHAuthorizedKey, strings.Join(sshKeyParts[0:sshKeyPartsLen], " "))
		}
	}
}

// ConfigureAndValidateVSphereTemplate validate the k8s version provided matches with VM template's k8s version
func (c *TkgClient) ConfigureAndValidateVSphereTemplate(vcClient vc.Client, tkrVersion, dc string) error {
	var err error
	if tkrVersion == "" {
		return errors.New("TKr version is empty")
	}

	templateName, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereTemplate)

	tkrBom, err := c.tkgBomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
	if err != nil {
		return err
	}

	vsphereVM, err := vcClient.GetAndValidateVirtualMachineTemplate(tkrBom.GetOVAVersions(), tkrVersion, templateName, dc, c.TKGConfigReaderWriter())
	if err != nil || vsphereVM == nil {
		return errors.Wrap(err, "unable to get or validate VM Template for given Tanzu Kubernetes release")
	}

	c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereTemplate, vsphereVM.Name)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSName, vsphereVM.DistroName)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSVersion, vsphereVM.DistroVersion)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSArch, vsphereVM.DistroArch)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereTemplateMoid, vsphereVM.Moid)
	return nil
}

// GetVSphereEndpoint gets vsphere client based on credentials set in config variables
func (c *TkgClient) GetVSphereEndpoint(clusterClient clusterclient.Client) (vc.Client, error) {
	if clusterClient != nil {
		regionContext, err := c.GetCurrentRegionContext()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get current region context")
		}
		username, password, err := clusterClient.GetVCCredentialsFromCluster(regionContext.ClusterName, constants.TkgNamespace)
		if err != nil {
			return nil, err
		}

		server, err := clusterClient.GetVCServer()
		if err != nil {
			return nil, err
		}

		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereServer, server)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereUsername, username)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVspherePassword, password)

		vsphereInsecureString, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereInsecure)
		vsphereInsecure := (vsphereInsecureString == trueString)
		vsphereThumbprint, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereTLSThumbprint)

		return vc.GetAuthenticatedVCClient(server, username, password, vsphereThumbprint, vsphereInsecure, c.vcClientFactory)
	}

	vcHost, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereServer)
	if err != nil {
		return nil, errors.Errorf("failed to get %s", constants.ConfigVariableVsphereServer)
	}
	vcUsername, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereUsername)
	if err != nil {
		return nil, errors.Errorf("failed to get %s", constants.ConfigVariableVsphereUsername)
	}
	vcPassword, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVspherePassword)
	if err != nil {
		return nil, errors.Errorf("failed to get %s", constants.ConfigVariableVspherePassword)
	}
	vsphereInsecure := false
	vsphereInsecureString, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereInsecure)
	if err == nil {
		vsphereInsecure = (vsphereInsecureString == trueString)
	}

	// If the VSPHERE_TLS_THUMBPRINT is not set, the default ssl certificate validation will be used.
	thumbprint, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereTLSThumbprint)

	host := strings.TrimSpace(vcHost)
	if !strings.HasPrefix(host, "http") {
		host = "https://" + host
	}
	vcURL, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse vc host")
	}
	vcURL.Path = "/sdk"
	vcClient, err := c.vcClientFactory.NewClient(vcURL, thumbprint, vsphereInsecure)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vc client")
	}
	_, err = vcClient.Login(context.TODO(), vcUsername, vcPassword)
	if err != nil {
		return nil, errors.Wrap(err, "failed to login to vSphere")
	}
	return vcClient, nil
}

// ConfigureAndValidateManagementClusterConfiguration configure and validate management cluster configuration
func (c *TkgClient) ConfigureAndValidateManagementClusterConfiguration(options *InitRegionOptions, skipValidation bool) *ValidationError { // nolint:gocyclo,funlen
	var err error
	if options.ClusterName != "" {
		if err := CheckClusterNameFormat(options.ClusterName, options.InfrastructureProvider); err != nil {
			return NewValidationError(ValidationErrorCode, err.Error())
		}
	}

	regions, err := c.regionManager.ListRegionContexts()
	if err != nil {
		return NewValidationError(ValidationErrorCode, "unable to verify cluster name uniqueness")
	}

	for _, region := range regions {
		if region.ClusterName == options.ClusterName {
			errMsg := fmt.Sprintf("cluster name %s matches another management cluster", options.ClusterName)
			return NewValidationError(ValidationErrorCode, errMsg)
		}
	}

	if options.Plan == "" {
		return NewValidationError(ValidationErrorCode, "required config variable 'CLUSTER_PLAN' is not set")
	}

	_, tkrVersion, err := c.ConfigureAndValidateTkrVersion("")
	if err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	// BUILD_EDITION is the Tanzu Edition, the plugin should be built for. Its value is supposed be constructed from
	// cmd/cli/plugin/managementcluster/create.go. So empty value at this point is not expected.
	if options.Edition == "" {
		return NewValidationError(ValidationErrorCode, "required config variable 'edition' is not set")
	}
	c.SetBuildEdition(options.Edition)

	c.SetTKGClusterRole(ManagementCluster)
	c.SetTKGVersion()

	idpType, err := c.TKGConfigReaderWriter().Get("IDENTITY_MANAGEMENT_TYPE")
	if err != nil || idpType == "" || idpType == "none" {
		log.Warningf("Identity Provider not configured. Some authentication features won't work.")
	}

	options.InfrastructureProvider, err = c.tkgConfigUpdaterClient.CheckInfrastructureVersion(options.InfrastructureProvider)
	if err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "unable to check infrastructure provider version").Error())
	}

	err = c.ConfigureAndValidateCNIType(options.CniType)
	if err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "unable to validate CNI type").Error())
	}

	name, _, err := ParseProviderName(options.InfrastructureProvider)
	if err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err = c.configureAndValidateIPFamilyConfiguration(TkgLabelClusterRoleManagement); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err = c.configureAndValidateCoreDNSIP(); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err = c.validateServiceCIDRNetmask(); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err = c.ConfigureAndValidateHTTPProxyConfiguration(name); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err = c.ConfigureAndValidateNameserverConfiguration(TkgLabelClusterRoleManagement); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err = c.ConfigureAndValidateAviConfiguration(); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	isProdPlan := IsProdPlan(options.Plan)
	_, workerMachineCount := c.getMachineCountForMC(options.Plan)

	switch name {
	case AWSProviderName:
		err = c.ConfigureAndValidateAWSConfig(tkrVersion, options.NodeSizeOptions, skipValidation, isProdPlan, int64(workerMachineCount), nil, true)
	case VSphereProviderName:
		err := c.ConfigureAndValidateVsphereConfig(tkrVersion, options.NodeSizeOptions, options.VsphereControlPlaneEndpoint, skipValidation, nil)
		if err != nil {
			return NewValidationError(ValidationErrorCode, err.Error())
		}
		err = c.ValidateVsphereControlPlaneEndpointIP(options.VsphereControlPlaneEndpoint)
		if err != nil {
			log.Warningf("WARNING: The control plane endpoint '%s' might already used by other cluster. This might affect the deployment of the cluster", options.VsphereControlPlaneEndpoint)
		}
	case AzureProviderName:
		err = c.ConfigureAndValidateAzureConfig(tkrVersion, options.NodeSizeOptions, skipValidation, nil)
	case DockerProviderName:
		err = c.ConfigureAndValidateDockerConfig(tkrVersion, options.NodeSizeOptions, skipValidation)
	}

	if err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	workerCounts, err := c.DistributeMachineDeploymentWorkers(int64(workerMachineCount), isProdPlan, true, name, false)
	if err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "failed to distribute machine deployments").Error())
	}
	c.SetMachineDeploymentWorkerCounts(workerCounts, int64(workerMachineCount), isProdPlan)

	return nil
}

// ValidateVsphereControlPlaneEndpointIP validates if the control plane endpoint has been used by another cluster in the same network
func (c *TkgClient) ValidateVsphereControlPlaneEndpointIP(endpointIP string) *ValidationError {
	log.V(6).Infof("Checking if VSPHERE_CONTROL_PLANE_ENDPOINT %s is already in use", endpointIP)
	currentNetwork, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereNetwork)
	if err != nil {
		return NewValidationError(ValidationErrorCode, "unable to read network name from the configs")
	}

	currentServer, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereServer)
	if err != nil {
		return NewValidationError(ValidationErrorCode, "unable to read vsphere server from the configs")
	}

	regions, _ := c.GetRegionContexts("")
	for _, regionContext := range regions {
		regionalClusterClient, err := c.getRegionClient(regionContext)
		if err != nil {
			log.V(6).Infof("Unable to create regionalClient")
			continue
		}

		vSphereMachineTemplate, err := getVsphereMachineTemplate(regionalClusterClient, regionContext.ClusterName)
		if err != nil {
			log.V(6).Infof("Unable to find Network name for context %s. Skipping validation for this context", regionContext.ContextName)
			continue
		}

		network := vSphereMachineTemplate.Spec.Template.Spec.Network.Devices[0].NetworkName
		server := vSphereMachineTemplate.Spec.Template.Spec.Server

		log.V(4).Infof("Network name: %s", network)

		if currentNetwork == network && currentServer == server {
			log.V(6).Infof("Network names, and server matched, validating...")
			managementClusters, err := regionalClusterClient.ListClusters(TKGsystemNamespace)
			if err != nil {
				log.V(6).Infof("Unable to list management clusters")
			}
			workloadClusters, err := regionalClusterClient.ListClusters("")
			if err != nil {
				log.V(6).Infof("Unable to list workload clusters")
				continue
			}

			clusters := append(managementClusters, workloadClusters...) //nolint:gocritic

			for i := range clusters {
				if clusters[i].Spec.ControlPlaneEndpoint.Host == endpointIP {
					return NewValidationError(ValidationErrorCode, "Control plane endpoint already exists")
				}
			}
		}
	}
	return nil
}

func (c *TkgClient) getRegionClient(regionContext region.RegionContext) (clusterclient.Client, error) {
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
		OperationTimeout:  c.timeout,
	}

	log.V(4).Infof("SourceFilePath: %s, ContextName: %s", regionContext.SourceFilePath, regionContext.ContextName)
	currentKubeConfig := clusterctlclient.Kubeconfig{Path: regionContext.SourceFilePath, Context: regionContext.ContextName}
	client, err := clusterclient.NewClient(currentKubeConfig.Path, currentKubeConfig.Context, clusterclientOptions)
	if err != nil {
		return nil, NewValidationError(ValidationErrorCode, "unable to get cluster client while creating cluster")
	}

	return client, nil
}

func getVsphereMachineTemplate(client clusterclient.Client, clusterName string) (*capvv1beta1.VSphereMachineTemplate, error) {
	vsphereMachineTemplate := &capvv1beta1.VSphereMachineTemplate{}
	nameSpace, err := client.GetCurrentNamespace()
	if err != nil {
		return nil, err
	}
	log.V(4).Infof("Namespace: %s, Cluster Name: %s", nameSpace, clusterName)
	kcp, err := client.GetKCPObjectForCluster(clusterName, "tkg-system")
	if err != nil {
		log.V(4).Infof("Error getting KCP Object")
		return nil, err
	}
	if err := client.GetResource(vsphereMachineTemplate, kcp.Spec.MachineTemplate.InfrastructureRef.Name, "tkg-system", nil, nil); err != nil {
		return nil, err
	}
	return vsphereMachineTemplate, nil
}

// ConfigureAndValidateVsphereConfig configures and validates vsphere configuration
func (c *TkgClient) ConfigureAndValidateVsphereConfig(tkrVersion string, nodeSizes NodeSizeOptions, vip string, skipValidation bool, clusterClient clusterclient.Client) *ValidationError {
	if err := c.OverrideVsphereNodeSizeWithOptions(nodeSizes); err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "unable to set vSphere node size").Error())
	}
	// set node size values that are still missing after overriding with user options
	c.SetVsphereNodeSize()
	c.SetProviderType(VSphereProviderName)

	if err := c.ValidateVsphereNodeSize(); err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "vSphere node size validation failed").Error())
	}

	// trim the ssh key in the last part of the validation step so that a trimmed key is used for payload
	c.TrimVsphereSSHKey()

	// configure and validate vip for vsphere cluster
	if err := c.configureAndValidateVIPForVsphereCluster(vip); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	// if skipValidation is true skip the subsequent validation of talking to VC
	if skipValidation {
		return nil
	}

	vcClient, err := c.GetVSphereEndpoint(clusterClient)
	if err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "failed to get VC client").Error())
	}

	dc, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereDatacenter)
	if err != nil {
		return NewValidationError(ValidationErrorCode, errors.Errorf("failed to get %s", constants.ConfigVariableVsphereDatacenter).Error())
	}

	if err := c.ConfigureAndValidateVSphereTemplate(vcClient, tkrVersion, dc); err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "vSphere template kubernetes version validation failed").Error())
	}

	if err := c.ValidateVsphereResources(vcClient, dc); err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "vSphere resources validation failed").Error())
	}

	vsphereVersion, _, err := vcClient.GetVSphereVersion()
	if err != nil {
		return NewValidationError(ValidationErrorCode, errors.Errorf("failed to get vSphere version from VC client").Error())
	}

	c.SetVsphereVersion(vsphereVersion)

	return nil
}

// configure and validate vip for vsphere cluster
func (c *TkgClient) configureAndValidateVIPForVsphereCluster(vip string) error {
	haProvider, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereHaProvider)
	if vip != "" {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereControlPlaneEndpoint, vip)
	} else {
		vip, _ = c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint)
	}
	if vip == "" && haProvider != trueString {
		// for backward compatibility check _VSPHERE_CONTROL_PLANE_ENDPOINT variable as well
		vip, _ = c.TKGConfigReaderWriter().Get("_VSPHERE_CONTROL_PLANE_ENDPOINT")
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereControlPlaneEndpoint, vip)
	}

	if vip == "" && haProvider != trueString {
		return NewValidationError(ValidationErrorCode, errors.Errorf("'%s' config variable is required for infrastructure provider vsphere", constants.ConfigVariableVsphereControlPlaneEndpoint).Error())
	}

	if vip != "" {
		vsphereServer, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereServer)
		if vsphereServer == vip {
			return NewValidationError(ValidationErrorCode, "The vSphere Control Plane Endpoint should not match the vSphere Server address")
		}

		vsphereIP := net.ParseIP(vsphereServer)
		if vIP := net.ParseIP(vip); vsphereIP != nil && vIP != nil {
			if vIP.Equal(vsphereIP) {
				return NewValidationError(ValidationErrorCode, "vSphere Server IP should be different from the vSphere Control Plane Endpoint")
			}
		}
	}

	return nil
}

// ValidateVsphereVipWorkloadCluster validates that the control plane endpoint is unique
func (c *TkgClient) ValidateVsphereVipWorkloadCluster(clusterClient clusterclient.Client, vip string, skipValidation bool) error {
	if skipValidation {
		return nil
	}
	clusterList, err := clusterClient.ListClusters("")
	if err != nil {
		return errors.Wrapf(err, "failed to get clusters")
	}
	for i := range clusterList {
		if clusterList[i].Spec.ControlPlaneEndpoint.Host == vip {
			return errors.Errorf("control plane endpoint '%s' already in use by cluster '%s' and cannot be reused", vip, clusterList[i].ObjectMeta.Name)
		}
	}
	return nil
}

// ValidateVsphereResources validates vsphere resource path specified in tkgconfig
func (c *TkgClient) ValidateVsphereResources(vcClient vc.Client, dcPath string) error {
	var path, resourceMoid string
	var err error

	for _, vsphereResourceConfigKey := range VsphereResourceConfigKeys {
		path, err = c.TKGConfigReaderWriter().Get(vsphereResourceConfigKey)
		if err != nil {
			continue
		}

		// finder return an error when multiple vsphere resources with the same name are present
		switch vsphereResourceConfigKey {
		case constants.ConfigVariableVsphereDatacenter:
			resourceMoid, err = vcClient.FindDataCenter(context.Background(), dcPath)
			if err != nil {
				return errors.Wrapf(err, "invalid %s", vsphereResourceConfigKey)
			}
		case constants.ConfigVariableVsphereNetwork:
			resourceMoid, err = vcClient.FindNetwork(context.Background(), path, dcPath)
			if err != nil {
				return errors.Wrapf(err, "invalid %s", vsphereResourceConfigKey)
			}
		case constants.ConfigVariableVsphereResourcePool:
			resourceMoid, err = vcClient.FindResourcePool(context.Background(), path, dcPath)
			if err != nil {
				return errors.Wrapf(err, "invalid %s", vsphereResourceConfigKey)
			}
		case constants.ConfigVariableVsphereDatastore:
			if path != "" {
				resourceMoid, err = vcClient.FindDatastore(context.Background(), path, dcPath)
				if err != nil {
					return errors.Wrapf(err, "invalid %s", vsphereResourceConfigKey)
				}
			}
		case constants.ConfigVariableVsphereFolder:
			resourceMoid, err = vcClient.FindFolder(context.Background(), path, dcPath)
			if err != nil {
				return errors.Wrapf(err, "invalid %s", vsphereResourceConfigKey)
			}

		default:
			return errors.Errorf("unknown vsphere resource type %s", vsphereResourceConfigKey)
		}

		err = c.setFullPath(vcClient, vsphereResourceConfigKey, path, resourceMoid)
		if err != nil {
			return err
		}
	}

	return c.verifyDatastoreOrStoragePolicySet()
}

// set full inventory path if the config variable value is not already an absolute path string or if it is not MOID
func (c *TkgClient) setFullPath(vcClient vc.Client, vsphereResourceConfigKey, path, resourceMoid string) error {
	if path != "" && !strings.HasPrefix(path, "/") && !strings.Contains(path, resourceMoid) {
		resourcePath, _, err := vcClient.GetPath(context.Background(), resourceMoid)
		if err != nil {
			return err
		}

		if resourcePath != path {
			log.Infof("Setting config variable %q to value %q", vsphereResourceConfigKey, resourcePath)
			c.TKGConfigReaderWriter().Set(vsphereResourceConfigKey, resourcePath)
		}
	}

	return nil
}

func (c *TkgClient) verifyDatastoreOrStoragePolicySet() error {
	dataStore, dataStoreErr := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereDatastore)
	storagePolicy, storagePolicyErr := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereStoragePolicyID)

	if (dataStoreErr != nil || dataStore == "") && (storagePolicyErr != nil || storagePolicy == "") {
		return errors.Errorf("Neither %s or %s are set. At least one of them needs to be set", constants.ConfigVariableVsphereDatastore, constants.ConfigVariableVsphereStoragePolicyID)
	}

	return nil
}

// ValidateVSphereVersion validates vsphere version
func ValidateVSphereVersion(vcClient vc.Client) *ValidationError {
	version, build, err := vcClient.GetVSphereVersion()
	if err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "unable to verify vSphere version").Error())
	}

	if strings.HasPrefix(version, "7.") {
		hasPacific, err := vcClient.DetectPacific(context.TODO())

		if err == nil && hasPacific {
			return NewValidationError(PacificInVC7ErrorCode, "the vSphere has version higher or equal to 7.0 with Management Kubernetes cluster deployed")
		}

		return NewValidationError(PacificNotInVC7ErrorCode, "the vSphere has version higher or equal to 7.0")
	}

	versions := strings.Split(version, ".")
	for i, v := range versions {
		num, err := strconv.Atoi(v)
		if err != nil {
			return NewValidationError(ValidationErrorCode, errors.Wrap(err, "invalid vSphere version").Error())
		}

		if num < vsphereVersionMinimumRequirement[i] {
			return NewValidationError(ValidationErrorCode, vsphereVersionError)
		} else if num > vsphereVersionMinimumRequirement[i] {
			return nil
		}
	}

	buildNum, err := strconv.Atoi(build)
	if err != nil {
		return NewValidationError(ValidationErrorCode, errors.Wrap(err, "invalid vSphere build number").Error())
	}

	if buildNum < vsphereBuildMinimumRequirement {
		return NewValidationError(ValidationErrorCode, vsphereVersionError)
	}

	return nil
}

// SetAndValidateDefaultAWSVPCConfiguration sets default value for AWS configuration variables
// Depending on whether AWS_VPC_ID is set/unset in the tkgconfig,
// values pertaining to configuring for existing/new VPC needs to be initialized to ""
// It returns whether creating cluster within existing vpc.
func (c *TkgClient) SetAndValidateDefaultAWSVPCConfiguration(isProdConfig bool, awsClient aws.Client, skipValidation bool) (bool, error) { // nolint:gocyclo
	useExistingVPC := false
	var err error
	vpcID := ""
	if vpcID, err = c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSVPCID); err != nil || vpcID == "" {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAWSVPCID, "")
		for _, varName := range AWSSubnetIDConfigVariables {
			if _, err := c.TKGConfigReaderWriter().Get(varName); err != nil {
				c.TKGConfigReaderWriter().Set(varName, "")
			}
		}

		if isProdConfig {
			for _, varName := range AWSProdSubnetIDConfigVariables {
				if _, err := c.TKGConfigReaderWriter().Get(varName); err != nil {
					c.TKGConfigReaderWriter().Set(varName, "")
				}
			}
		}
		return useExistingVPC, nil
	}

	useExistingVPC = true

	for _, varName := range AWSNewVPCConfigVariables {
		c.TKGConfigReaderWriter().Set(varName, "")
	}

	if isProdConfig {
		for _, varName := range AWSProdNewVPCConfigVariables {
			c.TKGConfigReaderWriter().Set(varName, "")
		}
	}

	subnetIDs := make(map[string]bool)
	if !skipValidation {
		subnets, err := awsClient.ListSubnets(vpcID)
		if err != nil {
			return useExistingVPC, err
		}

		for _, subnet := range subnets {
			subnetIDs[subnet.ID] = true
		}
	}

	missingConfigVar := []string{}

	nonExistingSubnetIDs := []string{}

	for i, varName := range AWSPrivateSubnetIDConfigVariables {
		if !isProdConfig && i == 1 {
			break
		}
		if val, err := c.TKGConfigReaderWriter().Get(varName); err != nil || val == "" {
			missingConfigVar = append(missingConfigVar, varName)
		} else if !skipValidation {
			if _, ok := subnetIDs[val]; !ok {
				nonExistingSubnetIDs = append(nonExistingSubnetIDs, val)
			}
		}
	}

	if publicSubnetID, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPublicSubnetID); err != nil || publicSubnetID != "" {
		// validate public subnets when there the vpc is public facing
		for i, varName := range AWSPublicSubnetIDConfigVariables {
			if !isProdConfig && i == 1 {
				break
			}
			if val, err := c.TKGConfigReaderWriter().Get(varName); err != nil || val == "" {
				missingConfigVar = append(missingConfigVar, varName)
			} else if !skipValidation {
				if _, ok := subnetIDs[val]; !ok {
					nonExistingSubnetIDs = append(nonExistingSubnetIDs, val)
				}
			}
		}
	} else {
		log.Warning("public subnet ID(s) not found")

		if publicSubentID1, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPublicSubnetID1); err == nil && publicSubentID1 != "" {
			return useExistingVPC, errors.Errorf("%s cannot be used without %s", constants.ConfigVariableAWSPublicSubnetID1, constants.ConfigVariableAWSPublicSubnetID)
		}

		if publicSubentID2, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSPublicSubnetID2); err == nil && publicSubentID2 != "" {
			return useExistingVPC, errors.Errorf("%s cannot be used without %s", constants.ConfigVariableAWSPublicSubnetID2, constants.ConfigVariableAWSPublicSubnetID)
		}
	}

	if len(missingConfigVar) != 0 {
		return useExistingVPC, errors.Errorf("configuration variable(s) %s not set", strings.Join(missingConfigVar, ","))
	}

	if len(nonExistingSubnetIDs) != 0 {
		return useExistingVPC, errors.Errorf("cannot find subnet(s) %s in VPC %s", strings.Join(nonExistingSubnetIDs, ","), vpcID)
	}

	return useExistingVPC, nil
}

func (c *TkgClient) checkVsphereNodeSize(key string, minReq int) error {
	var val string
	var err error

	if val, err = c.TKGConfigReaderWriter().Get(key); err != nil || val == "" {
		// if key cannot be found, leave to the later configration generating step to throw out error
		return nil
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return errors.Wrapf(err, "invalid %s", key)
	}

	if intVal < minReq {
		return errors.Errorf("the minimum requirement of %s is %d", key, minReq)
	}

	return nil
}

// ValidateVsphereNodeSize validates vsphere node size
func (c *TkgClient) ValidateVsphereNodeSize() error {
	minReqNodeSize := tkgconfigproviders.NodeTypes["small"]

	minCPU, _ := strconv.Atoi(minReqNodeSize.Cpus)
	minMem, _ := strconv.Atoi(minReqNodeSize.Memory)
	minDisk, _ := strconv.Atoi(minReqNodeSize.Disk)

	for _, varName := range VsphereNodeCPUVarName {
		if err := c.checkVsphereNodeSize(varName, minCPU); err != nil {
			return err
		}
	}

	for _, varName := range VsphereNodeMemVarName {
		if err := c.checkVsphereNodeSize(varName, minMem); err != nil {
			return err
		}
	}

	for _, varName := range VsphereNodeDiskVarName {
		if err := c.checkVsphereNodeSize(varName, minDisk); err != nil {
			return err
		}
	}

	return nil
}

// SetVsphereNodeSize sets vsphere node size
func (c *TkgClient) SetVsphereNodeSize() {
	// get the base node type
	cpu, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereNumCpus)
	memory, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereMemMib)
	disk, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereDiskGib)

	for _, varName := range VsphereNodeCPUVarName {
		if val, err := c.TKGConfigReaderWriter().Get(varName); err != nil || val == "" {
			if cpu != "" {
				c.TKGConfigReaderWriter().Set(varName, cpu)
			}
		}
	}

	for _, varName := range VsphereNodeMemVarName {
		if val, err := c.TKGConfigReaderWriter().Get(varName); err != nil || val == "" {
			if memory != "" {
				c.TKGConfigReaderWriter().Set(varName, memory)
			}
		}
	}

	for _, varName := range VsphereNodeDiskVarName {
		if val, err := c.TKGConfigReaderWriter().Get(varName); err != nil || val == "" {
			if disk != "" {
				c.TKGConfigReaderWriter().Set(varName, disk)
			}
		}
	}
}

// OverrideAzureNodeSizeWithOptions overrides azure node size with options
func (c *TkgClient) OverrideAzureNodeSizeWithOptions(client azure.Client, options NodeSizeOptions, skipValidation bool) error {
	location, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureLocation)
	if err != nil {
		return nil
	}

	azureNodeTypes := map[string]bool{}
	azureNodeSizeOptions := make([]string, 0)

	if !skipValidation {
		instanceTypes, err := client.GetAzureInstanceTypesForRegion(context.Background(), location)
		if err != nil {
			return errors.Wrapf(err, "unable to list the available node size options for Azure")
		}

		for _, instanceType := range instanceTypes {
			azureNodeTypes[instanceType.Name] = true
			azureNodeSizeOptions = append(azureNodeSizeOptions, instanceType.Name)
		}
	}

	if options.Size != "" {
		if _, ok := azureNodeTypes[options.Size]; !skipValidation && !ok {
			return errors.Errorf("node size %s cannot be used with TKG, please select among [%s]", options.Size, strings.Join(azureNodeSizeOptions, ","))
		}

		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureCPMachineType, options.Size)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureNodeMachineType, options.Size)
	}

	if options.ControlPlaneSize != "" {
		if _, ok := azureNodeTypes[options.ControlPlaneSize]; !skipValidation && !ok {
			return errors.Errorf("node size %s cannot be used with TKG, please select among [%s]", options.Size, strings.Join(azureNodeSizeOptions, ","))
		}
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureCPMachineType, options.ControlPlaneSize)
	}

	if options.WorkerSize != "" {
		if _, ok := azureNodeTypes[options.WorkerSize]; !skipValidation && !ok {
			return errors.Errorf("node size %s cannot be used with TKG, please select among [%s]", options.Size, strings.Join(azureNodeSizeOptions, ","))
		}
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureNodeMachineType, options.WorkerSize)
	}

	return nil
}

// OverrideAWSNodeSizeWithOptions overrides aws node size with options
func (c *TkgClient) OverrideAWSNodeSizeWithOptions(options NodeSizeOptions, awsClient aws.Client, skipValidation bool) error {
	if options.Size != "" {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableCPMachineType, options.Size)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableNodeMachineType, options.Size)
	}
	if options.ControlPlaneSize != "" {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableCPMachineType, options.ControlPlaneSize)
	}
	if options.WorkerSize != "" {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableNodeMachineType, options.WorkerSize)
	}

	if !skipValidation {
		err := c.validateAwsInstanceTypes(awsClient)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *TkgClient) validateAwsInstanceTypes(awsClient aws.Client) error {
	awsRegion, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSRegion)
	if err != nil {
		return nil
	}
	nodeTypes, err := awsClient.ListInstanceTypes("")
	if err != nil {
		return err
	}
	nodeMap := make(map[string]bool)
	for _, t := range nodeTypes {
		nodeMap[t] = true
	}

	controlplaneMachineType, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCPMachineType)
	if err != nil {
		return err
	}
	if _, ok := nodeMap[controlplaneMachineType]; !ok {
		return errors.Errorf("instance type %s is not supported in region %s", controlplaneMachineType, awsRegion)
	}

	var nodeMachineTypes []string
	nodeMachineType, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableNodeMachineType)
	if err != nil {
		return err
	}
	nodeMachineTypes = append(nodeMachineTypes, nodeMachineType)

	nodeMachineType1, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableNodeMachineType1)
	if err != nil {
		log.Infof("NODE_MACHINE_TYPE_1 not set, using the default NODE_MACHINE_TYPE instead")
	} else if nodeMachineType1 != "" {
		nodeMachineTypes = append(nodeMachineTypes, nodeMachineType1)
	}
	nodeMachineType2, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableNodeMachineType2)
	if err != nil {
		log.Infof("NODE_MACHINE_TYPE_2 not set, using the default NODE_MACHINE_TYPE instead")
	} else if nodeMachineType2 != "" {
		nodeMachineTypes = append(nodeMachineTypes, nodeMachineType2)
	}

	for _, machineType := range nodeMachineTypes {
		if _, ok := nodeMap[machineType]; !ok {
			return errors.Errorf("instance type %s is not supported in region %s", nodeMachineType, awsRegion)
		}
	}

	return nil
}

// OverrideVsphereNodeSizeWithOptions overrides vsphere node size with options
func (c *TkgClient) OverrideVsphereNodeSizeWithOptions(options NodeSizeOptions) error {
	if options.Size != "" {
		nodeSizes, ok := tkgconfigproviders.NodeTypes[options.Size]
		if !ok {
			return errors.Errorf("node size %s is not defined, please select among %s", options.Size, tkgconfigproviders.GetVsphereNodeSizeOptions())
		}

		for _, varName := range VsphereNodeCPUVarName {
			c.TKGConfigReaderWriter().Set(varName, nodeSizes.Cpus)
		}

		for _, varName := range VsphereNodeMemVarName {
			c.TKGConfigReaderWriter().Set(varName, nodeSizes.Memory)
		}

		for _, varName := range VsphereNodeDiskVarName {
			c.TKGConfigReaderWriter().Set(varName, nodeSizes.Disk)
		}
	}

	if options.ControlPlaneSize != "" {
		nodeSizes, ok := tkgconfigproviders.NodeTypes[options.ControlPlaneSize]
		if !ok {
			return errors.Errorf("node size %s is not defined, please select among %s", options.ControlPlaneSize, tkgconfigproviders.GetVsphereNodeSizeOptions())
		}

		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereCPNumCpus, nodeSizes.Cpus)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereCPMemMib, nodeSizes.Memory)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereCPDiskGib, nodeSizes.Disk)
	}

	if options.WorkerSize != "" {
		nodeSizes, ok := tkgconfigproviders.NodeTypes[options.WorkerSize]
		if !ok {
			return errors.Errorf("node size %s is not defined, please select among %s", options.WorkerSize, tkgconfigproviders.GetVsphereNodeSizeOptions())
		}

		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereWorkerNumCpus, nodeSizes.Cpus)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereWorkerMemMib, nodeSizes.Memory)
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereWorkerDiskGib, nodeSizes.Disk)
	}
	return nil
}

// SetTKGClusterRole sets the value of label tkg.tanzu.vmware.com/cluster-role
// for CAPI Cluster object.
func (c *TkgClient) SetTKGClusterRole(clusterType TKGClusterType) {
	forceRole, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableForceRole)
	if err == nil {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterRole, forceRole)
		return
	}

	switch clusterType {
	case ManagementCluster:
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterRole, TkgLabelClusterRoleManagement)
	case WorkloadCluster:
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterRole, TkgLabelClusterRoleWorkload)
	default:
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterRole, "")
	}
}

// EncodeAzureCredentialsAndGetClient encodes azure credentials and returns azure client
func (c *TkgClient) EncodeAzureCredentialsAndGetClient(clusterClient clusterclient.Client) (azure.Client, error) {
	var creds azure.Credentials
	var err error
	if clusterClient != nil {
		creds, err = clusterClient.GetAzureCredentialsFromSecret()
		if err != nil {
			return nil, err
		}
	} else {
		subscriptionID, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureSubscriptionID)
		if err != nil {
			return nil, errors.Errorf("failed to get Azure Subscription ID")
		}

		tenantID, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureTenantID)
		if err != nil {
			return nil, errors.Errorf("failed to get Azure Tenant ID")
		}

		clientID, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureClientID)
		if err != nil {
			return nil, errors.Errorf("failed to get Azure Client ID")
		}

		clientSecret, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureClientSecret)
		if err != nil {
			return nil, errors.Errorf("failed to get Azure Client Secret")
		}

		creds = azure.Credentials{
			SubscriptionID: subscriptionID,
			ClientID:       clientID,
			ClientSecret:   clientSecret,
			TenantID:       tenantID,
		}
	}
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureSubscriptionIDB64, base64.StdEncoding.EncodeToString([]byte(creds.SubscriptionID)))
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureTenantIDB64, base64.StdEncoding.EncodeToString([]byte(creds.TenantID)))
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureClientSecretB64, base64.StdEncoding.EncodeToString([]byte(creds.ClientSecret)))
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableAzureClientIDB64, base64.StdEncoding.EncodeToString([]byte(creds.ClientID)))

	azureCloud, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureEnvironment)
	if err != nil {
		log.V(6).Info("Setting config 'AZURE_ENVIRONMENT' to 'AzurePublicCloud'")
		azureCloud = azure.PublicCloud
	}
	creds.AzureCloud = azureCloud

	azureClient, err := azure.New(&creds)
	if err != nil {
		return nil, err
	}

	return azureClient, nil
}

// ConfigureAndValidateCNIType configures and validates cni
func (c *TkgClient) ConfigureAndValidateCNIType(cniType string) error {
	if cniType == "" {
		// if CNI not provided by CLI, check config
		cniType, _ = c.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)
		if cniType == "" {
			// if CNI not provided in config, use default
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableCNI, constants.DefaultCNIType)
			return nil
		}
	}
	// validate and set the provided CNI type
	if _, ok := CNITypes[cniType]; !ok {
		return errors.Errorf("provided CNI type '%s' is not in the available options: antrea, calico, none", cniType)
	}
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableCNI, cniType)
	return nil
}

// DistributeMachineDeploymentWorkers distributes machine deployment for worker nodes
func (c *TkgClient) DistributeMachineDeploymentWorkers(workerMachineCount int64, isProdConfig, isManagementCluster bool, infraProviderName string, isWindowsWorkloadCluster bool) ([]int, error) { // nolint:gocyclo
	workerCounts := make([]int, 3)
	if infraProviderName == DockerProviderName || isWindowsWorkloadCluster {
		workerCounts[0] = int(workerMachineCount)
		return workerCounts, nil
	}
	workerCount1Str, err1 := c.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerMachineCount0)
	workerCount2Str, err2 := c.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerMachineCount1)
	workerCount3Str, err3 := c.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerMachineCount2)
	predefinedDistribution := err1 == nil && err2 == nil && err3 == nil && workerCount1Str != "" && workerCount2Str != "" && workerCount3Str != ""

	if isProdConfig && !predefinedDistribution {
		workersPerAz := workerMachineCount / 3
		remainder := workerMachineCount % 3
		workerCounts[0] = int(workersPerAz)
		workerCounts[1] = int(workersPerAz)
		workerCounts[2] = int(workersPerAz)
		for remainder > 0 {
			for i := range workerCounts {
				if remainder == 0 {
					break
				}
				workerCounts[i]++
				remainder--
			}
		}
	} else if isProdConfig {
		workerCount1, e1 := strconv.Atoi(workerCount1Str)
		workerCount2, e2 := strconv.Atoi(workerCount2Str)
		workerCount3, e3 := strconv.Atoi(workerCount3Str)
		if e1 != nil || e2 != nil || e3 != nil {
			return nil, errors.Errorf("failed to parse provided WORKER_MACHINE_COUNT_0/2/3 vars as integers: %s, %s, %s", workerCount1Str, workerCount2Str, workerCount3Str)
		}
		workerCounts[0] = workerCount1
		workerCounts[1] = workerCount2
		workerCounts[2] = workerCount3
	} else if err1 == nil && workerCount1Str != "" {
		workerCount1, err := strconv.Atoi(workerCount1Str)
		if err != nil {
			return nil, errors.Errorf("failed to parse provided WORKER_MACHINE_COUNT_0 var as integer: %s", workerCount1Str)
		}
		workerCounts[0] = workerCount1
	} else {
		workerCounts[0] = int(workerMachineCount)
	}

	if !isManagementCluster && isProdConfig && workerCounts[0]+workerCounts[1]+workerCounts[2] < 3 {
		return nil, errors.Errorf("prod plan requires at least 3 workers")
	}

	return workerCounts, nil
}

// SetMachineDeploymentWorkerCounts sets machine deployment counts
func (c *TkgClient) SetMachineDeploymentWorkerCounts(workerCounts []int, totalWorkerMachineCount int64, isProdConfig bool) {
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableWorkerMachineCount, strconv.Itoa(int(totalWorkerMachineCount)))
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableWorkerMachineCount0, strconv.Itoa(workerCounts[0]))
	if isProdConfig {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableWorkerMachineCount1, strconv.Itoa(workerCounts[1]))
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableWorkerMachineCount2, strconv.Itoa(workerCounts[2]))
	}
}

// EncodeAWSCredentialsAndGetClient encodes aws credentials and returns aws client
func (c *TkgClient) EncodeAWSCredentialsAndGetClient(clusterClient clusterclient.Client) (aws.Client, error) {
	creds, err := c.GetAWSCreds()
	if err != nil {
		return nil, err
	}
	awsClient, err := aws.New(*creds)
	if err != nil {
		return nil, err
	}

	b64Creds, err := awsClient.EncodeCredentials()
	if err != nil {
		return awsClient, err
	}
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableAWSB64Credentials, b64Creds)

	return awsClient, nil
}

// ConfigureAndValidateHTTPProxyConfiguration configures and validates http proxy configuration
func (c *TkgClient) ConfigureAndValidateHTTPProxyConfiguration(infrastructureName string) error {
	c.SetDefaultProxySettings()
	proxyEnabled, err := c.TKGConfigReaderWriter().Get(constants.TKGHTTPProxyEnabled)
	if err != nil || proxyEnabled != trueString {
		httpProxy, err2 := c.TKGConfigReaderWriter().Get(constants.TKGHTTPProxy)
		if httpProxy == "" || err2 != nil {
			return nil
		}
		// httpProxy and httpsProxy are presents, check if the TKGHTTPProxyEnabled
		return errors.Wrapf(err, "cannot get %s", constants.TKGHTTPProxyEnabled)
	}

	httpProxy, err := c.TKGConfigReaderWriter().Get(constants.TKGHTTPProxy)
	if err != nil || httpProxy == "" {
		return errors.Wrapf(err, "cannot get %s", constants.TKGHTTPProxy)
	}

	if _, err = tkgconfigproviders.CheckAndGetProxyURL("", "", httpProxy); err != nil {
		return errors.Wrapf(err, "error validating %s", constants.TKGHTTPProxy)
	}

	httpsProxy, _ := c.TKGConfigReaderWriter().Get(constants.TKGHTTPSProxy)
	if err != nil || httpsProxy == "" {
		return errors.Wrapf(err, "cannot get %s", constants.TKGHTTPSProxy)
	}
	if _, err = tkgconfigproviders.CheckAndGetProxyURL("", "", httpsProxy); err != nil {
		return errors.Wrapf(err, "error validating %s", constants.TKGHTTPSProxy)
	}

	noProxy, err := c.getFullTKGNoProxy(infrastructureName)
	if err != nil {
		return err
	}

	c.TKGConfigReaderWriter().Set(constants.TKGNoProxy, noProxy)

	return nil
}

// SetDefaultProxySettings is used to configure default proxy settings.
// The TKG proxy variables are required for cloud-api component templates
// rendering. Need to set them to "" if the proxy variables are not available.
func (c *TkgClient) SetDefaultProxySettings() {
	if _, err := c.TKGConfigReaderWriter().Get(constants.TKGHTTPProxy); err != nil {
		c.TKGConfigReaderWriter().Set(constants.TKGHTTPProxy, "")
	}
	if _, err := c.TKGConfigReaderWriter().Get(constants.TKGHTTPSProxy); err != nil {
		c.TKGConfigReaderWriter().Set(constants.TKGHTTPSProxy, "")
	}
	if _, err := c.TKGConfigReaderWriter().Get(constants.TKGNoProxy); err != nil {
		c.TKGConfigReaderWriter().Set(constants.TKGNoProxy, "")
	}
}

func (c *TkgClient) getFullTKGNoProxy(providerName string) (string, error) {
	noProxyMap := make(map[string]bool)

	noProxyMap[constants.ServiceDNSClusterLocalSuffix] = true
	noProxyMap[constants.ServiceDNSSuffix] = true
	noProxyMap[constants.LocalHost] = true
	noProxyMap[constants.LocalHostIP] = true

	if serviceCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableServiceCIDR); serviceCIDR != "" {
		for _, np := range strings.Split(serviceCIDR, ",") {
			noProxyMap[np] = true
		}
	}
	if clusterCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterCIDR); clusterCIDR != "" {
		for _, np := range strings.Split(clusterCIDR, ",") {
			noProxyMap[np] = true
		}
	}
	if ipfamily, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableIPFamily); ipfamily == constants.IPv6Family {
		noProxyMap[constants.LocalHostIPv6] = true
	}

	if noProxy, _ := c.TKGConfigReaderWriter().Get(constants.TKGNoProxy); noProxy != "" {
		// trim space
		replaceSpacePattern := regexp.MustCompile(`\s+|\t+|\n+|\r+`)
		noProxy = replaceSpacePattern.ReplaceAllString(noProxy, "")

		if strings.Contains(noProxy, "*") {
			return "", fmt.Errorf("invalid string '*' in %s", constants.TKGNoProxy)
		}
		for _, np := range strings.Split(noProxy, ",") {
			noProxyMap[np] = true
		}
	}
	// update provider specific no proxies has not been checked into tkg-cli-providers yet
	err := c.updateProviderSpecificNoProxy(providerName, noProxyMap)

	noProxyList := []string{}
	for np := range noProxyMap {
		noProxyList = append(noProxyList, np)
	}
	return strings.Join(noProxyList, ","), err
}

// updateProviderSpecificNoProxy updates provider specific no proxies to given input map
func (c *TkgClient) updateProviderSpecificNoProxy(providerName string, noProxyMap map[string]bool) error {
	switch providerName {
	case constants.InfrastructureProviderAWS:
		if vpcCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSVPCCIDR); vpcCIDR != "" {
			noProxyMap[vpcCIDR] = true
		}
		noProxyMap[constants.LinkLocalAddress] = true
	case constants.InfrastructureProviderAzure:
		if vnetCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureVnetCidr); vnetCIDR != "" {
			for _, np := range strings.Split(vnetCIDR, ",") {
				noProxyMap[np] = true
			}
		}
		noProxyMap[constants.LinkLocalAddress] = true
		noProxyMap[constants.AzurePublicVIP] = true
	case constants.InfrastructureProviderDocker:
		var dockerBridgeCidr string
		var err error
		if dockerBridgeCidr, err = getDockerBridgeNetworkCidr(); err != nil {
			return err
		}
		noProxyMap[dockerBridgeCidr] = true
	}
	return nil
}

func (c *TkgClient) configureVsphereCredentialsFromCluster(clusterClient clusterclient.Client) error {
	regionContext, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "failed to get current region context")
	}
	vsphereUsername, vspherePassword, err := clusterClient.GetVCCredentialsFromCluster(regionContext.ClusterName, constants.TkgNamespace)
	if err != nil {
		return errors.Wrap(err, "unable to get vsphere credentials from secret")
	}
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereUsername, vsphereUsername)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableVspherePassword, vspherePassword)
	return nil
}

// CheckClusterNameFormat ensures that the cluster name is valid for the given provider
func CheckClusterNameFormat(clusterName, infrastructureProvider string) error {
	var clusterNameRegex string
	if infrastructureProvider == AzureProviderName {
		// Azure limitation
		clusterNameRegex = "^[a-z][a-z0-9-.]{0,42}[a-z0-9]$"
	} else {
		// k8s resource name DNS limitation
		clusterNameRegex = "^[a-z0-9][a-z0-9-.]{0,61}[a-z0-9]$"
	}
	matched, err := regexp.MatchString(clusterNameRegex, clusterName)
	if err != nil {
		return errors.Wrap(err, "failed to validate cluster name")
	}
	if !matched {
		return errors.Errorf("cluster name doesn't match regex %s, can contain only lowercase alphanumeric characters, '.' and '-'", clusterNameRegex)
	}

	return nil
}

func (c *TkgClient) configureAndValidateIPFamilyConfiguration(clusterRole string) error {
	// ignoring error because IPFamily is an optional configuration
	// if not set Get will return an empty string
	ipFamily, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableIPFamily)
	if ipFamily == "" {
		ipFamily = constants.IPv4Family
	}

	err := c.checkIPFamilyFeatureFlags(ipFamily, clusterRole)
	if err != nil {
		return err
	}

	serviceCIDRs, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableServiceCIDR)
	clusterCIDRs, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterCIDR)

	if serviceCIDRs == "" {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableServiceCIDR, c.defaultServiceCIDR(ipFamily))
	} else if err := c.validateCIDRsForIPFamily(constants.ConfigVariableServiceCIDR, serviceCIDRs, ipFamily); err != nil {
		return err
	}
	if clusterCIDRs == "" {
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterCIDR, c.defaultClusterCIDR(ipFamily))
	} else if err := c.validateCIDRsForIPFamily(constants.ConfigVariableClusterCIDR, clusterCIDRs, ipFamily); err != nil {
		return err
	}
	if err := c.validateIPHostnameForIPFamily(constants.TKGHTTPProxy, ipFamily); err != nil {
		return err
	}
	if err := c.validateIPHostnameForIPFamily(constants.TKGHTTPSProxy, ipFamily); err != nil {
		return err
	}
	return nil
}

func (c *TkgClient) configureAndValidateCoreDNSIP() error {
	// Core DNS IP is the 10th index of service CIDR subnet
	// ServiceCIDR must not be empty as it should already been set by configureAndValidateIPFamilyConfiguration if it was omitted
	serviceCIDR, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableServiceCIDR)
	if err != nil {
		return err
	}

	svcSubnets, err := netutils.ParseCIDRs(strings.Split(serviceCIDR, ","))
	if err != nil {
		return err
	}
	dnsIP, err := netutils.GetIndexedIP(svcSubnets[0], 10)
	if err != nil {
		return err
	}

	c.TKGConfigReaderWriter().Set(constants.ConfigVariableCoreDNSIP, dnsIP.String())

	return nil
}

func (c *TkgClient) validateServiceCIDRNetmask() error {
	// kube-apiserver requires that the service CIDR be of limited size.
	// This validation avoids a case where the cluster never comes up when
	// the CIDR is too large.
	// https://github.com/kubernetes/kubernetes/blob/3c87c43ceff6122637c8d8070601f7271026360e/cmd/kube-apiserver/app/options/validation.go#L52
	configVariableName := constants.ConfigVariableServiceCIDR
	serviceCIDRs, _ := c.TKGConfigReaderWriter().Get(configVariableName)
	cidrSlice := strings.Split(serviceCIDRs, ",")
	for _, cidrString := range cidrSlice {
		_, cidr, err := net.ParseCIDR(cidrString)
		if err != nil {
			// This should never happen since CIDRs were already validated before
			return errors.Errorf("invalid %s \"%s\"", configVariableName, serviceCIDRs)
		}
		maxCIDRBits := 20
		var ones, bits = cidr.Mask.Size()
		if bits-ones > maxCIDRBits {
			return errors.Errorf("invalid %s \"%s\", expected netmask to be \"/%d\" or greater", configVariableName, cidrString, bits-maxCIDRBits)
		}
	}
	return nil
}

func (c *TkgClient) defaultClusterCIDR(ipFamily string) string {
	switch ipFamily {
	case constants.DualStackPrimaryIPv4Family:
		return constants.DefaultDualStackPrimaryIPv4ClusterCIDR
	case constants.DualStackPrimaryIPv6Family:
		return constants.DefaultDualStackPrimaryIPv6ClusterCIDR
	case constants.IPv6Family:
		return constants.DefaultIPv6ClusterCIDR
	default:
		return constants.DefaultIPv4ClusterCIDR
	}
}

func (c *TkgClient) defaultServiceCIDR(ipFamily string) string {
	switch ipFamily {
	case constants.DualStackPrimaryIPv4Family:
		return constants.DefaultDualStackPrimaryIPv4ServiceCIDR
	case constants.DualStackPrimaryIPv6Family:
		return constants.DefaultDualStackPrimaryIPv6ServiceCIDR
	case constants.IPv6Family:
		return constants.DefaultIPv6ServiceCIDR
	default:
		return constants.DefaultIPv4ServiceCIDR
	}
}

func (c *TkgClient) validateIPHostnameForIPFamily(configKey, ipFamily string) error {
	urlString, err := c.TKGConfigReaderWriter().Get(configKey)
	if err != nil {
		return nil
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil
	}

	ip := net.ParseIP(parsedURL.Hostname())
	if ip == nil {
		return nil
	}

	switch ipFamily {
	case constants.DualStackPrimaryIPv4Family, constants.DualStackPrimaryIPv6Family:
		return nil
	case constants.IPv6Family:
		if ip.To4() == nil {
			return nil
		}
	case constants.IPv4Family:
		if ip.To4() != nil {
			return nil
		}
	}

	return errors.Errorf("invalid %s \"%s\", expected to be an address of type \"%s\" (%s)",
		configKey, urlString, ipFamily, constants.ConfigVariableIPFamily)
}

func (c *TkgClient) validateCIDRsForIPFamily(configVariableName, cidrs, ipFamily string) error {
	switch ipFamily {
	case constants.IPv4Family:
		if !isCIDRIPv4(cidrs) {
			return invalidCIDRError(configVariableName, cidrs, ipFamily)
		}
	case constants.IPv6Family:
		if !isCIDRIPv6(cidrs) {
			return invalidCIDRError(configVariableName, cidrs, ipFamily)
		}
	case constants.DualStackPrimaryIPv4Family:
		cidrSlice := strings.Split(cidrs, ",")
		if len(cidrSlice) != 2 || !isCIDRIPv4(cidrSlice[0]) || !isCIDRIPv6(cidrSlice[1]) {
			return fmt.Errorf(`invalid %s %q, expected to have "<IPv4 CIDR>,<IPv6 CIDR>" for %s %q`,
				configVariableName, cidrs, constants.ConfigVariableIPFamily, ipFamily)
		}
	case constants.DualStackPrimaryIPv6Family:
		cidrSlice := strings.Split(cidrs, ",")
		if len(cidrSlice) != 2 || !isCIDRIPv6(cidrSlice[0]) || !isCIDRIPv4(cidrSlice[1]) {
			return fmt.Errorf(`invalid %s %q, expected to have "<IPv6 CIDR>,<IPv4 CIDR>" for %s %q`,
				configVariableName, cidrs, constants.ConfigVariableIPFamily, ipFamily)
		}
	}
	return nil
}

func isCIDRIPv4(cidr string) bool {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return ip.To4() != nil
}

func isCIDRIPv6(cidr string) bool {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return ip.To4() == nil
}

func invalidCIDRError(configKey, cidr, ipFamily string) error {
	return errors.Errorf("invalid %s \"%s\", expected to be a CIDR of type \"%s\" (%s)",
		configKey, cidr, ipFamily, constants.ConfigVariableIPFamily)
}

func getDockerBridgeNetworkCidr() (string, error) {
	var stdout bytes.Buffer
	var networkCidr string

	cmd := exec.Command("docker", "inspect", "-f", "'{{range .IPAM.Config}}{{.Subnet}}{{end}}'", "bridge")
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "failed to fetch the Subnet CIDR for docker 'bridge' network")
	}

	networkCidr = stdout.String()
	networkCidr = strings.TrimSpace(networkCidr)
	networkCidr = strings.Trim(networkCidr, "'")

	return networkCidr, nil
}

// ConfigureAndValidateAviConfiguration validates the configuration inputs of Avi aka. NSX Advanced Load Balancer
func (c *TkgClient) ConfigureAndValidateAviConfiguration() error {
	aviEnable, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviEnable)
	// ignoring error because AVI_ENABLE is an optional configuration
	if aviEnable == "" || aviEnable == "false" {
		return nil
	}
	// init avi client
	aviClient := avi.New()
	// if avi is enabled, then should verify following required fields
	err := c.ValidateAviControllerAccount(aviClient)
	if err != nil {
		return customAviConfigurationError(err, "avi controller account")
	}
	if err = c.ValidateAviCloud(aviClient); err != nil {
		return customAviConfigurationError(err, constants.ConfigVariableAviCloudName)
	}
	if err = c.ValidateAviServiceEngineGroup(aviClient); err != nil {
		return customAviConfigurationError(err, constants.ConfigVariableAviServiceEngineGroup)
	}
	if err = c.ValidateAviDataPlaneNetwork(aviClient); err != nil {
		return customAviConfigurationError(err, fmt.Sprintf("<%s,%s>", constants.ConfigVariableAviDataPlaneNetworkName, constants.ConfigVariableAviDataPlaneNetworkCIDR))
	}

	// validate following optional fields if configured
	if err = c.ValidateAviManagementClusterServiceEngineGroup(aviClient); err != nil {
		return customAviConfigurationError(err, constants.ConfigVariableAviManagementClusterServiceEngineGroup)
	}
	if err = c.ValidateAviControlPlaneNetwork(aviClient); err != nil {
		return customAviConfigurationError(err, fmt.Sprintf("<%s,%s>", constants.ConfigVariableAviControlPlaneNetworkName, constants.ConfigVariableAviControlPlaneNetworkCIDR))
	}
	if err = c.ValidateAviManagementClusterDataPlaneNetwork(aviClient); err != nil {
		return customAviConfigurationError(err, fmt.Sprintf("<%s,%s>", constants.ConfigVariableAviManagementClusterDataPlaneNetworkName, constants.ConfigVariableAviManagementClusterDataPlaneNetworkCIDR))
	}
	if err = c.ValidateAviManagementClusterControlPlaneNetwork(aviClient); err != nil {
		return customAviConfigurationError(err, fmt.Sprintf("<%s,%s>", constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkName, constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkCIDR))
	}
	return nil
}

// ValidateAviControllerAccount validates if provide avi credentials are able to connect to avi controller or not
func (c *TkgClient) ValidateAviControllerAccount(aviClient avi.Client) error {
	aviController, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviControllerAddress)
	if err != nil {
		return err
	}
	aviUserName, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviControllerUsername)
	if err != nil {
		return err
	}
	aviPassword, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviControllerPassword)
	if err != nil {
		return err
	}
	aviCAEncoded, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviControllerCA)
	if err != nil {
		return err
	}
	aviCA, err := base64.StdEncoding.DecodeString(aviCAEncoded)
	if err != nil {
		return err
	}
	// validate avi controller account
	aviControllerParams := &models.AviControllerParams{
		Username: aviUserName,
		Password: aviPassword,
		Host:     aviController,
		Tenant:   "admin",
		CAData:   string(aviCA),
	}
	authed, err := aviClient.VerifyAccount(aviControllerParams)
	if err != nil {
		return err
	}
	if !authed {
		return errors.Errorf("unable to authenticate avi controller due to incorrect credentials")
	}
	return nil
}

// ValidateAviCloud validates if configured cloud exists or not
func (c *TkgClient) ValidateAviCloud(aviClient avi.Client) error {
	aviCloud, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviCloudName)
	if err != nil {
		return err
	}
	if _, err = aviClient.GetCloudByName(aviCloud); err != nil {
		return err
	}
	return nil
}

// ValidateAviServiceEngineGroup validates if configured service engine group exists or not
func (c *TkgClient) ValidateAviServiceEngineGroup(aviClient avi.Client) error {
	aviServiceEngineGroup, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviServiceEngineGroup)
	if err != nil {
		return err
	}
	if _, err = aviClient.GetServiceEngineGroupByName(aviServiceEngineGroup); err != nil && !isAviResourceDuplicatedNameError(err) {
		return err
	}
	return nil
}

// ValidateAviManagementClusterServiceEngineGroup validates if configured management cluster service engine group exists or not
func (c *TkgClient) ValidateAviManagementClusterServiceEngineGroup(aviClient avi.Client) error {
	aviManagementClusterSEG, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviManagementClusterServiceEngineGroup)
	// this field is optional, only validates if it has value
	if aviManagementClusterSEG != "" {
		if _, err := aviClient.GetServiceEngineGroupByName(aviManagementClusterSEG); err != nil && !isAviResourceDuplicatedNameError(err) {
			return err
		}
	}
	return nil
}

// ValidateAviDataPlaneNetwork validates if workload clusters' data plane vip network is valid or not
func (c *TkgClient) ValidateAviDataPlaneNetwork(aviClient avi.Client) error {
	aviDataNetworkName, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviDataPlaneNetworkName)
	if err != nil {
		return err
	}
	aviDataNetworkCIDR, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviDataPlaneNetworkCIDR)
	if err != nil {
		return err
	}
	if err := c.ValidateAviNetwork(aviDataNetworkName, aviDataNetworkCIDR, aviClient); err != nil {
		return err
	}
	return nil
}

// ValidateAviControlPlaneNetwork validates if workload clusters' control plane vip network is valid or not
func (c *TkgClient) ValidateAviControlPlaneNetwork(aviClient avi.Client) error {
	aviControlPlaneNetworkName, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviControlPlaneNetworkName)
	aviControlPlaneNetworkCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviControlPlaneNetworkCIDR)
	// these field are optional, only validates if they have value
	if aviControlPlaneNetworkName != "" {
		if err := c.ValidateAviNetwork(aviControlPlaneNetworkName, aviControlPlaneNetworkCIDR, aviClient); err != nil {
			return err
		}
	}
	return nil
}

// ValidateAviManagementClusterDataPlaneNetwork checks if configured management cluster data plane vip network is valid or not
func (c *TkgClient) ValidateAviManagementClusterDataPlaneNetwork(aviClient avi.Client) error {
	aviManagementClusterDataPlaneNetworkName, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviManagementClusterDataPlaneNetworkName)
	aviManagementClusterDataPlaneNetworkCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviManagementClusterDataPlaneNetworkCIDR)
	// these field are optional, only validates if they have value
	if aviManagementClusterDataPlaneNetworkName != "" {
		if err := c.ValidateAviNetwork(aviManagementClusterDataPlaneNetworkName, aviManagementClusterDataPlaneNetworkCIDR, aviClient); err != nil {
			return err
		}
	}
	return nil
}

// ValidateAviManagementClusterControlPlaneNetwork checks if configured management cluster control plane vip network is valid or not
func (c *TkgClient) ValidateAviManagementClusterControlPlaneNetwork(aviClient avi.Client) error {
	aviManagementClusterControlPlaneNetworkName, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkName)
	aviManagementClusterControlPlaneNetworkCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkCIDR)
	// these field are optional, only validates if they have value
	if aviManagementClusterControlPlaneNetworkName != "" {
		if err := c.ValidateAviNetwork(aviManagementClusterControlPlaneNetworkName, aviManagementClusterControlPlaneNetworkCIDR, aviClient); err != nil {
			return err
		}
	}
	return nil
}

// ValidateAviNetwork validates if the network can be found in AVI controller or not and the subnet CIDR format is correct or not
func (c *TkgClient) ValidateAviNetwork(networkName, networkCIDR string, aviClient avi.Client) error {
	_, err := aviClient.GetVipNetworkByName(networkName)
	if err != nil && !isAviResourceDuplicatedNameError(err) {
		return err
	}
	_, _, err = net.ParseCIDR(networkCIDR)
	if err != nil {
		return err
	}
	return nil
}

// ConfigureAndValidateNameserverConfiguration validates the configuration of the control plane node and workload node nameservers
func (c *TkgClient) ConfigureAndValidateNameserverConfiguration(clusterRole string) error {
	err := c.validateNameservers(constants.ConfigVariableControlPlaneNodeNameservers, clusterRole)
	if err != nil {
		return err
	}

	return c.validateNameservers(constants.ConfigVariableWorkerNodeNameservers, clusterRole)
}

func (c *TkgClient) validateNameservers(nameserverConfigVariable, clusterRole string) error {
	// ignoring error because IPFamily is an optional configuration
	// if not set Get will return an empty string
	ipFamily, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableIPFamily)
	if ipFamily == "" {
		ipFamily = constants.IPv4Family
	}

	nameservers, err := c.TKGConfigReaderWriter().Get(nameserverConfigVariable)
	if err != nil {
		return nil
	}

	if clusterRole == TkgLabelClusterRoleManagement && !c.IsFeatureActivated(constants.FeatureFlagManagementClusterCustomNameservers) {
		return customNameserverFeatureFlagError(nameserverConfigVariable, nameservers, constants.FeatureFlagManagementClusterCustomNameservers)
	} else if clusterRole == TkgLabelClusterRoleWorkload && !c.IsFeatureActivated(constants.FeatureFlagClusterCustomNameservers) {
		return customNameserverFeatureFlagError(nameserverConfigVariable, nameservers, constants.FeatureFlagClusterCustomNameservers)
	}

	invalidNameservers := []string{}
	for _, nameserver := range strings.Split(nameservers, ",") {
		nameserver = strings.TrimSpace(nameserver)
		ip := net.ParseIP(nameserver)
		if ip == nil ||
			ipFamily == constants.IPv4Family && ip.To4() == nil ||
			ipFamily == constants.IPv6Family && ip.To4() != nil {
			invalidNameservers = append(invalidNameservers, nameserver)
		}
	}

	if len(invalidNameservers) > 0 {
		return fmt.Errorf("invalid %s %q, expected to be IP addresses that match TKG_IP_FAMILY %q", nameserverConfigVariable, strings.Join(invalidNameservers, ","), ipFamily)
	}
	return nil
}

func (c *TkgClient) checkIPFamilyFeatureFlags(ipFamily, clusterRole string) error {
	if clusterRole == TkgLabelClusterRoleManagement {
		dualIPv4PrimaryEnabled := c.IsFeatureActivated(constants.FeatureFlagManagementClusterDualStackIPv4Primary)
		if !dualIPv4PrimaryEnabled && ipFamily == constants.DualStackPrimaryIPv4Family {
			return dualStackFeatureFlagError(ipFamily, constants.FeatureFlagManagementClusterDualStackIPv4Primary)
		}
		dualIPv6PrimaryEnabled := c.IsFeatureActivated(constants.FeatureFlagManagementClusterDualStackIPv6Primary)
		if !dualIPv6PrimaryEnabled && ipFamily == constants.DualStackPrimaryIPv6Family {
			return dualStackFeatureFlagError(ipFamily, constants.FeatureFlagManagementClusterDualStackIPv6Primary)
		}
	} else {
		dualIPv4PrimaryEnabled := c.IsFeatureActivated(constants.FeatureFlagClusterDualStackIPv4Primary)
		if !dualIPv4PrimaryEnabled && ipFamily == constants.DualStackPrimaryIPv4Family {
			return dualStackFeatureFlagError(ipFamily, constants.FeatureFlagClusterDualStackIPv4Primary)
		}
		dualIPv6PrimaryEnabled := c.IsFeatureActivated(constants.FeatureFlagClusterDualStackIPv6Primary)
		if !dualIPv6PrimaryEnabled && ipFamily == constants.DualStackPrimaryIPv6Family {
			return dualStackFeatureFlagError(ipFamily, constants.FeatureFlagClusterDualStackIPv6Primary)
		}
	}

	return nil
}

func isAviResourceDuplicatedNameError(err error) bool {
	return strings.Contains(err.Error(), "More than one object of type ")
}

func dualStackFeatureFlagError(ipFamily, featureFlag string) error {
	return fmt.Errorf("option TKG_IP_FAMILY is set to %q, but dualstack support is not enabled (because it is under development). To enable dualstack, set %s to \"true\"", ipFamily, featureFlag)
}

func customNameserverFeatureFlagError(configVariable, nameservers, flagName string) error {
	return fmt.Errorf("option %s is set to %q, but custom nameserver support is not enabled (because it is not fully functional). To enable custom nameservers, run the command: tanzu config set %s true",
		configVariable,
		nameservers,
		flagName)
}

func customAviConfigurationError(err error, configVariable string) error {
	return errors.Wrapf(err, "nsx advanced load balancer configuration validation error, failed to validate %s", configVariable)
}

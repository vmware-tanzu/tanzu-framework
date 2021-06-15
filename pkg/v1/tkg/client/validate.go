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

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/aws"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/azure"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfighelper"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/vc"
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

const (
	// de-facto defaults initially chosen by kops: https://github.com/kubernetes/kops
	defaultIPv4ClusterCIDR = "100.96.0.0/11"
	defaultIPv4ServiceCIDR = "100.64.0.0/13"

	// chosen to match our IPv4 defaults
	// use /48 for cluster CIDR because each node gets a /64 by default in IPv6
	defaultIPv6ClusterCIDR = "fd00:100:96::/48"
	// use /108 is the max allowed for IPv6
	defaultIPv6ServiceCIDR = "fd00:100:64::/108"
)

var trueString = "true"

// VsphereResourceType vsphere resource types
var VsphereResourceType = []string{constants.ConfigVariableVsphereResourcePool, constants.ConfigVariableVsphereDatastore, constants.ConfigVariableVsphereFolder}

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
	log.V(1).Infof("Downloading bom for TKR %q", tkrName)

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

	tkrConfigMap := &corev1.ConfigMap{}
	if err := regionalClusterClient.GetResource(tkrConfigMap, tkrName, constants.TkrNamespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			return errors.Errorf("ConfigMap for TKR name %q not available to download bom", tkrName)
		}

		return err
	}

	bomData := tkrConfigMap.BinaryData["bomContent"]

	bomDir, err := c.tkgConfigPathsClient.GetTKGBoMDirectory()
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(bomDir, "tkr-bom-"+tkrName+".yaml"), bomData, 0o600); err != nil {
		return err
	}
	return nil
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
			return "", "", errors.Wrap(err, "unable to get default TKR bom")
		}
		tkrVersion = tkrBoMConfig.Release.Version
		k8sVersion, err = tkgconfigbom.GetK8sVersionFromTkrBoM(tkrBoMConfig)
		if err != nil {
			return "", "", errors.Wrap(err, "unable to get default k8s version from TKR bom")
		}
	} else {
		// BoM downloading should only be required if user are passing tkrName,
		// otherwise we should use default config which is always present on user's machine

		// download bom if not present locally for given TKR
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
		return errors.Errorf("unable to find the azure vm image info for TKR version: '%v' and os options: '(%v,%v,%v)'", tkrVersion, osInfo.Name, osInfo.Version, osInfo.Arch)
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

	return errors.Errorf("invalid azure image info: %v, for TKR version: %v", *azureVMImage, tkrVersion)
}

// ConfigureAndValidateAzureConfig configures and validates azure configurationn
func (c *TkgClient) ConfigureAndValidateAzureConfig(tkrVersion string, nodeSizes NodeSizeOptions, skipValidation, isProdConfig bool, workerMachineCount int64, clusterClient clusterclient.Client, isManagementCluster bool) error {
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
		return errors.New("TKR version is empty")
	}

	if err := c.ConfigureAzureVMImage(tkrVersion); err != nil {
		return err
	}

	workerCounts, err := c.DistributeMachineDeploymentWorkers(workerMachineCount, isProdConfig, isManagementCluster, "azure")
	if err != nil {
		return errors.Wrapf(err, "failed to distribute machine deployments")
	}
	c.SetMachineDeploymentWorkerCounts(workerCounts, workerMachineCount, isProdConfig)
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
		return errors.New("TKR version is empty")
	}

	bomConfiguration, err := c.tkgBomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to get bom configuration for TKR version %s", tkrVersion)
	}

	kindNodeImage := bomConfiguration.Components["kubernetes-sigs_kind"][0].Images["kindNodeImage"]
	dockerTemplateImage := tkgconfigbom.GetFullImagePath(kindNodeImage, bomConfiguration.ImageConfig.ImageRepository) + ":" + kindNodeImage.Tag
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableDockerMachineTemplateImage, dockerTemplateImage)
	return nil
}

// ConfigureAndValidateAwsConfig configures and validates aws configuration
func (c *TkgClient) ConfigureAndValidateAwsConfig(tkrVersion string, skipValidation, isProdConfig bool, workerMachineCount int64, isManagementCluster, useExistingVPC bool) error {
	if tkrVersion == "" {
		return errors.New("TKR version is empty")
	}

	bomConfiguration, err := c.tkgBomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to get bom configuration for TKR version %s", tkrVersion)
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
	workerCounts, err := c.DistributeMachineDeploymentWorkers(workerMachineCount, isProdConfig, isManagementCluster, "aws")
	if err != nil {
		return errors.Wrapf(err, "failed to distribute machine deployments")
	}
	c.SetMachineDeploymentWorkerCounts(workerCounts, workerMachineCount, isProdConfig)

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

func checkIfRequiredPermissionsPresent(awsClient aws.Client) error {
	stacks, err := awsClient.ListCloudFormationStacks()
	if err != nil {
		log.Warningf("unable to verify if the AWS CloudFormation stack %s is available in the AWS account.", aws.DefaultCloudFormationStackName)
		return nil
	}

	for _, stack := range stacks {
		if stack == aws.DefaultCloudFormationStackName {
			return nil
		}
	}
	// TODO: should have check on whether IAM permissions are present
	log.Warningf("cannot find AWS CloudFormation stack %s, which is used in the management of IAM groups and policies required by TKG.", aws.DefaultCloudFormationStackName)
	return nil
}

// ConfigureAndValidateAWSConfig configures and validates aws configuration
func (c *TkgClient) ConfigureAndValidateAWSConfig(tkrVersion string, nodeSizes NodeSizeOptions, skipValidation, isProdConfig bool, workerMachineCount int64, clusterClient clusterclient.Client, isManagementCluster bool) error {
	c.SetProviderType(AWSProviderName)

	awsClient, err := c.EncodeAWSCredentialsAndGetClient(clusterClient)
	if err != nil {
		return errors.Wrap(err, "failed to get AWS client")
	}

	if !skipValidation {
		if err := checkIfRequiredPermissionsPresent(awsClient); err != nil {
			return err
		}
	}

	if err := c.OverrideAWSNodeSizeWithOptions(nodeSizes, awsClient, skipValidation); err != nil {
		return errors.Wrap(err, "cannot set AWS node size")
	}

	useExistingVPC, err := c.SetAndValidateDefaultAWSVPCConfiguration(isProdConfig, awsClient, skipValidation)
	if err != nil {
		return errors.Wrap(err, "failed to validate VPC configuration variables")
	}

	return c.ConfigureAndValidateAwsConfig(tkrVersion, skipValidation, isProdConfig, workerMachineCount, isManagementCluster, useExistingVPC)
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
		return errors.New("TKR version is empty")
	}

	templateName, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereTemplate)

	tkrBom, err := c.tkgBomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
	if err != nil {
		return err
	}

	vsphereVM, err := vcClient.GetAndValidateVirtualMachineTemplate(tkrBom.GetOVAVersions(), tkrVersion, templateName, dc, c.TKGConfigReaderWriter())
	if err != nil || vsphereVM == nil {
		return errors.Wrapf(err, "unable to get or validate %s for given TanzuKubernetesRelease", constants.ConfigVariableVsphereTemplate)
	}

	c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereTemplate, vsphereVM.Name)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSName, vsphereVM.DistroName)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSVersion, vsphereVM.DistroVersion)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableOSArch, vsphereVM.DistroArch)
	return nil
}

// GetVSphereEndpoint gets vsphere client based on credentials set in config variables
func (c *TkgClient) GetVSphereEndpoint(clusterClient clusterclient.Client) (vc.Client, error) {
	if clusterClient != nil {
		username, password, err := clusterClient.GetVCCredentialsFromSecret()
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

		return vc.GetAuthenticatedVCClient(server, username, password, vsphereThumbprint, vsphereInsecure)
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
	vcClient, err := vc.NewClient(vcURL, thumbprint, vsphereInsecure)
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
func (c *TkgClient) ConfigureAndValidateManagementClusterConfiguration(options *InitRegionOptions, skipValidation bool) *ValidationError { // nolint:gocyclo
	var err error
	if options.ClusterName != "" {
		if err := checkClusterNameFormat(options.ClusterName); err != nil {
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

	if err = c.ConfigureAndValidateHTTPProxyConfiguration(name); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if err = c.configureAndValidateIPFamilyConfiguration(); err != nil {
		return NewValidationError(ValidationErrorCode, err.Error())
	}

	if name == AWSProviderName {
		if err := c.ConfigureAndValidateAWSConfig(tkrVersion, options.NodeSizeOptions, skipValidation, options.Plan == constants.PlanProd, constants.DefaultWorkerMachineCountForManagementCluster, nil, true); err != nil {
			return NewValidationError(ValidationErrorCode, err.Error())
		}
	}

	if name == VSphereProviderName {
		if err := c.ConfigureAndValidateVsphereConfig(tkrVersion, options.NodeSizeOptions, options.VsphereControlPlaneEndpoint, skipValidation, nil); err != nil {
			return err
		}
	}

	if name == AzureProviderName {
		if err := c.ConfigureAndValidateAzureConfig(tkrVersion, options.NodeSizeOptions, skipValidation, options.Plan == constants.PlanProd, constants.DefaultWorkerMachineCountForManagementCluster, nil, true); err != nil {
			return NewValidationError(ValidationErrorCode, err.Error())
		}
	}

	if name == DockerProviderName {
		if err := c.ConfigureAndValidateDockerConfig(tkrVersion, options.NodeSizeOptions, skipValidation); err != nil {
			return NewValidationError(ValidationErrorCode, err.Error())
		}
	}

	return nil
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
	for _, resourceType := range VsphereResourceType {
		path, err := c.TKGConfigReaderWriter().Get(resourceType)
		if err != nil {
			return nil
		}

		switch resourceType {
		case constants.ConfigVariableVsphereResourcePool:
			_, err := vcClient.FindResourcePool(context.Background(), path, dcPath)
			if err != nil {
				return errors.Wrapf(err, "invalid %s", resourceType)
			}
		case constants.ConfigVariableVsphereDatastore:
			_, err := vcClient.FindDatastore(context.Background(), path, dcPath)
			if err != nil {
				return errors.Wrapf(err, "invalid %s", resourceType)
			}
		case constants.ConfigVariableVsphereFolder:
			_, err := vcClient.FindFolder(context.Background(), path, dcPath)
			if err != nil {
				return errors.Wrapf(err, "invalid %s", resourceType)
			}

		default:
			return errors.Errorf("unknown vsphere resource type %s", resourceType)
		}
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
	region, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSRegion)
	if err != nil {
		return nil
	}

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
		nodeTypes, err := awsClient.ListInstanceTypes()
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
			return errors.Errorf("instance type %s is not supported in region %s", controlplaneMachineType, region)
		}

		nodeMachineType, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableNodeMachineType)
		if err != nil {
			return err
		}
		if _, ok := nodeMap[nodeMachineType]; !ok {
			return errors.Errorf("instance type %s is not supported in region %s", nodeMachineType, region)
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
func (c *TkgClient) DistributeMachineDeploymentWorkers(workerMachineCount int64, isProdConfig, isManagementCluster bool, infraProviderName string) ([]int, error) { // nolint:gocyclo
	workerCounts := make([]int, 3)
	if infraProviderName != "aws" && infraProviderName != "azure" {
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

func (c *TkgClient) getAWSCredentialsFromSecret(clusterClient clusterclient.Client) (aws.Client, error) {
	if clusterClient == nil {
		return nil, errors.New("cluster client is not initialized")
	}
	creds, err := clusterClient.GetAWSCredentialsFromSecret()
	if err != nil {
		return nil, err
	}

	awsClient, err := aws.NewFromEncodedCrendentials(creds)
	if err != nil {
		return nil, err
	}
	encodedCreds, err := awsClient.EncodeCredentials()
	if err != nil {
		return nil, err
	}
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableAWSB64Credentials, encodedCreds)
	return awsClient, nil
}

// EncodeAWSCredentialsAndGetClient encodes aws credentials and returns aws client
func (c *TkgClient) EncodeAWSCredentialsAndGetClient(clusterClient clusterclient.Client) (aws.Client, error) {
	if awsClient, err := c.getAWSCredentialsFromSecret(clusterClient); err == nil {
		return awsClient, nil
	}

	log.Warning("unable to get credentials from secret. Trying to get the AWS credentials from configuration file or default credentials provider chain")

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
		return nil
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
		noProxyMap[serviceCIDR] = true
	}
	if clusterCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterCIDR); clusterCIDR != "" {
		noProxyMap[clusterCIDR] = true
	}
	if ipfamily, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableIPFamily); ipfamily == constants.IPv6Family {
		noProxyMap[constants.LocalHostIPv6] = true
	}

	if noProxy, _ := c.TKGConfigReaderWriter().Get(constants.TKGNoProxy); noProxy != "" {
		for _, np := range strings.Split(noProxy, ",") {
			noProxyMap[np] = true
		}
	}
	// below provider specific no proxies has not been checked into tkg-cli-providers yet
	switch providerName {
	case constants.InfrastructureProviderAWS:
		if vpcCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSVPCCIDR); vpcCIDR != "" {
			noProxyMap[vpcCIDR] = true
		}
		noProxyMap[constants.LinkLocalAddress] = true
	case constants.InfrastructureProviderAzure:
		if vnetCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAzureVnetCidr); vnetCIDR != "" {
			noProxyMap[vnetCIDR] = true
		}
		noProxyMap[constants.LinkLocalAddress] = true
		noProxyMap[constants.AzurePublicVIP] = true
	case constants.InfrastructureProviderDocker:
		var dockerBridgeCidr string
		var err error
		if dockerBridgeCidr, err = getDockerBridgeNetworkCidr(); err != nil {
			return "", err
		}
		noProxyMap[dockerBridgeCidr] = true
	}

	noProxyList := []string{}

	for np := range noProxyMap {
		noProxyList = append(noProxyList, np)
	}

	return strings.Join(noProxyList, ","), nil
}

func (c *TkgClient) configureVsphereCredentialsFromCluster(clusterClient clusterclient.Client) error {
	vsphereUsername, vspherePassword, err := clusterClient.GetVCCredentialsFromSecret()
	if err != nil {
		return errors.Wrap(err, "unable to get vsphere credentials from secret")
	}
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereUsername, vsphereUsername)
	c.TKGConfigReaderWriter().Set(constants.ConfigVariableVspherePassword, vspherePassword)
	return nil
}

func checkClusterNameFormat(clusterName string) error {
	matched, err := regexp.MatchString("^[a-z][a-z0-9-.]{0,44}[a-z0-9]$", clusterName)
	if err != nil {
		return errors.Wrap(err, "failed to validate cluster name")
	}
	if !matched {
		return errors.New("cluster name doesn't match regex ^[a-z][a-z0-9-.]{0,44}[a-z0-9]$, can contain only lowercase alphanumeric characters, '.' and '-', must start/end with an alphanumeric character")
	}
	return nil
}

func (c *TkgClient) configureAndValidateIPFamilyConfiguration() error {
	// ignoring error because IPFamily is an optional configuration
	// if not set Get will return an empty string
	ipFamily, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableIPFamily)

	serviceCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableServiceCIDR)
	clusterCIDR, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterCIDR)

	if ipFamily == constants.IPv6Family {
		if serviceCIDR == "" {
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableServiceCIDR, defaultIPv6ServiceCIDR)
		} else if !c.validateIPv6CIDR(serviceCIDR) {
			return invalidCIDRError(constants.ConfigVariableServiceCIDR, serviceCIDR, ipFamily)
		}
		if clusterCIDR == "" {
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterCIDR, defaultIPv6ClusterCIDR)
		} else if !c.validateIPv6CIDR(clusterCIDR) {
			return invalidCIDRError(constants.ConfigVariableClusterCIDR, clusterCIDR, ipFamily)
		}
		if err := c.validateIPHostnameIsIPv6(constants.TKGHTTPProxy); err != nil {
			return err
		}
		if err := c.validateIPHostnameIsIPv6(constants.TKGHTTPSProxy); err != nil {
			return err
		}
	} else { // For cases when TKG_IP_FAMILY is empty or ipv4
		if serviceCIDR == "" {
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableServiceCIDR, defaultIPv4ServiceCIDR)
		}
		if clusterCIDR == "" {
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterCIDR, defaultIPv4ClusterCIDR)
		}
	}
	return nil
}

func (c *TkgClient) validateIPHostnameIsIPv6(configKey string) error {
	urlString, err := c.TKGConfigReaderWriter().Get(configKey)
	if err != nil {
		return nil
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil
	}

	ip := net.ParseIP(parsedURL.Host)
	if ip == nil {
		return nil
	}

	if ip.To4() != nil {
		return errors.Errorf("invalid %s \"%s\", expected to be an address of type \"ipv6\" (%s)",
			configKey, urlString, constants.ConfigVariableIPFamily)
	}

	return nil
}

func (c *TkgClient) validateIPv6CIDR(cidr string) bool {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	if ip.To4() != nil {
		return false
	}
	return true
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

// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	cfn "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	utilpointer "k8s.io/utils/pointer"
	"sigs.k8s.io/cluster-api-provider-aws/v2/cmd/clusterawsadm/cloudformation/bootstrap"
	cloudformation "sigs.k8s.io/cluster-api-provider-aws/v2/cmd/clusterawsadm/cloudformation/service"
	"sigs.k8s.io/cluster-api-provider-aws/v2/cmd/clusterawsadm/configreader"
	awscreds "sigs.k8s.io/cluster-api-provider-aws/v2/cmd/clusterawsadm/credentials"
	iamv1 "sigs.k8s.io/cluster-api-provider-aws/v2/iam/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

const (
	defaultNodeMinMemoryInGB       = 8
	defaultNodeMinCPU              = 2
	cpuArm64                       = "arm64"
	tanzuMissionControlPoliciesSID = "tmccloudvmwarecom"
	tanzuEBSCSIPoliciesSID         = "tanzuebscsi"
)

type client struct {
	credentials awscreds.AWSCredentials
	session     *session.Session
}

const (
	// DefaultCloudFormationBootstrapUserName is the default username of boostrap user
	DefaultCloudFormationBootstrapUserName = "bootstrapper.tkg.cloud.vmware.com"
	// DefaultCloudFormationNameSuffix is the default cloudformation suffix
	DefaultCloudFormationNameSuffix = ".tkg.cloud.vmware.com"
	// DefaultCloudFormationStackName is the default cloudformtation stack name
	DefaultCloudFormationStackName = "tkg-cloud-vmware-com"
	// KeyAWSRegion is the aws region key
	KeyAWSRegion = "region"
	// KeyAWSAccessKeyID is the aws_access_key_id key
	KeyAWSAccessKeyID = "aws_access_key_id"
	// KeyAWSSecretAccessKey is the aws_secret_access_key key
	KeyAWSSecretAccessKey = "aws_secret_access_key" // #nosec
	// KeyAWSSessionToken is the aws_session_token key
	KeyAWSSessionToken = "aws_session_token" // #nosec
)

// NewFromEncodedCrendentials creates an AWS Client from encoded credentials
func NewFromEncodedCrendentials(creds string) (Client, error) {
	awsCreds := awscreds.AWSCredentials{}

	credLines := strings.Split(creds, "\n")

	for _, line := range credLines {
		strs := strings.Split(line, " = ")

		if len(strs) != 2 {
			continue
		}

		if strs[0] == KeyAWSRegion {
			awsCreds.Region = strs[1]
		}

		if strs[0] == KeyAWSAccessKeyID {
			awsCreds.AccessKeyID = strs[1]
		}

		if strs[0] == KeyAWSSecretAccessKey {
			awsCreds.SecretAccessKey = strs[1]
		}

		if strs[0] == KeyAWSSessionToken {
			awsCreds.SessionToken = strs[1]
		}
	}

	if awsCreds.Region == "" || awsCreds.AccessKeyID == "" || awsCreds.SecretAccessKey == "" {
		return nil, errors.New("the AWS credentials are not valid")
	}

	return New(awsCreds)
}

// New creates an AWS client
func New(creds awscreds.AWSCredentials) (Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(creds.Region),
		Credentials: credentials.NewStaticCredentials(
			creds.AccessKeyID,
			creds.SecretAccessKey,
			creds.SessionToken,
		),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create aws session")
	}

	return &client{session: sess, credentials: creds}, nil
}

func (c *client) VerifyAccount() error {
	if c.session == nil {
		return errors.New("uninitialized aws client")
	}
	// check if the user credentials are correct by calling ListRegions api
	svc := ec2.New(c.session)

	_, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{})

	return err
}

func (c *client) ListVPCs() ([]*models.Vpc, error) {
	if c.session == nil {
		return nil, errors.New("uninitialized aws client")
	}
	svc := ec2.New(c.session)

	results, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, errors.Wrap(err, "cannot get the list of vpcs under current account")
	}

	vpcs := []*models.Vpc{}

	for _, r := range results.Vpcs {
		obj := &models.Vpc{ID: *r.VpcId, Cidr: *r.CidrBlock}
		vpcs = append(vpcs, obj)
	}

	return vpcs, nil
}

func (c *client) EncodeCredentials() (string, error) {
	return c.credentials.RenderBase64EncodedAWSDefaultProfile()
}

func (c *client) ListAvailabilityZones() ([]*models.AWSAvailabilityZone, error) {
	azs := []*models.AWSAvailabilityZone{}
	if c.session == nil {
		return nil, errors.New("uninitialized aws client")
	}
	svc := ec2.New(c.session)

	results, err := svc.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get availability zones under region %s", c.credentials.Region)
	}

	for _, r := range results.AvailabilityZones {
		obj := &models.AWSAvailabilityZone{ID: *r.ZoneId, Name: *r.ZoneName}
		azs = append(azs, obj)
	}
	return azs, nil
}

func (c *client) ListRegionsByUser() ([]string, error) {
	regions := []string{}

	if c.session == nil {
		return regions, errors.New("uninitialized aws client")
	}
	svc := ec2.New(c.session)

	results, err := svc.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get availability zones under region %s", c.credentials.Region)
	}

	for _, r := range results.Regions {
		regions = append(regions, *r.RegionName)
	}
	return regions, nil
}

func (c *client) GetSubnetGatewayAssociations(vpcID string) (map[string]bool, error) {
	subnetGatewayMap := map[string]bool{}
	if c.session == nil {
		return nil, errors.New("uninitialized aws client")
	}
	svc := ec2.New(c.session)

	results, err := svc.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{{
			Name: aws.String("vpc-id"), Values: []*string{aws.String(vpcID)},
		}},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get subnets for vpc %s", vpcID)
	}

	for _, routeTable := range results.RouteTables {
		var isPublic bool
		for _, route := range routeTable.Routes {
			if route.GatewayId != nil && strings.HasPrefix(*route.GatewayId, "igw") {
				isPublic = true
			}
		}

		for _, association := range routeTable.Associations {
			if association.SubnetId != nil && isPublic {
				subnetGatewayMap[*association.SubnetId] = true
			}
		}
	}

	return subnetGatewayMap, nil
}

func (c *client) ListSubnets(vpcID string) ([]*models.AWSSubnet, error) {
	subnets := []*models.AWSSubnet{}
	if c.session == nil {
		return nil, errors.New("uninitialized aws client")
	}
	svc := ec2.New(c.session)

	subnetGatewayMap, routeErr := c.GetSubnetGatewayAssociations(vpcID)

	if routeErr != nil {
		return nil, errors.Wrapf(routeErr, "cannot get subnets for vpc %s", vpcID)
	}

	results, err := svc.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{{
			Name:   aws.String("vpc-id"),
			Values: []*string{aws.String(vpcID)},
		}},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get subnets for vpc %s", vpcID)
	}

	for _, r := range results.Subnets {
		_, hasKey := subnetGatewayMap[*r.SubnetId]
		obj := &models.AWSSubnet{
			AvailabilityZoneID:   *r.AvailabilityZoneId,
			AvailabilityZoneName: *r.AvailabilityZone,
			Cidr:                 *r.CidrBlock,
			ID:                   *r.SubnetId,
			State:                *r.State,
			VpcID:                *r.VpcId,
			IsPublic:             &hasKey,
		}
		subnets = append(subnets, obj)
	}
	return subnets, nil
}

func (c *client) CreateCloudFormationStack() error {
	template, err := c.GenerateBootstrapTemplate(GenerateBootstrapTemplateInput{})
	if err != nil {
		return err
	}
	return c.CreateCloudFormationStackWithTemplate(template)
}

func (c *client) CreateCloudFormationStackWithTemplate(template *bootstrap.Template) error {
	cfnSvc := cloudformation.NewService(cfn.New(c.session))
	if err := cfnSvc.ReconcileBootstrapStack(template.Spec.StackName, *template.RenderCloudFormation(), nil); err != nil {
		return err
	}
	return cfnSvc.ShowStackResources(template.Spec.StackName)
}

// GenerateBootstrapTemplateInput is the input to the GenerateBootstrapTemplate func
type GenerateBootstrapTemplateInput struct {
	// BootstrapConfigFile is the path to a CAPA bootstrapv1 configuration file that can be used
	// to customize IAM policies
	BootstrapConfigFile string
	// DisableTanzuMissionControlPermissions if true will remove IAM permissions for use by Tanzu Mission Control
	// from all nodes
	DisableTanzuMissionControlPermissions bool
}

// GenerateBootstrapTemplate generates a wrapped CAPA bootstrapv1 configuration specification that controls
// the generation of CloudFormation stacks
func (c *client) GenerateBootstrapTemplate(i GenerateBootstrapTemplateInput) (*bootstrap.Template, error) {
	template := bootstrap.NewTemplate()
	if i.BootstrapConfigFile != "" {
		spec, err := configreader.LoadConfigFile(i.BootstrapConfigFile)
		if err != nil {
			return nil, err
		}
		template.Spec = &spec.Spec
	}
	setDefaultsBootstrapTemplate(&template)
	if !i.DisableTanzuMissionControlPermissions {
		ensureTanzuMissionControlPermissions(&template)
	}
	ensureEBSCSiPermissions(&template)
	return &template, nil
}

func ensureEBSCSiPermissions(t *bootstrap.Template) {
	// WA https://github.com/kubernetes-sigs/cluster-api-provider-aws/issues/3527
	// t.Spec.ControlPlane.EnableCSIPolicy = true not work
	t.Spec.ControlPlane.ExtraStatements = ensureTanzuEBSCSIPermissionsForRole(t.Spec.ControlPlane.ExtraStatements)
}

func ensureTanzuEBSCSIPermissionsForRole(statements []iamv1.StatementEntry) []iamv1.StatementEntry {
	ebscsiStatementEntry := iamv1.StatementEntry{
		Sid:      tanzuEBSCSIPoliciesSID,
		Effect:   iamv1.EffectAllow,
		Resource: iamv1.Resources{iamv1.Any},
		Action: iamv1.Actions{
			"ec2:AttachVolume",
			"ec2:CreateSnapshot",
			"ec2:CreateTags",
			"ec2:CreateVolume",
			"ec2:DeleteSnapshot",
			"ec2:DeleteTags",
			"ec2:DeleteVolume",
			"ec2:DescribeAvailabilityZones",
			"ec2:DescribeInstances",
			"ec2:DescribeSnapshots",
			"ec2:DescribeTags",
			"ec2:DescribeVolumes",
			"ec2:DescribeVolumesModifications",
			"ec2:DetachVolume",
			"ec2:ModifyVolume",
		},
	}
	for i, statementEntry := range statements {
		if statementEntry.Sid == tanzuEBSCSIPoliciesSID {
			statements[i] = ebscsiStatementEntry
			return statements
		}
	}
	statements = append(statements, ebscsiStatementEntry)
	return statements
}

func ensureTanzuMissionControlPermissions(t *bootstrap.Template) {
	t.Spec.Nodes.ExtraStatements = ensureTanzuMissionControlPermissionsForRole(t.Spec.Nodes.ExtraStatements)
	t.Spec.ControlPlane.ExtraStatements = ensureTanzuMissionControlPermissionsForRole(t.Spec.ControlPlane.ExtraStatements)
}

func ensureTanzuMissionControlPermissionsForRole(statements []iamv1.StatementEntry) []iamv1.StatementEntry {
	tmcStatementEntry := iamv1.StatementEntry{
		Sid:    tanzuMissionControlPoliciesSID,
		Effect: iamv1.EffectAllow,
		Action: iamv1.Actions{
			"ec2:DescribeKeyPairs",
			"ec2:DescribeInstanceTypeOfferings",
			"ec2:DescribeInstanceTypes",
			"ec2:DescribeAvailabilityZones",
			"ec2:DescribeSubnets",
			"ec2:DescribeRouteTables",
			"ec2:DescribeVpcs",
			"ec2:DescribeNatGateways",
			"ec2:DescribeAddresses",
			"elasticloadbalancing:DescribeLoadBalancers",
			"servicequotas:ListServiceQuotas",
			"iam:GetPolicy",
			"iam:ListAttachedRolePolicies",
			"iam:GetPolicyVersion",
			"iam:ListRoleTags",
		},
		Resource: iamv1.Resources{iamv1.Any},
	}
	for i, statementEntry := range statements {
		if statementEntry.Sid == tanzuMissionControlPoliciesSID {
			statements[i] = tmcStatementEntry
			return statements
		}
	}
	statements = append(statements, tmcStatementEntry)
	return statements
}

func setDefaultsBootstrapTemplate(t *bootstrap.Template) {
	t.Spec.NameSuffix = utilpointer.StringPtr(DefaultCloudFormationNameSuffix)
	t.Spec.StackName = DefaultCloudFormationStackName
	t.Spec.BootstrapUser.UserName = DefaultCloudFormationBootstrapUserName
	// Experimental EKS support in CAPA graduated and is enabled by default.
	// Explicitly disabling it since TKG doesn't support creating EKS clusters.
	t.Spec.EKS.Disable = true
}

func (c *client) ListInstanceTypes(optionalAZName string) ([]string, error) {
	if c.session == nil {
		return nil, errors.New("uninitialized aws client")
	}
	svc := ec2.New(c.session)

	azs, err := svc.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{})
	if err != nil {
		return nil, errors.Wrap(err, "cannot get availability zones under region")
	}

	// Return instance types for a single AZ if the AZ name is specified
	if optionalAZName != "" {
		filters := []*ec2.Filter{
			{
				Name:   aws.String("location"),
				Values: []*string{&optionalAZName},
			},
		}
		filter := &ec2.DescribeInstanceTypeOfferingsInput{Filters: filters, LocationType: aws.String(ec2.LocationTypeAvailabilityZone)}
		// retrieve instance type offered for particular availability zone
		instances, err := getInstanceTypeOffering(svc, filter)
		if err != nil {
			return nil, err
		}
		candidates, err := getInstanceTypes(svc)
		if err != nil {
			return nil, err
		}
		// filter and build a list of unsupported instance types
		unsupportedInstances := getUnsupportedInstanceTypes(candidates)
		diffInstancesPerAz := getSetDifference(instances, unsupportedInstances)

		return diffInstancesPerAz, nil
	}

	candidates, err := getInstanceTypes(svc)
	if err != nil {
		return nil, err
	}

	// filter and build a list of unsupported instance types
	unsupportedInstances := getUnsupportedInstanceTypes(candidates)
	var res []string
	for i, az := range azs.AvailabilityZones {
		filters := []*ec2.Filter{
			{
				Name:   aws.String("location"),
				Values: []*string{az.ZoneName},
			},
		}
		filter := &ec2.DescribeInstanceTypeOfferingsInput{Filters: filters, LocationType: aws.String(ec2.LocationTypeAvailabilityZone)}
		// retrieve instance type offered for particular availability zone
		instances, err := getInstanceTypeOffering(svc, filter)

		// Do the set difference operation to filter out the unsupported instances
		diffInstancesPerAz := getSetDifference(instances, unsupportedInstances)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			res = diffInstancesPerAz
		} else {
			res = union(res, diffInstancesPerAz)
		}
	}
	return res, nil
}

// meetsNodeMinimumRequirements checks if the instance type meets the minimum requirements of the target node
func meetsNodeMinimumRequirements(instanceInfo *ec2.InstanceTypeInfo) bool {
	return ((*instanceInfo.MemoryInfo.SizeInMiB / 1024) >= defaultNodeMinMemoryInGB) &&
		(*instanceInfo.VCpuInfo.DefaultVCpus >= defaultNodeMinCPU)
}

// isSupportedInstance checks feature flag to determine if CPU supported architecture filters should be applied
func isSupportedInstance(instanceInfo *ec2.InstanceTypeInfo) bool {
	if config.IsFeatureActivated(constants.FeatureFlagAwsInstanceTypesExcludeArm) {
		// feature flag is active; apply filtering
		return isSupportedCPUArchitecture(instanceInfo)
	}
	return true
}

// isSupportedCPUArchitecture checks if the node includes unsupported CPU architecture
func isSupportedCPUArchitecture(instanceInfo *ec2.InstanceTypeInfo) bool {
	// SupportedArchitectures []*string is a list of CPU architectures which are supported in the instance type
	// if arm64 is present then exclude this instance type
	for _, cpu := range instanceInfo.ProcessorInfo.SupportedArchitectures {
		if *cpu == cpuArm64 {
			return false
		}
	}
	return true
}

func getInstanceTypes(svc *ec2.EC2) ([]*ec2.InstanceTypeInfo, error) {
	var descInstanceTypeOps []*ec2.DescribeInstanceTypesOutput

	err := svc.DescribeInstanceTypesPages(nil,
		func(page *ec2.DescribeInstanceTypesOutput, lastPage bool) bool {
			descInstanceTypeOps = append(descInstanceTypeOps, page)
			return true
		})
	if err != nil {
		return nil, err
	}
	// iterate over all pages of the response and create a single array
	res := []*ec2.InstanceTypeInfo{}
	for _, descInstanceTypeOp := range descInstanceTypeOps {
		res = append(res, descInstanceTypeOp.InstanceTypes...)
	}
	return res, err
}

func getInstanceTypeOffering(svc *ec2.EC2, filter *ec2.DescribeInstanceTypeOfferingsInput) ([]string, error) {
	var instanceTypeOfferingOps []*ec2.DescribeInstanceTypeOfferingsOutput

	err := svc.DescribeInstanceTypeOfferingsPages(filter,
		func(page *ec2.DescribeInstanceTypeOfferingsOutput, lastPage bool) bool {
			instanceTypeOfferingOps = append(instanceTypeOfferingOps, page)
			return true
		})
	if err != nil {
		return nil, err
	}
	var res []string

	// iterate over all pages of the response and create a single array
	for _, instanceTypeOfferingOp := range instanceTypeOfferingOps {
		for _, t := range instanceTypeOfferingOp.InstanceTypeOfferings {
			res = append(res, *t.InstanceType)
		}
	}

	return res, err
}

func getSetDifference(a, b []string) (diff []string) {
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return diff
}

// getUnsupportedInstanceTypes generates list of unsupported instance types
func getUnsupportedInstanceTypes(instanceTypes []*ec2.InstanceTypeInfo) (ret []string) {
	for _, instanceTypeObj := range instanceTypes {
		// using node requirements as the criteria to filter the unsupported instance type
		supportedInstanceType := meetsNodeMinimumRequirements(instanceTypeObj) && isSupportedInstance(instanceTypeObj)
		if !supportedInstanceType {
			ret = append(ret, aws.StringValue(instanceTypeObj.InstanceType))
		}
	}
	return
}

func union(a, b []string) (c []string) {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
		c = append(c, item)
	}

	for _, item := range b {
		if _, ok := m[item]; !ok {
			c = append(c, item)
		}
	}
	return c
}

func (c *client) ListCloudFormationStacks() ([]string, error) {
	res := []string{}
	if c.session == nil {
		return res, errors.New("uninitialized aws client")
	}

	svc := cfn.New(c.session)

	stacks, err := svc.DescribeStacks(&cfn.DescribeStacksInput{})
	if err != nil {
		return res, errors.Wrap(err, "cannot get list of CloudFormation Stacks")
	}

	for _, stack := range stacks.Stacks {
		if stack.StackName == nil {
			continue
		}
		res = append(res, *stack.StackName)
	}

	return res, nil
}

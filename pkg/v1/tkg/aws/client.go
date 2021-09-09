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
	"sigs.k8s.io/cluster-api-provider-aws/cmd/clusterawsadm/cloudformation/bootstrap"
	cloudformation "sigs.k8s.io/cluster-api-provider-aws/cmd/clusterawsadm/cloudformation/service"
	awscreds "sigs.k8s.io/cluster-api-provider-aws/cmd/clusterawsadm/credentials"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

const (
	defaultControlPlaneMinMemoryInGB = 8
	defaultControlPlaneMinCPU        = 2
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
	template := bootstrap.NewTemplate()
	setDefaultCloudFormationTemplateValue(&template)
	cfnSvc := cloudformation.NewService(cfn.New(c.session))
	err := cfnSvc.ReconcileBootstrapStack(template.Spec.StackName, *template.RenderCloudFormation())
	if err != nil {
		return err
	}

	return cfnSvc.ShowStackResources(template.Spec.StackName)
}

func setDefaultCloudFormationTemplateValue(t *bootstrap.Template) {
	t.Spec.NameSuffix = utilpointer.StringPtr(DefaultCloudFormationNameSuffix)
	t.Spec.StackName = DefaultCloudFormationStackName
	t.Spec.BootstrapUser.UserName = DefaultCloudFormationBootstrapUserName
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
		filteredInstances := filterInstanceType(candidates)
		diffInstancesPerAz := getSetDifference(instances, filteredInstances)

		return diffInstancesPerAz, nil
	}

	candidates, err := getInstanceTypes(svc)
	if err != nil {
		return nil, err
	}

	// here we filter the unsupported instance types
	filteredInstances := filterInstanceType(candidates)
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
		diffInstancesPerAz := getSetDifference(instances, filteredInstances)
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

// meetsNodeMininumRequirements checks if the node meet the minimum requirements of the controlplane node
func meetsNodeMininumRequirements(instanceInfo *ec2.InstanceTypeInfo) bool {
	return ((*instanceInfo.MemoryInfo.SizeInMiB / 1024) >= defaultControlPlaneMinMemoryInGB) &&
		(*instanceInfo.VCpuInfo.DefaultVCpus >= defaultControlPlaneMinCPU)
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

func filterInstanceType(ss []*ec2.InstanceTypeInfo) (ret []string) {
	for _, s := range ss {
		// using control-plane node requirements as the criteria to filter the unsupported instance type
		controlPlaneTest := meetsNodeMininumRequirements(s)
		if !controlPlaneTest {
			ret = append(ret, aws.StringValue(s.InstanceType))
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

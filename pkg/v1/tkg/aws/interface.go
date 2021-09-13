// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package aws defines functions to connect to the AWS cloud provider
package aws

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

//go:generate counterfeiter -o ../fakes/awsclient.go --fake-name AWSClient . Client

// Client defines methods to access AWS inventory
type Client interface {
	VerifyAccount() error
	ListVPCs() ([]*models.Vpc, error)
	EncodeCredentials() (string, error)
	ListAvailabilityZones() ([]*models.AWSAvailabilityZone, error)
	ListRegionsByUser() ([]string, error)
	GetSubnetGatewayAssociations(vpcID string) (map[string]bool, error)
	ListSubnets(vpcID string) ([]*models.AWSSubnet, error)
	CreateCloudFormationStack() error
	ListInstanceTypes(optionalAZName string) ([]string, error)
	ListCloudFormationStacks() ([]string, error)
}

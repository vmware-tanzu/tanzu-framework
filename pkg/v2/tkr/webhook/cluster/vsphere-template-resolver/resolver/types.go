// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package templateresolver provides vsphere template resolution.
package templateresolver

import (
	"fmt"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

type VSphereContext struct {
	Server             string
	Username           string
	Password           string
	DataCenter         string
	TLSThumbprint      string
	InsecureSkipVerify bool
}

// Query sets constraints for resolution of vSphere OVA templates. Its structure reflects Cluster API cluster topology.
type Query struct {
	// ControlPlane specifies the Query for the control plane.
	// Set to nil if we want to skip resolving the control plane part.
	ControlPlane *TemplateQuery

	// MachineDeployments specifies the OSImageQueries for worker machine deployments.
	// An individual machine deployment query part may be set to nil if we want to skip resolving it.
	MachineDeployments []*TemplateQuery
}

func (q Query) String() string {
	return fmt.Sprintf("{controlPlane: %s, machineDeployments: %s}", q.ControlPlane, q.MachineDeployments)
}

// TemplateQuery sets constraints for resolution of vSphere OVA templates for the control plane or a machine deployment of a cluster.
type TemplateQuery struct {
	// OVAVersion is a version of the template
	OVAVersion string
	// OSInfo is the OS information the template should satisfy
	OSInfo runv1.OSInfo
}

const strNil = "nil"

func (q *TemplateQuery) String() string {
	if q == nil {
		return strNil
	}
	return fmt.Sprintf("{OVA version: '%s', OSInfo: '%s'}", q.OVAVersion, q.OSInfo)
}

// Result carries the results of vSphere OVA template resolution. Its structure reflects Cluster API cluster topology.
type Result struct {
	// ControlPlane carries the Result for the  control plane.
	// It is set to nil if resolving the control plane part was skipped.
	ControlPlane *OVATemplateResult

	// ControlPlane carries the Result for worker machine deployments.
	// An individual machine deployment result is set to nil if resolving it was skipped.
	MachineDeployments []*OVATemplateResult

	// UsefulErrorMessage carries the errors resulted in template resolution
	UsefulErrorMessage string
}

func (r Result) String() string {
	return fmt.Sprintf("{controlPlane: %s, machineDeployments: %s usefulErrorMessage:'%s'}", r.ControlPlane, r.MachineDeployments, r.UsefulErrorMessage)
}

// OVATemplateResult carries the results of OVA template resolution for the control plane or a machine deployment of a cluster.
type OVATemplateResult struct {
	// TemplatePath is the path of the template. If empty, then no template satisfied the query.
	TemplatePath string
	// TemplateMOID is the MOID of the template. If empty, then no template satisfied the query.
	TemplateMOID string
}

func (r *OVATemplateResult) String() string {
	if r == nil {
		return strNil
	}
	return fmt.Sprintf("{TemplatePath: '%s', TemplateMOID: '%s'}", r.TemplatePath, r.TemplateMOID)
}

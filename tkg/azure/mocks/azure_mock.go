// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/azure/interface.go

// Package azure is a generated GoMock package.
package azure

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// VerifyAccount mocks base method.
func (m *MockClient) VerifyAccount(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyAccount", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyAccount indicates an expected call of VerifyAccount.
func (mr *MockClientMockRecorder) VerifyAccount(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyAccount", reflect.TypeOf((*MockClient)(nil).VerifyAccount), ctx)
}

// ListResourceGroups mocks base method.
func (m *MockClient) ListResourceGroups(ctx context.Context, location string) ([]*models.AzureResourceGroup, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListResourceGroups", ctx, location)
	ret0, _ := ret[0].([]*models.AzureResourceGroup)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListResourceGroups indicates an expected call of ListResourceGroups.
func (mr *MockClientMockRecorder) ListResourceGroups(ctx, location interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListResourceGroups", reflect.TypeOf((*MockClient)(nil).ListResourceGroups), ctx, location)
}

// ListVirtualNetworks mocks base method.
func (m *MockClient) ListVirtualNetworks(ctx context.Context, resourceGroup, location string) ([]*models.AzureVirtualNetwork, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListVirtualNetworks", ctx, resourceGroup, location)
	ret0, _ := ret[0].([]*models.AzureVirtualNetwork)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListVirtualNetworks indicates an expected call of ListVirtualNetworks.
func (mr *MockClientMockRecorder) ListVirtualNetworks(ctx, resourceGroup, location interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListVirtualNetworks", reflect.TypeOf((*MockClient)(nil).ListVirtualNetworks), ctx, resourceGroup, location)
}

// CreateResourceGroup mocks base method.
func (m *MockClient) CreateResourceGroup(ctx context.Context, resourceGroupName, location string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateResourceGroup", ctx, resourceGroupName, location)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateResourceGroup indicates an expected call of CreateResourceGroup.
func (mr *MockClientMockRecorder) CreateResourceGroup(ctx, resourceGroupName, location interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateResourceGroup", reflect.TypeOf((*MockClient)(nil).CreateResourceGroup), ctx, resourceGroupName, location)
}

// CreateVirtualNetwork mocks base method.
func (m *MockClient) CreateVirtualNetwork(ctx context.Context, resourceGroupName, virtualNetworkName, cidrBlock, location string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateVirtualNetwork", ctx, resourceGroupName, virtualNetworkName, cidrBlock, location)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateVirtualNetwork indicates an expected call of CreateVirtualNetwork.
func (mr *MockClientMockRecorder) CreateVirtualNetwork(ctx, resourceGroupName, virtualNetworkName, cidrBlock, location interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateVirtualNetwork", reflect.TypeOf((*MockClient)(nil).CreateVirtualNetwork), ctx, resourceGroupName, virtualNetworkName, cidrBlock, location)
}

// GetAzureRegions mocks base method.
func (m *MockClient) GetAzureRegions(ctx context.Context) ([]*models.AzureLocation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAzureRegions", ctx)
	ret0, _ := ret[0].([]*models.AzureLocation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAzureRegions indicates an expected call of GetAzureRegions.
func (mr *MockClientMockRecorder) GetAzureRegions(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAzureRegions", reflect.TypeOf((*MockClient)(nil).GetAzureRegions), ctx)
}

// GetAzureInstanceTypesForRegion mocks base method.
func (m *MockClient) GetAzureInstanceTypesForRegion(ctx context.Context, region string) ([]*models.AzureInstanceType, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAzureInstanceTypesForRegion", ctx, region)
	ret0, _ := ret[0].([]*models.AzureInstanceType)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAzureInstanceTypesForRegion indicates an expected call of GetAzureInstanceTypesForRegion.
func (mr *MockClientMockRecorder) GetAzureInstanceTypesForRegion(ctx, region interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAzureInstanceTypesForRegion", reflect.TypeOf((*MockClient)(nil).GetAzureInstanceTypesForRegion), ctx, region)
}

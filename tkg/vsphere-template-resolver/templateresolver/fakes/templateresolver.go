// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"context"
	"sync"

	"github.com/vmware-tanzu/tanzu-framework/tkg/vc"
	"github.com/vmware-tanzu/tanzu-framework/tkg/vsphere-template-resolver/templateresolver"
)

type TemplateResolver struct {
	GetVSphereEndpointStub        func(*templateresolver.VSphereContext) (vc.Client, error)
	getVSphereEndpointMutex       sync.RWMutex
	getVSphereEndpointArgsForCall []struct {
		arg1 *templateresolver.VSphereContext
	}
	getVSphereEndpointReturns struct {
		result1 vc.Client
		result2 error
	}
	getVSphereEndpointReturnsOnCall map[int]struct {
		result1 vc.Client
		result2 error
	}
	ResolveStub        func(context.Context, *templateresolver.VSphereContext, templateresolver.Query, vc.Client) templateresolver.Result
	resolveMutex       sync.RWMutex
	resolveArgsForCall []struct {
		arg1 context.Context
		arg2 *templateresolver.VSphereContext
		arg3 templateresolver.Query
		arg4 vc.Client
	}
	resolveReturns struct {
		result1 templateresolver.Result
	}
	resolveReturnsOnCall map[int]struct {
		result1 templateresolver.Result
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *TemplateResolver) GetVSphereEndpoint(arg1 *templateresolver.VSphereContext) (vc.Client, error) {
	fake.getVSphereEndpointMutex.Lock()
	ret, specificReturn := fake.getVSphereEndpointReturnsOnCall[len(fake.getVSphereEndpointArgsForCall)]
	fake.getVSphereEndpointArgsForCall = append(fake.getVSphereEndpointArgsForCall, struct {
		arg1 *templateresolver.VSphereContext
	}{arg1})
	stub := fake.GetVSphereEndpointStub
	fakeReturns := fake.getVSphereEndpointReturns
	fake.recordInvocation("GetVSphereEndpoint", []interface{}{arg1})
	fake.getVSphereEndpointMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *TemplateResolver) GetVSphereEndpointCallCount() int {
	fake.getVSphereEndpointMutex.RLock()
	defer fake.getVSphereEndpointMutex.RUnlock()
	return len(fake.getVSphereEndpointArgsForCall)
}

func (fake *TemplateResolver) GetVSphereEndpointCalls(stub func(*templateresolver.VSphereContext) (vc.Client, error)) {
	fake.getVSphereEndpointMutex.Lock()
	defer fake.getVSphereEndpointMutex.Unlock()
	fake.GetVSphereEndpointStub = stub
}

func (fake *TemplateResolver) GetVSphereEndpointArgsForCall(i int) *templateresolver.VSphereContext {
	fake.getVSphereEndpointMutex.RLock()
	defer fake.getVSphereEndpointMutex.RUnlock()
	argsForCall := fake.getVSphereEndpointArgsForCall[i]
	return argsForCall.arg1
}

func (fake *TemplateResolver) GetVSphereEndpointReturns(result1 vc.Client, result2 error) {
	fake.getVSphereEndpointMutex.Lock()
	defer fake.getVSphereEndpointMutex.Unlock()
	fake.GetVSphereEndpointStub = nil
	fake.getVSphereEndpointReturns = struct {
		result1 vc.Client
		result2 error
	}{result1, result2}
}

func (fake *TemplateResolver) GetVSphereEndpointReturnsOnCall(i int, result1 vc.Client, result2 error) {
	fake.getVSphereEndpointMutex.Lock()
	defer fake.getVSphereEndpointMutex.Unlock()
	fake.GetVSphereEndpointStub = nil
	if fake.getVSphereEndpointReturnsOnCall == nil {
		fake.getVSphereEndpointReturnsOnCall = make(map[int]struct {
			result1 vc.Client
			result2 error
		})
	}
	fake.getVSphereEndpointReturnsOnCall[i] = struct {
		result1 vc.Client
		result2 error
	}{result1, result2}
}

func (fake *TemplateResolver) Resolve(arg1 context.Context, arg2 *templateresolver.VSphereContext, arg3 templateresolver.Query, arg4 vc.Client) templateresolver.Result {
	fake.resolveMutex.Lock()
	ret, specificReturn := fake.resolveReturnsOnCall[len(fake.resolveArgsForCall)]
	fake.resolveArgsForCall = append(fake.resolveArgsForCall, struct {
		arg1 context.Context
		arg2 *templateresolver.VSphereContext
		arg3 templateresolver.Query
		arg4 vc.Client
	}{arg1, arg2, arg3, arg4})
	stub := fake.ResolveStub
	fakeReturns := fake.resolveReturns
	fake.recordInvocation("Resolve", []interface{}{arg1, arg2, arg3, arg4})
	fake.resolveMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *TemplateResolver) ResolveCallCount() int {
	fake.resolveMutex.RLock()
	defer fake.resolveMutex.RUnlock()
	return len(fake.resolveArgsForCall)
}

func (fake *TemplateResolver) ResolveCalls(stub func(context.Context, *templateresolver.VSphereContext, templateresolver.Query, vc.Client) templateresolver.Result) {
	fake.resolveMutex.Lock()
	defer fake.resolveMutex.Unlock()
	fake.ResolveStub = stub
}

func (fake *TemplateResolver) ResolveArgsForCall(i int) (context.Context, *templateresolver.VSphereContext, templateresolver.Query, vc.Client) {
	fake.resolveMutex.RLock()
	defer fake.resolveMutex.RUnlock()
	argsForCall := fake.resolveArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *TemplateResolver) ResolveReturns(result1 templateresolver.Result) {
	fake.resolveMutex.Lock()
	defer fake.resolveMutex.Unlock()
	fake.ResolveStub = nil
	fake.resolveReturns = struct {
		result1 templateresolver.Result
	}{result1}
}

func (fake *TemplateResolver) ResolveReturnsOnCall(i int, result1 templateresolver.Result) {
	fake.resolveMutex.Lock()
	defer fake.resolveMutex.Unlock()
	fake.ResolveStub = nil
	if fake.resolveReturnsOnCall == nil {
		fake.resolveReturnsOnCall = make(map[int]struct {
			result1 templateresolver.Result
		})
	}
	fake.resolveReturnsOnCall[i] = struct {
		result1 templateresolver.Result
	}{result1}
}

func (fake *TemplateResolver) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getVSphereEndpointMutex.RLock()
	defer fake.getVSphereEndpointMutex.RUnlock()
	fake.resolveMutex.RLock()
	defer fake.resolveMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *TemplateResolver) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ templateresolver.TemplateResolver = new(TemplateResolver)

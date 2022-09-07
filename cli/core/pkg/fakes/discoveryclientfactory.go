// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cluster"
)

type DiscoveryClientFactory struct {
	NewDiscoveryClientForConfigStub        func(*rest.Config) (discovery.DiscoveryInterface, error)
	newDiscoveryClientForConfigMutex       sync.RWMutex
	newDiscoveryClientForConfigArgsForCall []struct {
		arg1 *rest.Config
	}
	newDiscoveryClientForConfigReturns struct {
		result1 discovery.DiscoveryInterface
		result2 error
	}
	newDiscoveryClientForConfigReturnsOnCall map[int]struct {
		result1 discovery.DiscoveryInterface
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *DiscoveryClientFactory) NewDiscoveryClientForConfig(arg1 *rest.Config) (discovery.DiscoveryInterface, error) {
	fake.newDiscoveryClientForConfigMutex.Lock()
	ret, specificReturn := fake.newDiscoveryClientForConfigReturnsOnCall[len(fake.newDiscoveryClientForConfigArgsForCall)]
	fake.newDiscoveryClientForConfigArgsForCall = append(fake.newDiscoveryClientForConfigArgsForCall, struct {
		arg1 *rest.Config
	}{arg1})
	stub := fake.NewDiscoveryClientForConfigStub
	fakeReturns := fake.newDiscoveryClientForConfigReturns
	fake.recordInvocation("NewDiscoveryClientForConfig", []interface{}{arg1})
	fake.newDiscoveryClientForConfigMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *DiscoveryClientFactory) NewDiscoveryClientForConfigCallCount() int {
	fake.newDiscoveryClientForConfigMutex.RLock()
	defer fake.newDiscoveryClientForConfigMutex.RUnlock()
	return len(fake.newDiscoveryClientForConfigArgsForCall)
}

func (fake *DiscoveryClientFactory) NewDiscoveryClientForConfigCalls(stub func(*rest.Config) (discovery.DiscoveryInterface, error)) {
	fake.newDiscoveryClientForConfigMutex.Lock()
	defer fake.newDiscoveryClientForConfigMutex.Unlock()
	fake.NewDiscoveryClientForConfigStub = stub
}

func (fake *DiscoveryClientFactory) NewDiscoveryClientForConfigArgsForCall(i int) *rest.Config {
	fake.newDiscoveryClientForConfigMutex.RLock()
	defer fake.newDiscoveryClientForConfigMutex.RUnlock()
	argsForCall := fake.newDiscoveryClientForConfigArgsForCall[i]
	return argsForCall.arg1
}

func (fake *DiscoveryClientFactory) NewDiscoveryClientForConfigReturns(result1 discovery.DiscoveryInterface, result2 error) {
	fake.newDiscoveryClientForConfigMutex.Lock()
	defer fake.newDiscoveryClientForConfigMutex.Unlock()
	fake.NewDiscoveryClientForConfigStub = nil
	fake.newDiscoveryClientForConfigReturns = struct {
		result1 discovery.DiscoveryInterface
		result2 error
	}{result1, result2}
}

func (fake *DiscoveryClientFactory) NewDiscoveryClientForConfigReturnsOnCall(i int, result1 discovery.DiscoveryInterface, result2 error) {
	fake.newDiscoveryClientForConfigMutex.Lock()
	defer fake.newDiscoveryClientForConfigMutex.Unlock()
	fake.NewDiscoveryClientForConfigStub = nil
	if fake.newDiscoveryClientForConfigReturnsOnCall == nil {
		fake.newDiscoveryClientForConfigReturnsOnCall = make(map[int]struct {
			result1 discovery.DiscoveryInterface
			result2 error
		})
	}
	fake.newDiscoveryClientForConfigReturnsOnCall[i] = struct {
		result1 discovery.DiscoveryInterface
		result2 error
	}{result1, result2}
}

func (fake *DiscoveryClientFactory) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newDiscoveryClientForConfigMutex.RLock()
	defer fake.newDiscoveryClientForConfigMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *DiscoveryClientFactory) recordInvocation(key string, args []interface{}) {
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

var _ cluster.DiscoveryClientFactory = new(DiscoveryClientFactory)

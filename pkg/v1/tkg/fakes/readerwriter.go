// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
)

type TKGConfigReaderWriter struct {
	GetStub        func(string) (string, error)
	getMutex       sync.RWMutex
	getArgsForCall []struct {
		arg1 string
	}
	getReturns struct {
		result1 string
		result2 error
	}
	getReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	InitStub        func(string) error
	initMutex       sync.RWMutex
	initArgsForCall []struct {
		arg1 string
	}
	initReturns struct {
		result1 error
	}
	initReturnsOnCall map[int]struct {
		result1 error
	}
	MergeInConfigStub        func(string) error
	mergeInConfigMutex       sync.RWMutex
	mergeInConfigArgsForCall []struct {
		arg1 string
	}
	mergeInConfigReturns struct {
		result1 error
	}
	mergeInConfigReturnsOnCall map[int]struct {
		result1 error
	}
	SetStub        func(string, string)
	setMutex       sync.RWMutex
	setArgsForCall []struct {
		arg1 string
		arg2 string
	}
	UnmarshalKeyStub        func(string, interface{}) error
	unmarshalKeyMutex       sync.RWMutex
	unmarshalKeyArgsForCall []struct {
		arg1 string
		arg2 interface{}
	}
	unmarshalKeyReturns struct {
		result1 error
	}
	unmarshalKeyReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *TKGConfigReaderWriter) Get(arg1 string) (string, error) {
	fake.getMutex.Lock()
	ret, specificReturn := fake.getReturnsOnCall[len(fake.getArgsForCall)]
	fake.getArgsForCall = append(fake.getArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.GetStub
	fakeReturns := fake.getReturns
	fake.recordInvocation("Get", []interface{}{arg1})
	fake.getMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *TKGConfigReaderWriter) GetCallCount() int {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	return len(fake.getArgsForCall)
}

func (fake *TKGConfigReaderWriter) GetCalls(stub func(string) (string, error)) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = stub
}

func (fake *TKGConfigReaderWriter) GetArgsForCall(i int) string {
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	argsForCall := fake.getArgsForCall[i]
	return argsForCall.arg1
}

func (fake *TKGConfigReaderWriter) GetReturns(result1 string, result2 error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	fake.getReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *TKGConfigReaderWriter) GetReturnsOnCall(i int, result1 string, result2 error) {
	fake.getMutex.Lock()
	defer fake.getMutex.Unlock()
	fake.GetStub = nil
	if fake.getReturnsOnCall == nil {
		fake.getReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.getReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *TKGConfigReaderWriter) Init(arg1 string) error {
	fake.initMutex.Lock()
	ret, specificReturn := fake.initReturnsOnCall[len(fake.initArgsForCall)]
	fake.initArgsForCall = append(fake.initArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.InitStub
	fakeReturns := fake.initReturns
	fake.recordInvocation("Init", []interface{}{arg1})
	fake.initMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *TKGConfigReaderWriter) InitCallCount() int {
	fake.initMutex.RLock()
	defer fake.initMutex.RUnlock()
	return len(fake.initArgsForCall)
}

func (fake *TKGConfigReaderWriter) InitCalls(stub func(string) error) {
	fake.initMutex.Lock()
	defer fake.initMutex.Unlock()
	fake.InitStub = stub
}

func (fake *TKGConfigReaderWriter) InitArgsForCall(i int) string {
	fake.initMutex.RLock()
	defer fake.initMutex.RUnlock()
	argsForCall := fake.initArgsForCall[i]
	return argsForCall.arg1
}

func (fake *TKGConfigReaderWriter) InitReturns(result1 error) {
	fake.initMutex.Lock()
	defer fake.initMutex.Unlock()
	fake.InitStub = nil
	fake.initReturns = struct {
		result1 error
	}{result1}
}

func (fake *TKGConfigReaderWriter) InitReturnsOnCall(i int, result1 error) {
	fake.initMutex.Lock()
	defer fake.initMutex.Unlock()
	fake.InitStub = nil
	if fake.initReturnsOnCall == nil {
		fake.initReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.initReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *TKGConfigReaderWriter) MergeInConfig(arg1 string) error {
	fake.mergeInConfigMutex.Lock()
	ret, specificReturn := fake.mergeInConfigReturnsOnCall[len(fake.mergeInConfigArgsForCall)]
	fake.mergeInConfigArgsForCall = append(fake.mergeInConfigArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.MergeInConfigStub
	fakeReturns := fake.mergeInConfigReturns
	fake.recordInvocation("MergeInConfig", []interface{}{arg1})
	fake.mergeInConfigMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *TKGConfigReaderWriter) MergeInConfigCallCount() int {
	fake.mergeInConfigMutex.RLock()
	defer fake.mergeInConfigMutex.RUnlock()
	return len(fake.mergeInConfigArgsForCall)
}

func (fake *TKGConfigReaderWriter) MergeInConfigCalls(stub func(string) error) {
	fake.mergeInConfigMutex.Lock()
	defer fake.mergeInConfigMutex.Unlock()
	fake.MergeInConfigStub = stub
}

func (fake *TKGConfigReaderWriter) MergeInConfigArgsForCall(i int) string {
	fake.mergeInConfigMutex.RLock()
	defer fake.mergeInConfigMutex.RUnlock()
	argsForCall := fake.mergeInConfigArgsForCall[i]
	return argsForCall.arg1
}

func (fake *TKGConfigReaderWriter) MergeInConfigReturns(result1 error) {
	fake.mergeInConfigMutex.Lock()
	defer fake.mergeInConfigMutex.Unlock()
	fake.MergeInConfigStub = nil
	fake.mergeInConfigReturns = struct {
		result1 error
	}{result1}
}

func (fake *TKGConfigReaderWriter) MergeInConfigReturnsOnCall(i int, result1 error) {
	fake.mergeInConfigMutex.Lock()
	defer fake.mergeInConfigMutex.Unlock()
	fake.MergeInConfigStub = nil
	if fake.mergeInConfigReturnsOnCall == nil {
		fake.mergeInConfigReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.mergeInConfigReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *TKGConfigReaderWriter) Set(arg1 string, arg2 string) {
	fake.setMutex.Lock()
	fake.setArgsForCall = append(fake.setArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.SetStub
	fake.recordInvocation("Set", []interface{}{arg1, arg2})
	fake.setMutex.Unlock()
	if stub != nil {
		fake.SetStub(arg1, arg2)
	}
}

func (fake *TKGConfigReaderWriter) SetCallCount() int {
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	return len(fake.setArgsForCall)
}

func (fake *TKGConfigReaderWriter) SetCalls(stub func(string, string)) {
	fake.setMutex.Lock()
	defer fake.setMutex.Unlock()
	fake.SetStub = stub
}

func (fake *TKGConfigReaderWriter) SetArgsForCall(i int) (string, string) {
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	argsForCall := fake.setArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *TKGConfigReaderWriter) UnmarshalKey(arg1 string, arg2 interface{}) error {
	fake.unmarshalKeyMutex.Lock()
	ret, specificReturn := fake.unmarshalKeyReturnsOnCall[len(fake.unmarshalKeyArgsForCall)]
	fake.unmarshalKeyArgsForCall = append(fake.unmarshalKeyArgsForCall, struct {
		arg1 string
		arg2 interface{}
	}{arg1, arg2})
	stub := fake.UnmarshalKeyStub
	fakeReturns := fake.unmarshalKeyReturns
	fake.recordInvocation("UnmarshalKey", []interface{}{arg1, arg2})
	fake.unmarshalKeyMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *TKGConfigReaderWriter) UnmarshalKeyCallCount() int {
	fake.unmarshalKeyMutex.RLock()
	defer fake.unmarshalKeyMutex.RUnlock()
	return len(fake.unmarshalKeyArgsForCall)
}

func (fake *TKGConfigReaderWriter) UnmarshalKeyCalls(stub func(string, interface{}) error) {
	fake.unmarshalKeyMutex.Lock()
	defer fake.unmarshalKeyMutex.Unlock()
	fake.UnmarshalKeyStub = stub
}

func (fake *TKGConfigReaderWriter) UnmarshalKeyArgsForCall(i int) (string, interface{}) {
	fake.unmarshalKeyMutex.RLock()
	defer fake.unmarshalKeyMutex.RUnlock()
	argsForCall := fake.unmarshalKeyArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *TKGConfigReaderWriter) UnmarshalKeyReturns(result1 error) {
	fake.unmarshalKeyMutex.Lock()
	defer fake.unmarshalKeyMutex.Unlock()
	fake.UnmarshalKeyStub = nil
	fake.unmarshalKeyReturns = struct {
		result1 error
	}{result1}
}

func (fake *TKGConfigReaderWriter) UnmarshalKeyReturnsOnCall(i int, result1 error) {
	fake.unmarshalKeyMutex.Lock()
	defer fake.unmarshalKeyMutex.Unlock()
	fake.UnmarshalKeyStub = nil
	if fake.unmarshalKeyReturnsOnCall == nil {
		fake.unmarshalKeyReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.unmarshalKeyReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *TKGConfigReaderWriter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getMutex.RLock()
	defer fake.getMutex.RUnlock()
	fake.initMutex.RLock()
	defer fake.initMutex.RUnlock()
	fake.mergeInConfigMutex.RLock()
	defer fake.mergeInConfigMutex.RUnlock()
	fake.setMutex.RLock()
	defer fake.setMutex.RUnlock()
	fake.unmarshalKeyMutex.RLock()
	defer fake.unmarshalKeyMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *TKGConfigReaderWriter) recordInvocation(key string, args []interface{}) {
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

var _ tkgconfigreaderwriter.TKGConfigReaderWriter = new(TKGConfigReaderWriter)

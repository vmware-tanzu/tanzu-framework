// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"
	"time"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Poller struct {
	PollImmediateStub        func(time.Duration, time.Duration, wait.ConditionFunc) error
	pollImmediateMutex       sync.RWMutex
	pollImmediateArgsForCall []struct {
		arg1 time.Duration
		arg2 time.Duration
		arg3 wait.ConditionFunc
	}
	pollImmediateReturns struct {
		result1 error
	}
	pollImmediateReturnsOnCall map[int]struct {
		result1 error
	}
	PollImmediateInfiniteStub        func(time.Duration, wait.ConditionFunc) error
	pollImmediateInfiniteMutex       sync.RWMutex
	pollImmediateInfiniteArgsForCall []struct {
		arg1 time.Duration
		arg2 wait.ConditionFunc
	}
	pollImmediateInfiniteReturns struct {
		result1 error
	}
	pollImmediateInfiniteReturnsOnCall map[int]struct {
		result1 error
	}
	PollImmediateInfiniteWithGetterStub        func(time.Duration, clusterclient.GetterFunc) error
	pollImmediateInfiniteWithGetterMutex       sync.RWMutex
	pollImmediateInfiniteWithGetterArgsForCall []struct {
		arg1 time.Duration
		arg2 clusterclient.GetterFunc
	}
	pollImmediateInfiniteWithGetterReturns struct {
		result1 error
	}
	pollImmediateInfiniteWithGetterReturnsOnCall map[int]struct {
		result1 error
	}
	PollImmediateWithGetterStub        func(time.Duration, time.Duration, clusterclient.GetterFunc) (interface{}, error)
	pollImmediateWithGetterMutex       sync.RWMutex
	pollImmediateWithGetterArgsForCall []struct {
		arg1 time.Duration
		arg2 time.Duration
		arg3 clusterclient.GetterFunc
	}
	pollImmediateWithGetterReturns struct {
		result1 interface{}
		result2 error
	}
	pollImmediateWithGetterReturnsOnCall map[int]struct {
		result1 interface{}
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *Poller) PollImmediate(arg1 time.Duration, arg2 time.Duration, arg3 wait.ConditionFunc) error {
	fake.pollImmediateMutex.Lock()
	ret, specificReturn := fake.pollImmediateReturnsOnCall[len(fake.pollImmediateArgsForCall)]
	fake.pollImmediateArgsForCall = append(fake.pollImmediateArgsForCall, struct {
		arg1 time.Duration
		arg2 time.Duration
		arg3 wait.ConditionFunc
	}{arg1, arg2, arg3})
	fake.recordInvocation("PollImmediate", []interface{}{arg1, arg2, arg3})
	fake.pollImmediateMutex.Unlock()
	if fake.PollImmediateStub != nil {
		return fake.PollImmediateStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.pollImmediateReturns
	return fakeReturns.result1
}

func (fake *Poller) PollImmediateCallCount() int {
	fake.pollImmediateMutex.RLock()
	defer fake.pollImmediateMutex.RUnlock()
	return len(fake.pollImmediateArgsForCall)
}

func (fake *Poller) PollImmediateCalls(stub func(time.Duration, time.Duration, wait.ConditionFunc) error) {
	fake.pollImmediateMutex.Lock()
	defer fake.pollImmediateMutex.Unlock()
	fake.PollImmediateStub = stub
}

func (fake *Poller) PollImmediateArgsForCall(i int) (time.Duration, time.Duration, wait.ConditionFunc) {
	fake.pollImmediateMutex.RLock()
	defer fake.pollImmediateMutex.RUnlock()
	argsForCall := fake.pollImmediateArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *Poller) PollImmediateReturns(result1 error) {
	fake.pollImmediateMutex.Lock()
	defer fake.pollImmediateMutex.Unlock()
	fake.PollImmediateStub = nil
	fake.pollImmediateReturns = struct {
		result1 error
	}{result1}
}

func (fake *Poller) PollImmediateReturnsOnCall(i int, result1 error) {
	fake.pollImmediateMutex.Lock()
	defer fake.pollImmediateMutex.Unlock()
	fake.PollImmediateStub = nil
	if fake.pollImmediateReturnsOnCall == nil {
		fake.pollImmediateReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.pollImmediateReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *Poller) PollImmediateInfinite(arg1 time.Duration, arg2 wait.ConditionFunc) error {
	fake.pollImmediateInfiniteMutex.Lock()
	ret, specificReturn := fake.pollImmediateInfiniteReturnsOnCall[len(fake.pollImmediateInfiniteArgsForCall)]
	fake.pollImmediateInfiniteArgsForCall = append(fake.pollImmediateInfiniteArgsForCall, struct {
		arg1 time.Duration
		arg2 wait.ConditionFunc
	}{arg1, arg2})
	fake.recordInvocation("PollImmediateInfinite", []interface{}{arg1, arg2})
	fake.pollImmediateInfiniteMutex.Unlock()
	if fake.PollImmediateInfiniteStub != nil {
		return fake.PollImmediateInfiniteStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.pollImmediateInfiniteReturns
	return fakeReturns.result1
}

func (fake *Poller) PollImmediateInfiniteCallCount() int {
	fake.pollImmediateInfiniteMutex.RLock()
	defer fake.pollImmediateInfiniteMutex.RUnlock()
	return len(fake.pollImmediateInfiniteArgsForCall)
}

func (fake *Poller) PollImmediateInfiniteCalls(stub func(time.Duration, wait.ConditionFunc) error) {
	fake.pollImmediateInfiniteMutex.Lock()
	defer fake.pollImmediateInfiniteMutex.Unlock()
	fake.PollImmediateInfiniteStub = stub
}

func (fake *Poller) PollImmediateInfiniteArgsForCall(i int) (time.Duration, wait.ConditionFunc) {
	fake.pollImmediateInfiniteMutex.RLock()
	defer fake.pollImmediateInfiniteMutex.RUnlock()
	argsForCall := fake.pollImmediateInfiniteArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *Poller) PollImmediateInfiniteReturns(result1 error) {
	fake.pollImmediateInfiniteMutex.Lock()
	defer fake.pollImmediateInfiniteMutex.Unlock()
	fake.PollImmediateInfiniteStub = nil
	fake.pollImmediateInfiniteReturns = struct {
		result1 error
	}{result1}
}

func (fake *Poller) PollImmediateInfiniteReturnsOnCall(i int, result1 error) {
	fake.pollImmediateInfiniteMutex.Lock()
	defer fake.pollImmediateInfiniteMutex.Unlock()
	fake.PollImmediateInfiniteStub = nil
	if fake.pollImmediateInfiniteReturnsOnCall == nil {
		fake.pollImmediateInfiniteReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.pollImmediateInfiniteReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *Poller) PollImmediateInfiniteWithGetter(arg1 time.Duration, arg2 clusterclient.GetterFunc) error {
	fake.pollImmediateInfiniteWithGetterMutex.Lock()
	ret, specificReturn := fake.pollImmediateInfiniteWithGetterReturnsOnCall[len(fake.pollImmediateInfiniteWithGetterArgsForCall)]
	fake.pollImmediateInfiniteWithGetterArgsForCall = append(fake.pollImmediateInfiniteWithGetterArgsForCall, struct {
		arg1 time.Duration
		arg2 clusterclient.GetterFunc
	}{arg1, arg2})
	fake.recordInvocation("PollImmediateInfiniteWithGetter", []interface{}{arg1, arg2})
	fake.pollImmediateInfiniteWithGetterMutex.Unlock()
	if fake.PollImmediateInfiniteWithGetterStub != nil {
		return fake.PollImmediateInfiniteWithGetterStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.pollImmediateInfiniteWithGetterReturns
	return fakeReturns.result1
}

func (fake *Poller) PollImmediateInfiniteWithGetterCallCount() int {
	fake.pollImmediateInfiniteWithGetterMutex.RLock()
	defer fake.pollImmediateInfiniteWithGetterMutex.RUnlock()
	return len(fake.pollImmediateInfiniteWithGetterArgsForCall)
}

func (fake *Poller) PollImmediateInfiniteWithGetterCalls(stub func(time.Duration, clusterclient.GetterFunc) error) {
	fake.pollImmediateInfiniteWithGetterMutex.Lock()
	defer fake.pollImmediateInfiniteWithGetterMutex.Unlock()
	fake.PollImmediateInfiniteWithGetterStub = stub
}

func (fake *Poller) PollImmediateInfiniteWithGetterArgsForCall(i int) (time.Duration, clusterclient.GetterFunc) {
	fake.pollImmediateInfiniteWithGetterMutex.RLock()
	defer fake.pollImmediateInfiniteWithGetterMutex.RUnlock()
	argsForCall := fake.pollImmediateInfiniteWithGetterArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *Poller) PollImmediateInfiniteWithGetterReturns(result1 error) {
	fake.pollImmediateInfiniteWithGetterMutex.Lock()
	defer fake.pollImmediateInfiniteWithGetterMutex.Unlock()
	fake.PollImmediateInfiniteWithGetterStub = nil
	fake.pollImmediateInfiniteWithGetterReturns = struct {
		result1 error
	}{result1}
}

func (fake *Poller) PollImmediateInfiniteWithGetterReturnsOnCall(i int, result1 error) {
	fake.pollImmediateInfiniteWithGetterMutex.Lock()
	defer fake.pollImmediateInfiniteWithGetterMutex.Unlock()
	fake.PollImmediateInfiniteWithGetterStub = nil
	if fake.pollImmediateInfiniteWithGetterReturnsOnCall == nil {
		fake.pollImmediateInfiniteWithGetterReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.pollImmediateInfiniteWithGetterReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *Poller) PollImmediateWithGetter(arg1 time.Duration, arg2 time.Duration, arg3 clusterclient.GetterFunc) (interface{}, error) {
	fake.pollImmediateWithGetterMutex.Lock()
	ret, specificReturn := fake.pollImmediateWithGetterReturnsOnCall[len(fake.pollImmediateWithGetterArgsForCall)]
	fake.pollImmediateWithGetterArgsForCall = append(fake.pollImmediateWithGetterArgsForCall, struct {
		arg1 time.Duration
		arg2 time.Duration
		arg3 clusterclient.GetterFunc
	}{arg1, arg2, arg3})
	fake.recordInvocation("PollImmediateWithGetter", []interface{}{arg1, arg2, arg3})
	fake.pollImmediateWithGetterMutex.Unlock()
	if fake.PollImmediateWithGetterStub != nil {
		return fake.PollImmediateWithGetterStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.pollImmediateWithGetterReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Poller) PollImmediateWithGetterCallCount() int {
	fake.pollImmediateWithGetterMutex.RLock()
	defer fake.pollImmediateWithGetterMutex.RUnlock()
	return len(fake.pollImmediateWithGetterArgsForCall)
}

func (fake *Poller) PollImmediateWithGetterCalls(stub func(time.Duration, time.Duration, clusterclient.GetterFunc) (interface{}, error)) {
	fake.pollImmediateWithGetterMutex.Lock()
	defer fake.pollImmediateWithGetterMutex.Unlock()
	fake.PollImmediateWithGetterStub = stub
}

func (fake *Poller) PollImmediateWithGetterArgsForCall(i int) (time.Duration, time.Duration, clusterclient.GetterFunc) {
	fake.pollImmediateWithGetterMutex.RLock()
	defer fake.pollImmediateWithGetterMutex.RUnlock()
	argsForCall := fake.pollImmediateWithGetterArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *Poller) PollImmediateWithGetterReturns(result1 interface{}, result2 error) {
	fake.pollImmediateWithGetterMutex.Lock()
	defer fake.pollImmediateWithGetterMutex.Unlock()
	fake.PollImmediateWithGetterStub = nil
	fake.pollImmediateWithGetterReturns = struct {
		result1 interface{}
		result2 error
	}{result1, result2}
}

func (fake *Poller) PollImmediateWithGetterReturnsOnCall(i int, result1 interface{}, result2 error) {
	fake.pollImmediateWithGetterMutex.Lock()
	defer fake.pollImmediateWithGetterMutex.Unlock()
	fake.PollImmediateWithGetterStub = nil
	if fake.pollImmediateWithGetterReturnsOnCall == nil {
		fake.pollImmediateWithGetterReturnsOnCall = make(map[int]struct {
			result1 interface{}
			result2 error
		})
	}
	fake.pollImmediateWithGetterReturnsOnCall[i] = struct {
		result1 interface{}
		result2 error
	}{result1, result2}
}

func (fake *Poller) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.pollImmediateMutex.RLock()
	defer fake.pollImmediateMutex.RUnlock()
	fake.pollImmediateInfiniteMutex.RLock()
	defer fake.pollImmediateInfiniteMutex.RUnlock()
	fake.pollImmediateInfiniteWithGetterMutex.RLock()
	defer fake.pollImmediateInfiniteWithGetterMutex.RUnlock()
	fake.pollImmediateWithGetterMutex.RLock()
	defer fake.pollImmediateWithGetterMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *Poller) recordInvocation(key string, args []interface{}) {
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

var _ clusterclient.Poller = new(Poller)

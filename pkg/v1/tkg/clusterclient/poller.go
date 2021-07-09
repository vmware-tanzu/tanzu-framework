// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterclient

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// GetterFunc is a function which should be used as closure
type GetterFunc func() (interface{}, error)

//go:generate counterfeiter -o ../fakes/poller.go --fake-name Poller . Poller

// Poller implements polling helper functions
type Poller interface {
	// PollImmediate is a wrapper on top of wait.PollImmediate
	// use this function to exercise your condition function with unit tests
	PollImmediate(interval, timeout time.Duration, condition wait.ConditionFunc) error

	// PollImmediateWithGetter is a generic implementation of polling mechanism
	// it will periodically call getterFunc and will return error based on the getterFunc error message if any
	PollImmediateWithGetter(interval, timeout time.Duration, getterFunc GetterFunc) (interface{}, error)

	// PollImmediateInfinite is a wrapper on top of wait.PollImmediateInfinite
	PollImmediateInfinite(interval time.Duration, condition wait.ConditionFunc) error

	// PollImmediateInfiniteWithGetter is a generic implementation of polling mechanism
	// it will periodically call getterFunc and will return error based on the getterFunc error message if any
	PollImmediateInfiniteWithGetter(interval time.Duration, getterFunc GetterFunc) error
}

type pollerProxy struct{}

// NewPoller returns new poller
func NewPoller() Poller {
	return &pollerProxy{}
}

// PollImmediate is a wrapper on top of wait.PollImmediate
// use this function to exercise your condition function with unit tests
func (p *pollerProxy) PollImmediate(interval, timeout time.Duration, condition wait.ConditionFunc) error {
	return wait.PollImmediate(interval, timeout, condition)
}

// PollImmediateInfinite is a wrapper on top of wait.PollImmediateInfinite
func (p *pollerProxy) PollImmediateInfinite(interval time.Duration, condition wait.ConditionFunc) error {
	return wait.PollImmediateInfinite(interval, condition)
}

// PollImmediateWithGetter is a generic implementation of polling mechanism
// it will periodically call getterFunc and will return error based on the getterFunc error message if any
func (p *pollerProxy) PollImmediateWithGetter(interval, timeout time.Duration, getterFunc GetterFunc) (interface{}, error) {
	var result interface{}
	var err error
	pollerFunc := func() (bool, error) {
		result, err = getterFunc()
		if err != nil {
			log.V(6).Info(err.Error() + ", retrying")
			return false, nil
		}
		return true, nil
	}

	errPoll := p.PollImmediate(interval, timeout, pollerFunc)
	if errPoll != nil {
		// note: this function will return actual error which is thrown by getterFunc
		// and will not return error of the PollImmediate call (which is always time-out...)
		return nil, err
	}

	return result, nil
}

// PollImmediateInfiniteWithGetter is a generic implementation of polling mechanism
// it will periodically call getterFunc and will return error based on the getterFunc error message if any
func (p *pollerProxy) PollImmediateInfiniteWithGetter(interval time.Duration, getterFunc GetterFunc) error {
	var timeout interface{}
	var err error
	pollerFunc := func() (bool, error) {
		timeout, err = getterFunc()
		timeoutBool := timeout.(bool)
		if timeoutBool {
			log.V(6).Info(err.Error())
			return false, err
		}
		if err != nil {
			log.V(6).Info(err.Error() + ", retrying")
			return false, nil
		}
		return true, nil
	}

	errPoll := p.PollImmediateInfinite(interval, pollerFunc)
	if errPoll != nil {
		// note: this function will return actual error which is thrown by getterFunc
		// and will not return error of the PollImmediateInfinite call
		return err
	}
	return nil
}

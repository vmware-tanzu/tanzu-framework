// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package conversion implements conversion utilities. The conversion utilities have been copied from cluster API and adapted, see <https://github.com/kubernetes-sigs/cluster-api/blob/f4a76c749d91b7ae654ae4dd838a199939bc6323/util/conversion/conversion.go>.
package conversion

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/google/go-cmp/cmp"
	fuzz "github.com/google/gofuzz"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metafuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

const (
	// DataAnnotation is the annotation that conversion webhooks
	// use to retain the data in case of down-conversion from the hub.
	DataAnnotation = "run.tanzu.vmware.com/conversion-data"
)

// MarshalData stores the source object as json data in the destination object annotations map.
// It ignores the metadata of the source object.
func MarshalData(src, dst metav1.Object) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(src)
	if err != nil {
		return err
	}
	delete(u, "metadata")

	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	annotations := dst.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[DataAnnotation] = string(data)
	dst.SetAnnotations(annotations)
	return nil
}

// UnmarshalData tries to retrieve the data from the annotation and unmarshals it into the object passed as input.
func UnmarshalData(from metav1.Object, to interface{}) (bool, error) {
	annotations := from.GetAnnotations()
	data, ok := annotations[DataAnnotation]
	if !ok {
		return false, nil
	}
	if err := json.Unmarshal([]byte(data), to); err != nil {
		return false, err
	}
	delete(annotations, DataAnnotation)
	from.SetAnnotations(annotations)
	return true, nil
}

// GetFuzzer returns a new fuzzer to be used for testing.
func GetFuzzer(runtimeScheme *runtime.Scheme, funcs ...fuzzer.FuzzerFuncs) *fuzz.Fuzzer {
	funcs = append([]fuzzer.FuzzerFuncs{
		metafuzzer.Funcs,
		func(_ runtimeserializer.CodecFactory) []interface{} {
			return []interface{}{
				// Custom fuzzer for metav1.Time pointers which weren't
				// fuzzed and always resulted in `nil` values.
				// This implementation is somewhat similar to the one provided
				// in the metafuzzer.Funcs.
				func(input *metav1.Time, c fuzz.Continue) {
					if input != nil {
						var sec, nsec uint32
						c.Fuzz(&sec)
						c.Fuzz(&nsec)
						fuzzed := metav1.Unix(int64(sec), int64(nsec)).Rfc3339Copy()
						input.Time = fuzzed.Time
					}
				},
			}
		},
	}, funcs...)
	return fuzzer.FuzzerFor(
		fuzzer.MergeFuzzerFuncs(funcs...),
		rand.NewSource(rand.Int63()), //nolint:gosec
		runtimeserializer.NewCodecFactory(runtimeScheme),
	)
}

// FuzzTestFuncInput contains input parameters
// for the FuzzTestFunc function.
type FuzzTestFuncInput struct {
	Scheme *runtime.Scheme

	Hub              conversion.Hub
	HubAfterMutation func(conversion.Hub)

	Spoke                      conversion.Convertible
	SpokeAfterMutation         func(convertible conversion.Convertible)
	SkipSpokeAnnotationCleanup bool

	FuzzerFuncs []fuzzer.FuzzerFuncs
}

// FuzzTestFunc returns a new testing function to be used in tests to make sure conversions between
// the Hub version of an object and an older version aren't lossy.
func FuzzTestFunc(input *FuzzTestFuncInput) func(*testing.T) {
	if input.Scheme == nil {
		input.Scheme = scheme.Scheme
	}

	return func(t *testing.T) {
		t.Helper()
		t.Run("spoke-hub-spoke", func(t *testing.T) {
			g := gomega.NewWithT(t)
			fuzzr := GetFuzzer(input.Scheme, input.FuzzerFuncs...)

			for i := 0; i < 10000; i++ {
				// Create the spoke and fuzz it
				spokeBefore := input.Spoke.DeepCopyObject().(conversion.Convertible)
				fuzzr.Fuzz(spokeBefore)

				// First convert spoke to hub
				hubCopy := input.Hub.DeepCopyObject().(conversion.Hub)
				g.Expect(spokeBefore.ConvertTo(hubCopy)).To(gomega.Succeed())

				// Convert hub back to spoke and check if the resulting spoke is equal to the spoke before the round trip
				spokeAfter := input.Spoke.DeepCopyObject().(conversion.Convertible)
				g.Expect(spokeAfter.ConvertFrom(hubCopy)).To(gomega.Succeed())

				// Remove data annotation eventually added by ConvertFrom for avoiding data loss in hub-spoke-hub round trips
				// NOTE: There are use case when we want to skip this operation, e.g. if the spoke object does not have ObjectMeta (e.g. kubeadm types).
				if !input.SkipSpokeAnnotationCleanup {
					metaAfter := spokeAfter.(metav1.Object)
					delete(metaAfter.GetAnnotations(), DataAnnotation)
				}

				if input.SpokeAfterMutation != nil {
					input.SpokeAfterMutation(spokeAfter)
				}

				g.Expect(apiequality.Semantic.DeepEqual(spokeBefore, spokeAfter)).To(gomega.BeTrue(), cmp.Diff(spokeBefore, spokeAfter))
			}
		})
		t.Run("hub-spoke-hub", func(t *testing.T) {
			g := gomega.NewWithT(t)
			fuzzr := GetFuzzer(input.Scheme, input.FuzzerFuncs...)

			for i := 0; i < 10000; i++ {
				// Create the hub and fuzz it
				hubBefore := input.Hub.DeepCopyObject().(conversion.Hub)
				fuzzr.Fuzz(hubBefore)

				// First convert hub to spoke
				dstCopy := input.Spoke.DeepCopyObject().(conversion.Convertible)
				g.Expect(dstCopy.ConvertFrom(hubBefore)).To(gomega.Succeed())

				// Convert spoke back to hub and check if the resulting hub is equal to the hub before the round trip
				hubAfter := input.Hub.DeepCopyObject().(conversion.Hub)
				g.Expect(dstCopy.ConvertTo(hubAfter)).To(gomega.Succeed())

				if input.HubAfterMutation != nil {
					input.HubAfterMutation(hubAfter)
				}

				// Remove data annotation eventually added by ConvertFrom for avoiding data loss in hub-spoke-hub round trips
				// NOTE: There are use case when we want to skip this operation, e.g. if the spoke object does not have ObjectMeta (e.g. kubeadm types).
				if !input.SkipSpokeAnnotationCleanup {
					metaAfter := hubAfter.(metav1.Object)
					delete(metaAfter.GetAnnotations(), DataAnnotation)

					metaBefore := hubBefore.(metav1.Object)
					if metaBefore.GetAnnotations() == nil {
						metaBefore.SetAnnotations(map[string]string{})
					}
				}

				g.Expect(apiequality.Semantic.DeepEqual(hubBefore, hubAfter)).To(gomega.BeTrue(), cmp.Diff(hubBefore, hubAfter))
			}
		})
	}
}

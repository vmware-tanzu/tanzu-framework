// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package patchset

import (
	"context"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
)

const (
	v1156 = "1.15.6+vmware.1-tkg.1"
	v116  = "1.16.7+vmware.1-tkg.1"
	v1169 = "1.16.9+vmware.1-tkg.1"
	v117  = "1.17.9+vmware.1-tkg.1"
	v118  = "1.18.2+vmware.1-tkg.1"
)

func TestPatchsetUnit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vlabel Package Unit Tests")
}

var _ = Describe("Patchset", func() {
	var (
		tkr1156 *runv1.TanzuKubernetesRelease
		tkr116  *runv1.TanzuKubernetesRelease
		tkr1169 *runv1.TanzuKubernetesRelease
		tkr117  *runv1.TanzuKubernetesRelease
		tkr118  *runv1.TanzuKubernetesRelease

		c       client.Client
		objects []client.Object
		ps      PatchSet
		tkrs    []*runv1.TanzuKubernetesRelease
	)

	BeforeEach(func() {
		tkr1156 = tkrForVersion(v1156)
		tkr116 = tkrForVersion(v116)
		tkr1169 = tkrForVersion(v1169)
		tkr117 = tkrForVersion(v117)
		tkr118 = tkrForVersion(v118)

		tkrs = []*runv1.TanzuKubernetesRelease{tkr1156, tkr116, tkr1169, tkr117} // tkr118 not included

		objects = []client.Object{tkr1156, tkr116, tkr1169, tkr117, tkr118}
		c = fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(objects...).Build()
		ps = New(c)
	})

	JustBeforeEach(func() {
		for _, tkr := range tkrs {
			ps.Add(tkr)
		}
	})

	When("adding an object", func() {
		var tkrs []*runv1.TanzuKubernetesRelease

		It("should create a patch helper for the object", func() {
			for _, tkr := range tkrs {
				Expect(ps.Objects()[tkr.UID]).To(Equal(tkr))
			}
		})

		When("trying to add an object with the same UID later", func() {
			It("should retain the already existing object", func() {
				tkr118.UID = tkr1156.UID
				ps.Add(tkr118)

				Expect(ps.Objects()[tkr1156.UID]).To(Equal(tkr1156))
			})
		})
	})

	When("applying the patchset", func() {
		BeforeEach(func() {
			ps = New(&countingPatcher{Client: c})
		})

		It("should only patch changed objects", func() {
			changedTKRs := []*runv1.TanzuKubernetesRelease{tkr116, tkr117}
			for _, tkr := range changedTKRs {
				tkr.Labels = labels.Set{"newLabel" + tkr.Name: ""}
			}
			Expect(ps.Apply(context.Background())).To(Succeed())
			Expect(ps.(*patchSet).client.(*countingPatcher).count).To(Equal(len(changedTKRs)))
		})
	})

	When("a patched object is slated for deletion", func() {
		BeforeEach(func() {
			tkr117.DeletionTimestamp = &metav1.Time{Time: time.Now()}
			ps = New(fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(objects...).Build())
		})

		It("should not return a NotFound error", func() {
			conditions.MarkTrue(tkr117, "Whatever")
			err := ps.Apply(context.Background())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("a patched object no longer exists", func() {
		BeforeEach(func() {
			ps = New(fake.NewClientBuilder().WithScheme(newScheme()).WithObjects(nil...).Build())
		})

		It("should return a NotFound error", func() {
			conditions.MarkTrue(tkr117, "Whatever")
			err := ps.Apply(context.Background())
			Expect(err).To(HaveOccurred())
			Expect(isNotFound(err))
		})
	})
})

type countingPatcher struct {
	client.Client
	count int
}

func (p *countingPatcher) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	p.count++
	return p.Client.Patch(ctx, obj, patch, opts...)
}

func newScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = runv1.AddToScheme(s)
	return s
}

func tkrForVersion(version string) *runv1.TanzuKubernetesRelease {
	return &runv1.TanzuKubernetesRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:            strings.ReplaceAll(version, "+", "---"),
			ResourceVersion: "1",
			UID:             uuid.NewUUID(),
			Labels:          labels.Set{"whatever": ""},
		},
		Spec: runv1.TanzuKubernetesReleaseSpec{
			Version: version,
		},
	}
}

// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
)

const tkr1_17 = "v1.17"

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TKR test")
}

var _ = Describe("runGetKubernetesReleases", func() {
	var (
		tkrName       string
		err           error
		clusterClient *fakes.ClusterClient
		tkrs          []runv1alpha3.TanzuKubernetesRelease
		stdOutput     string
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		tkrName = ""
	})

	JustBeforeEach(func() {
		reader, writer, perr := os.Pipe()
		if perr != nil {
			panic(perr)
		}
		stdout := os.Stdout
		stderr := os.Stderr
		defer func() {
			os.Stdout = stdout
			os.Stderr = stderr
		}()
		os.Stdout = writer
		os.Stderr = writer
		gtkr.output = writer
		out := make(chan string)
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go func() {
			var buf bytes.Buffer
			wg.Done()
			_, copyErr := io.Copy(&buf, reader)
			Expect(copyErr).ToNot(HaveOccurred())
			out <- buf.String()
		}()
		wg.Wait()
		err = runGetKubernetesReleases(clusterClient, tkrName)
		writer.Close()
		stdOutput = <-out

	})

	Context("When the GetTanzuKubernetesReleases return error", func() {
		BeforeEach(func() {
			clusterClient.ListResourcesReturns(errors.New("fake TKR error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("When the GetTanzuKubernetesReleases returns the TKRs successfully", func() {
		BeforeEach(func() {
			tkr1 := getFakeTKR("v1.17.17---vmware.1-tkg.2", "v1.17.17+vmware.1", corev1.ConditionFalse)
			tkr2 := getFakeTKR("v1.17.18---vmware.1-tkg.1", "v1.17.18+vmware.1", corev1.ConditionTrue)

			tkrs = []runv1alpha3.TanzuKubernetesRelease{tkr1, tkr2}
			clusterClient.ListResourcesCalls(func(tkrl interface{}, option ...crtclient.ListOption) error {
				tkrList := tkrl.(*runv1alpha3.TanzuKubernetesReleaseList)
				tkrList.Items = append(tkrList.Items, tkrs...)
				return nil
			})
		})
		It("should not return error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(stdOutput).To(ContainSubstring("v1.17.18---vmware.1-tkg.1"))
			Expect(stdOutput).ToNot(ContainSubstring("v1.17.17---vmware.1-tkg.2"))
		})
	})
	Context("When the TKR name prefix is given and listing TKRs returns the TKRs successfully", func() {
		BeforeEach(func() {
			tkrName = tkr1_17
			tkr1 := getFakeTKR("v1.17.17---vmware.1-tkg.2", "v1.17.17+vmware.1", corev1.ConditionFalse)
			tkr2 := getFakeTKR("v1.17.18---vmware.1-tkg.1", "v1.17.18+vmware.1", corev1.ConditionTrue)

			tkrs = []runv1alpha3.TanzuKubernetesRelease{tkr1, tkr2}
			clusterClient.ListResourcesCalls(func(tkrl interface{}, option ...crtclient.ListOption) error {
				tkrList := tkrl.(*runv1alpha3.TanzuKubernetesReleaseList)
				tkrList.Items = append(tkrList.Items, tkrs...)
				return nil
			})
		})
		It("should not return error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(stdOutput).To(ContainSubstring("v1.17.18---vmware.1-tkg.1"))
			Expect(stdOutput).ToNot(ContainSubstring("v1.17.17---vmware.1-tkg.2"))
		})
	})
	Context("When the TKRs are deactivated", func() {
		BeforeEach(func() {
			tkrName = tkr1_17
			tkr1 := getFakeTKR("v1.17.17---vmware.1-tkg.2", "v1.17.17+vmware.1", corev1.ConditionFalse)
			tkr2 := getFakeTKR("v1.17.18---vmware.1-tkg.1", "v1.17.18+vmware.1", corev1.ConditionTrue)
			tkr1.Labels[runv1alpha3.LabelDeactivated] = ""
			tkr2.Labels[runv1alpha3.LabelDeactivated] = ""
			tkrs = []runv1alpha3.TanzuKubernetesRelease{tkr1, tkr2}
			clusterClient.ListResourcesCalls(func(tkrl interface{}, option ...crtclient.ListOption) error {
				tkrList := tkrl.(*runv1alpha3.TanzuKubernetesReleaseList)
				tkrList.Items = append(tkrList.Items, tkrs...)
				return nil
			})
		})
		It("should not return error and the output should not show any TKR", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(stdOutput).ToNot(ContainSubstring("v1.17.18---vmware.1-tkg.1"))
			Expect(stdOutput).ToNot(ContainSubstring("v1.17.17---vmware.1-tkg.2"))
		})
	})

})

func getFakeTKR(tkrName, k8sversion string, compatibleStatus corev1.ConditionStatus) runv1alpha3.TanzuKubernetesRelease {
	tkr := runv1alpha3.TanzuKubernetesRelease{}
	tkr.Name = tkrName
	tkr.Labels = make(map[string]string)
	tkr.Spec.Version = strings.ReplaceAll(tkrName, "---", "+")
	tkr.Spec.Kubernetes.Version = k8sversion
	tkr.Status.Conditions = []clusterv1.Condition{
		{
			Type:   runv1alpha3.ConditionCompatible,
			Status: compatibleStatus,
		},
	}
	return tkr
}

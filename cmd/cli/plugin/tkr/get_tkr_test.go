// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"

	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TKR test")
}

var _ = Describe("runGetKubernetesReleases", func() {
	var (
		tkrName       string
		err           error
		clusterClient *fakes.ClusterClient
		tkrs          []runv1alpha1.TanzuKubernetesRelease
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
			clusterClient.GetTanzuKubernetesReleasesReturns(tkrs, errors.New("fake TKR error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("When the GetTanzuKubernetesReleases returns the TKRs successfully", func() {
		BeforeEach(func() {
			tkr1 := getFakeTKR("v1.17.17---vmware.1-tkg.2", "v1.17.17+vmware.1", corev1.ConditionFalse, "")
			tkr2 := getFakeTKR("v1.17.18---vmware.1-tkg.1", "v1.17.18+vmware.1", corev1.ConditionTrue, "")

			tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr2}
			clusterClient.GetTanzuKubernetesReleasesReturns(tkrs, nil)
		})
		It("should not return error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(stdOutput).To(ContainSubstring("v1.17.18---vmware.1-tkg.1"))
			Expect(stdOutput).ToNot(ContainSubstring("v1.17.17---vmware.1-tkg.2"))
		})
	})

})

func getFakeTKR(tkrName, k8sversion string, compatibleStatus corev1.ConditionStatus, updatesAvailableMsg string) runv1alpha1.TanzuKubernetesRelease {
	tkr := runv1alpha1.TanzuKubernetesRelease{}
	tkr.Name = tkrName
	tkr.Spec.Version = strings.ReplaceAll(tkrName, "---", "+")
	tkr.Spec.KubernetesVersion = k8sversion
	tkr.Status.Conditions = []clusterv1.Condition{
		{
			Type:   clusterv1.ConditionType(runv1alpha1.ConditionCompatible),
			Status: compatibleStatus,
		},
		{
			Type:    clusterv1.ConditionType(runv1alpha1.ConditionUpgradeAvailable),
			Status:  corev1.ConditionTrue,
			Message: updatesAvailableMsg,
		},
	}
	return tkr
}

// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aunum/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"

	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "create cluster test")
}

var _ = Describe("getLatestTKRVersionMatchingTKRPrefix", func() {
	var (
		tkrsWithPrefixMatch []runv1alpha1.TanzuKubernetesRelease
		tkrNamePrefix       string
		latestTKRVersion    string
		err                 error
	)
	const (
		TkrVersionPrefix_v1_17 = "v1.17" //nolint
	)

	JustBeforeEach(func() {
		latestTKRVersion, err = getLatestTKRVersionMatchingTKRPrefix(tkrNamePrefix, tkrsWithPrefixMatch)
	})

	Context("When the list of prefix matched TKRs has highest version TKR as incompatible", func() {
		BeforeEach(func() {
			tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionFalse, "")
			tkr2 := getFakeTKR("v1.17.8---vmware.1-tkg.1", "v1.17.8+vmware.1", corev1.ConditionTrue, "")
			tkr3 := getFakeTKR("v1.17.17---vmware.2-tkg.1", "v1.17.17---vmware.2", corev1.ConditionTrue, "")
			tkr4 := getFakeTKR("v1.17.14---vmware.1-tkg.1-rc.1", "v1.17.14---vmware.1", corev1.ConditionTrue, "")
			tkr5 := getFakeTKR("v1.17.17---vmware.1-tkg.2", "v1.17.17---vmware.1", corev1.ConditionTrue, "")

			tkrsWithPrefixMatch = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}
			tkrNamePrefix = TkrVersionPrefix_v1_17
		})
		It("should return the next latest TKR version that is compatible", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(latestTKRVersion).To(Equal("v1.17.17+vmware.2-tkg.1"))
		})
	})
	Context("When the list of prefix matched TKRs has multiple latest TKRs", func() {
		BeforeEach(func() {
			tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18---vmware.1", corev1.ConditionTrue, "")
			tkr2 := getFakeTKR("v1.17.18---vmware.2-tkg.1-rc.1", "v1.17.18---vmware.2-tkg.1", corev1.ConditionTrue, "")
			tkr3 := getFakeTKR("v1.17.15---vmware.1-tkg.1", "v1.17.15---vmware.1", corev1.ConditionTrue, "")
			tkr4 := getFakeTKR("v1.17.18---vmware.2-tkg.1-zlatest1", "1.17.18---vmware.2-tkg.1", corev1.ConditionTrue, "")
			tkrsWithPrefixMatch = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr2, tkr3, tkr4}
			tkrNamePrefix = TkrVersionPrefix_v1_17
		})
		It("should return error ", func() {
			Expect(err).To(HaveOccurred())
			errString := "found multiple TKrs [v1.17.18---vmware.2-tkg.1-zlatest1 v1.17.18---vmware.2-tkg.1-rc.1] matching the criteria"
			Expect(err.Error()).To(ContainSubstring(errString))
		})
	})
	Context("When the list of prefix matched TKRs has no compatible TKRs", func() {
		BeforeEach(func() {
			tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18---vmware.1", corev1.ConditionFalse, "")
			tkr2 := getFakeTKR("v1.17.8---vmware.1-tkg.1", "v1.17.8---vmware.1", corev1.ConditionFalse, "")
			tkr3 := getFakeTKR("v1.17.17---vmware.2-tkg.1", "v1.17.17---vmware.2", corev1.ConditionFalse, "")
			tkrsWithPrefixMatch = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr3, tkr2}
			tkrNamePrefix = TkrVersionPrefix_v1_17
		})
		It("should return error as there is no single compatible TKR", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not find a matching compatible Tanzu Kubernetes release for name \"v1.17\""))
		})
	})
})

func Test_CreateClusterCommand(t *testing.T) {
	for _, test := range []struct {
		testcase    string
		stringMatch []string
		preConfig   func()
	}{
		{
			testcase:    "When default tanzu config file does not exist",
			stringMatch: []string{"kind: Cluster", "name: test-cluster"},
		},
		{
			testcase:    "When default tanzu config file exists but current server is not configured",
			stringMatch: []string{"kind: Cluster", "name: test-cluster"},
			preConfig: func() {
				configureTanzuConfig("./testdata/tanzuconfig/config1.yaml")
			},
		},
		{
			testcase:    "When default tanzu config file exists and current server is configured",
			stringMatch: []string{"kind: Cluster", "name: test-cluster"},
			preConfig: func() {
				configureTanzuConfig("./testdata/tanzuconfig/config2.yaml")
			},
		},
	} {
		t.Run(test.testcase, func(t *testing.T) {
			defer configureHomeDirectory()()
			out := captureStdoutStderr(runCreateClusterCmd)
			for _, str := range test.stringMatch {
				if !strings.Contains(out, str) {
					t.Fatalf("expected \"%s\" to contain \"%s\"", out, str)
				}
			}
		})
	}
}

func runCreateClusterCmd() {
	cmd := createClusterCmd
	cmd.SetArgs([]string{"test-cluster", "-i", "docker:v0.3.16", "-p", "dev", "-d"})
	_ = cmd.Execute()
}

func captureStdoutStderr(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func configureHomeDirectory() func() {
	fs := new(afero.MemMapFs)
	f, err := afero.TempDir(fs, "", "CreateClusterTest")
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("HOME", f)
	return func() {
		os.Unsetenv("HOME")
	}
}

func configureTanzuConfig(file string) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	tanzuDir := filepath.Join(dirname, ".tanzu")
	err = os.Mkdir(tanzuDir, 0600)
	if err != nil {
		log.Fatal(err)
	}
	err = utils.CopyFile(file, filepath.Join(tanzuDir, "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}
}

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

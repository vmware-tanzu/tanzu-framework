// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	k8sv1172    = "v1.17.2+vmware.2"
	k8sv1174    = "v1.17.4+vmware.2"
	k8sv1182    = "v1.18.2+vmware.2"
	k8sv1191    = "v1.19.1+vmware.2"
	vABC        = "vA.B.C"
	fakePathABC = "gcr.io/fakepath/tkg/kind/node:vA.B.C"
)

func Test_ReplaceVersionInDockerImage(t *testing.T) {
	preimage := "registry.tkg.vmware.run/kind/node:vX.Y.Z"
	newVersion := vABC
	expected := "registry.tkg.vmware.run/kind/node:vA.B.C"
	actual, err := ReplaceVersionInDockerImage(preimage, newVersion)
	if err != nil {
		t.Fatalf("Error replacing version in valid docker image: %s", err)
	}

	if actual != expected {
		t.Errorf("Incorrect replaced image string. Expected: %s, actual: %s", expected, actual)
	}
}

func Test_ReplaceVersionInDockerImageLongPath(t *testing.T) {
	preimage := "gcr.io/fakepath/tkg/kind/node:vX.Y.Z"
	newVersion := vABC
	expected := fakePathABC
	actual, err := ReplaceVersionInDockerImage(preimage, newVersion)
	if err != nil {
		t.Fatalf("Error replacing version in valid docker image: %s", err)
	}

	if actual != expected {
		t.Errorf("Incorrect replaced image string. Expected: %s, actual: %s", expected, actual)
	}
}

func Test_ReplaceVersionInDockerImageTwoColons(t *testing.T) {
	preimage := "gcr.io:port/fakepath/tkg/kind/node:vX.Y.Z"
	newVersion := vABC
	expected := "gcr.io:port/fakepath/tkg/kind/node:vA.B.C"
	actual, err := ReplaceVersionInDockerImage(preimage, newVersion)
	if err != nil {
		t.Fatalf("Error replacing version in valid docker image: %s", err)
	}

	if actual != expected {
		t.Errorf("Incorrect replaced image string. Expected: %s, actual: %s", expected, actual)
	}
}

func Test_ReplaceVersionInDockerImageNoTag(t *testing.T) {
	preimage := "gcr.io/fakepath/tkg/kind/node"
	newVersion := vABC
	actual, err := ReplaceVersionInDockerImage(preimage, newVersion)
	expected := fakePathABC
	if err != nil {
		t.Fatalf("Error replacing version in valid docker image: %s", err)
	}

	if actual != expected {
		t.Errorf("Incorrect replaced image string. Expected: %s, actual: %s", expected, actual)
	}
}

func Test_ReplaceVersionInDockerImageNoSlashes(t *testing.T) {
	preimage := "gcr.io"
	newVersion := vABC
	_, err := ReplaceVersionInDockerImage(preimage, newVersion)
	if err == nil {
		t.Errorf("There should have been an error when getting an image with no slashes but there was none")
	}
}

func Test_CheckKubernetesUpgradeCompatibilityWithSkippingMinorVersions(t *testing.T) {
	fromVersion := "v1.17.1+vmware.2"
	toVersion := k8sv1191

	upgradable := CheckKubernetesUpgradeCompatibility(fromVersion, toVersion)
	if upgradable {
		t.Errorf("Upgrading between skipping minor versions should not be allowed")
	}
}

func Test_CheckKubernetesUpgradeCompatibilityWithDifferentMajorVersions(t *testing.T) {
	fromVersion := "v0.19.1+vmware.2"
	toVersion := k8sv1191

	upgradable := CheckKubernetesUpgradeCompatibility(fromVersion, toVersion)
	if upgradable {
		t.Errorf("Upgrading between different major versions should not be allowed")
	}
}

func Test_CheckKubernetesUpgradeCompatibilityWithoutContinuousMinorVersions(t *testing.T) {
	fromVersion := "v1.18.1+vmware.2"
	toVersion := k8sv1191

	upgradable := CheckKubernetesUpgradeCompatibility(fromVersion, toVersion)
	if !upgradable {
		t.Errorf("Upgrading between continuous minor versions should be allowed")
	}
}

func Test_CheckKubernetesUpgradeCompatibilityWithIncreasingPatchVersions(t *testing.T) {
	fromVersion := k8sv1191
	toVersion := "v1.19.3+vmware.2"

	upgradable := CheckKubernetesUpgradeCompatibility(fromVersion, toVersion)
	if !upgradable {
		t.Errorf("Upgrading between the same minor versions but increasing patch version should be allowed")
	}
}

func Test_CheckKubernetesUpgradeCompatibilityWithBackwardMinorVersions(t *testing.T) {
	fromVersion := k8sv1191
	toVersion := k8sv1182

	upgradable := CheckKubernetesUpgradeCompatibility(fromVersion, toVersion)
	if upgradable {
		t.Errorf("Upgrading between backward minor versions should not be allowed")
	}
}

func Test_CheckKubernetesUpgradeCompatibilityWithBackwardPatchVersions(t *testing.T) {
	fromVersion := "v1.19.2+vmware.2"
	toVersion := k8sv1191

	upgradable := CheckKubernetesUpgradeCompatibility(fromVersion, toVersion)
	if upgradable {
		t.Errorf("Upgrading between backward patch versions should not be allowed")
	}
}

var _ = Describe("CompareVMwareVersionStrings", func() {
	var (
		fromVersion   string
		toVersion     string
		err           error
		compareResult int
	)
	JustBeforeEach(func() {
		compareResult, err = CompareVMwareVersionStrings(fromVersion, toVersion)
	})

	Context("When minor version in multiple digits, From: v1.17.4+vmware.2 To: v1.17.33+vmware.2", func() {
		BeforeEach(func() {
			fromVersion = k8sv1174
			toVersion = "v1.17.33+vmware.2"
		})
		It("expect compareResult < 0", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(compareResult < 0).To(Equal(true))
		})
	})

	Context("When version are same v2=v1", func() {
		BeforeEach(func() {
			fromVersion = k8sv1174
			toVersion = k8sv1174
		})
		It("expect compareResult == 0", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(compareResult == 0).To(Equal(true))
		})
	})

	Context("When patch version is greater, v2.patch > v1.patch", func() {
		BeforeEach(func() {
			fromVersion = k8sv1174
			toVersion = "v1.17.5+vmware.2"
		})
		It("expect compareResult < 0", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(compareResult < 0).To(Equal(true))
		})
	})

	Context("When patch version is lesser, v2.patch < v1.patch", func() {
		BeforeEach(func() {
			fromVersion = k8sv1174
			toVersion = "v1.17.3+vmware.2"
		})
		It("expect compareResult > 0", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(compareResult > 0).To(Equal(true))
		})
	})

	Context("When minor version is greater, v2.minor > v1.minor", func() {
		BeforeEach(func() {
			fromVersion = k8sv1172
			toVersion = k8sv1182
		})
		It("expect compareResult < 0", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(compareResult < 0).To(Equal(true))
		})
	})

	Context("When minor version is lesser, v2.minor < v1.minor", func() {
		BeforeEach(func() {
			fromVersion = k8sv1182
			toVersion = k8sv1172
		})
		It("expect compareResult > 0", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(compareResult > 0).To(Equal(true))
		})
	})

	Context("When vmware build version is greater, v2.vmwarebuild > v1.vmwarebuild", func() {
		BeforeEach(func() {
			fromVersion = k8sv1172
			toVersion = "v1.17.2+vmware.3"
		})
		It("expect compareResult < 0", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(compareResult < 0).To(Equal(true))
		})
	})

	Context("When vmware build version is lesser, v2.vmwarebuild < v1.vmwarebuild", func() {
		BeforeEach(func() {
			fromVersion = k8sv1172
			toVersion = "v1.17.2+vmware.1"
		})
		It("expect compareResult > 0", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(compareResult > 0).To(Equal(true))
		})
	})

	Context("When vmware build version is greater, v2.vmwarebuild < v1.vmwarebuild", func() {
		BeforeEach(func() {
			fromVersion = k8sv1172
			toVersion = "v1.17.2+vmware.11"
		})
		It("expect compareResult < 0", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(compareResult < 0).To(Equal(true))
		})
	})
})

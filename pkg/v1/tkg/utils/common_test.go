// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	k8sv1172           = "v1.17.2+vmware.2"
	k8sv1174           = "v1.17.4+vmware.2"
	k8sv1182           = "v1.18.2+vmware.2"
	k8sv1191           = "v1.19.1+vmware.2"
	vABC               = "vA.B.C"
	fakePathABC        = "gcr.io/fakepath/tkg/kind/node:vA.B.C"
	tkrNameSample      = "tkr---version"
	tkrVersionSample   = "tkrversion"
	expectedTKRVersion = "tkr+version"
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

func Test_ReplaceVersionInDockerImage_EmptyParam(t *testing.T) {
	preimage := ""
	_, err := ReplaceVersionInDockerImage(preimage, "")
	if err == nil {
		t.Error("Enpty image string should return an error")
	}
}

func Test_ReplaceVersionInDockerImage_TagWithExcessColons(t *testing.T) {
	preimage := "registry.tkg.vmware.run/kind/node:vX.Y.Z:W"
	_, err := ReplaceVersionInDockerImage(preimage, "")

	if err == nil {
		t.Error("Tag with multiple colons should not be accepted")
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

func Test_ContainsString_EmptyArray(t *testing.T) {
	array := []string{}

	contains := ContainsString(array, "")
	if contains {
		t.Error("ContainsString should return false with an empty input array")
	}
}

func Test_ContainsString_Success(t *testing.T) {
	array := []string{"one fish", "two fish", "red fish", "blue fish"}

	contains := ContainsString(array, "blue fish")
	if !contains {
		t.Error("ContainsString should have found blue fish in the array of strings")
	}
}

func Test_ReplaceSpecialChars_Success(t *testing.T) {
	principal := "this.is+a+test.string"
	expected := "this-is-a-test-string"

	if actual := ReplaceSpecialChars(principal); actual != expected {
		t.Errorf("ReplaceSpecialChars did not replace the special chars, output was %s", actual)
	}
}

func Test_ToSnakeCase_Success(t *testing.T) {
	camelCase := "camelCase"
	expected := "CAMEL_CASE"

	if actual := ToSnakeCase(camelCase); actual != expected {
		t.Error("Did not successfully convert to snake case")
	}
}

func Test_IsValidURL_Success(t *testing.T) {
	testURL := "https://vmware.com"

	if !IsValidURL(testURL) {
		t.Errorf("Valid URL %s was marked as invalid", testURL)
	}
}

func Test_GetTkrNameFromTkrVersion_Success(t *testing.T) {
	tkrVersion := expectedTKRVersion
	expected := tkrNameSample
	if actual := GetTkrNameFromTkrVersion(tkrVersion); actual != expected {
		t.Errorf("Expected tkr name %s, got %s", expected, actual)
	}
}

func Test_GetTkrNameFromTkrVersion_NotTkrVersion(t *testing.T) {
	tkrVersion := tkrVersionSample
	if actual := GetTkrNameFromTkrVersion(tkrVersion); actual != tkrVersion {
		t.Errorf("The actual output %s, did not match the input %s", actual, tkrVersion)
	}
}

func Test_GetTKRVersionFromTKRName_Success(t *testing.T) {
	tkrName := tkrNameSample
	expected := expectedTKRVersion
	if actual := GetTKRVersionFromTKRName(tkrName); actual != expected {
		t.Errorf("Expected version %s, got %s", expected, actual)
	}
}

func Test_GetTKRVersionFromTKRName_NotTkrName(t *testing.T) {
	tkrName := tkrVersionSample
	if actual := GetTKRVersionFromTKRName(tkrName); actual != tkrName {
		t.Errorf("The actual output %s, did not match the input %s", actual, tkrName)
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

	Context("When fromVersion is invalid", func() {
		BeforeEach(func() {
			fromVersion = "not a version"
			toVersion = k8sv1174
		})
		It("should fail to parse", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("invalid version string"))
		})
	})

	Context("When toVersion is invalid", func() {
		BeforeEach(func() {
			fromVersion = k8sv1174
			toVersion = "incorrect version"
		})
		It("hould fail to parse", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("invalid version string"))
		})
	})

	Context("When fromVersion does not contain valid Semver", func() {
		BeforeEach(func() {
			fromVersion = "not.a.version+vmware.2"
			toVersion = k8sv1174
		})
		It("should fail to validate semver", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When toVersion does not contain valid Semver", func() {
		BeforeEach(func() {
			toVersion = "not.a.version+vmware.3"
			fromVersion = k8sv1174
		})
		It("should fail to validate semver", func() {
			Expect(err).To(HaveOccurred())
		})
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

	Context("When fromVersion vmware version is not an int", func() {
		BeforeEach(func() {
			toVersion = k8sv1174
			fromVersion = "v1.17.4+vmware.x"
		})
		It("should fail to validate semver", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("invalid version string"))
		})
	})

	Context("When toVersion vmware version is not an int", func() {
		BeforeEach(func() {
			toVersion = "v1.17.4+vmware.y"
			fromVersion = k8sv1174
		})
		It("should fail to validate semver", func() {
			Expect(compareResult).To(Equal(0))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("invalid version string"))
		})
	})
})

var _ = Describe("GenerateRandomID", func() {
	var (
		length                int
		excludeCapitalLetters bool
		actual                string
	)
	JustBeforeEach(func() {
		actual = GenerateRandomID(length, excludeCapitalLetters)
	})

	Context("When length is zero", func() {
		It("Should return an empty string", func() {
			Expect(actual).To(Equal(""))
		})
	})

	Context("When excludeCapitalLetters is true", func() {
		BeforeEach(func() {
			length = 32
			excludeCapitalLetters = true
		})
		It("result should not include capital letters", func() {
			Expect(actual).To(Equal(strings.ToLower(actual)))
		})
	})
})

var _ = Describe("DivideVPCCidr", func() {
	var (
		cidrStr      string
		extendedBits int
		numSubsets   int
		actual       []string
		err          error
	)
	JustBeforeEach(func() {
		actual, err = DivideVPCCidr(cidrStr, extendedBits, numSubsets)
	})

	Context("Given a valid cidr and subsets", func() {
		BeforeEach(func() {
			cidrStr = "192.168.1.1/24"
			extendedBits = 1
			numSubsets = 2
		})
		It("should succeed", func() {
			Expect(err).NotTo(HaveOccurred())
			expected := []string{"192.168.1.0/25", "192.168.1.128/25"}
			Expect(actual).To(Equal(expected))
		})
	})
})

var _ = Describe("CompareMajorMinorPatchVersion", func() {
	var (
		v1     string
		v2     string
		result bool
	)
	JustBeforeEach(func() {
		result = CompareMajorMinorPatchVersion(v1, v2)
	})

	Context("When v1=v1.4.2 and v2=v1.4.2", func() {
		BeforeEach(func() {
			v1 = "v1.4.2" // nolint:goconst
			v2 = "v1.4.2"
		})
		It("should return true", func() {
			Expect(result).To(BeTrue())
		})
	})
	Context("When v1=v1.4.2+vmware.1 and v2=v1.4.2", func() {
		BeforeEach(func() {
			v1 = "v1.4.2+vmware.1"
			v2 = "v1.4.2"
		})
		It("should return true", func() {
			Expect(result).To(BeTrue())
		})
	})
	Context("When v1=v1.4.2-latest and v2=v1.4.2", func() {
		BeforeEach(func() {
			v1 = "v1.4.2-latest" // nolint:goconst
			v2 = "v1.4.2"
		})
		It("should return true", func() {
			Expect(result).To(BeTrue())
		})
	})
	Context("When v1=v1.4.2-latest and v2=v1.4.2+vmware", func() {
		BeforeEach(func() {
			v1 = "v1.4.2-latest"
			v2 = "v1.4.2+vmware"
		})
		It("should return true", func() {
			Expect(result).To(BeTrue())
		})
	})
	Context("When v1=v1.4.2-latest and v2=v1.4.0+vmware", func() {
		BeforeEach(func() {
			v1 = "v1.4.2-latest"
			v2 = "v1.4.0+vmware"
		})
		It("should return false", func() {
			Expect(result).To(BeFalse())
		})
	})
	Context("When v1=v1.4.2-latest and v2=''", func() {
		BeforeEach(func() {
			v1 = "v1.4.2-latest"
			v2 = ""
		})
		It("should return false", func() {
			Expect(result).To(BeFalse())
		})
	})
	Context("When both versions are empty", func() {
		BeforeEach(func() {
			v1 = ""
			v2 = ""
		})
		It("should return false", func() {
			Expect(result).To(BeFalse())
		})
	})
})

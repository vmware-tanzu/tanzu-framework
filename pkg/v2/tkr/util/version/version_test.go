// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/labels"
)

func TestUtilVersionUnit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "util/version Unit Tests")
}

var _ = Describe("ParseSemantic()", func() {
	When("given an invalid semver string", func() {
		It("should return an error", func() {
			v, err := ParseSemantic("covfefe")
			Expect(err).To(HaveOccurred())
			Expect(v).To(BeNil())
		})
	})
	When("given a valid semver string", func() {
		It("should parse it and return a non-nil *Version", func() {
			v, err := ParseSemantic("1.16.2+vmware.1-tkg.1.cafebabe")
			Expect(err).ToNot(HaveOccurred())
			Expect(v).ToNot(BeNil())
		})
	})
})

var _ = Describe("Version", func() {
	Context("v.LessThan(other)", func() {
		var v, other *Version
		BeforeEach(func() {
			v, other = nil, nil
		})

		When("v is nil, other is nil", func() {
			It("should return false", func() {
				Expect(v.LessThan(other)).To(BeFalse())
			})
		})

		When("v is nil, other is not nil", func() {
			BeforeEach(func() {
				other, _ = ParseSemantic("1.16.2+vmware.1-tkg.1.cafebabe")
				Expect(other).ToNot(BeNil())
			})

			It("should return true", func() {
				Expect(v.LessThan(other)).To(BeTrue())
			})
		})

		When("v is not nil, other is nil", func() {
			BeforeEach(func() {
				v, _ = ParseSemantic("1.16.2+vmware.1-tkg.1.cafebabe")
				Expect(v).ToNot(BeNil())
			})

			It("should return false", func() {
				Expect(v.LessThan(nil)).To(BeFalse())
			})
		})

		When("v is not nil, other is not nil", func() {
			When("the v version part is lower", func() {
				BeforeEach(func() {
					v, _ = ParseSemantic("1.16.1+vmware.1-tkg.1.cafebabe")
					other, _ = ParseSemantic("1.16.2+vmware.1-tkg.1.cafebabe")
					Expect(v.version).ToNot(BeNil())
					Expect(other.version).ToNot(BeNil())
					Expect(v.version.LessThan(other.version)).To(BeTrue())
				})

				It("should return true", func() {
					Expect(v.LessThan(other)).To(BeTrue())
				})
			})

			When("the v version part is higher", func() {
				BeforeEach(func() {
					v, _ = ParseSemantic("1.16.2+vmware.1-tkg.1.cafebabe")
					other, _ = ParseSemantic("1.16.1+vmware.1-tkg.1.cafebabe")
					Expect(v.version).ToNot(BeNil())
					Expect(other.version).ToNot(BeNil())
					Expect(other.version.LessThan(v.version)).To(BeTrue())
				})

				It("should return false", func() {
					Expect(v.LessThan(other)).To(BeFalse())
				})
			})

			When("the v version part is the same", func() {
				When("the v build metadata part is lower", func() {
					BeforeEach(func() {
						v, _ = ParseSemantic("1.16.2+vmware.1-tkg.2.ab123fed")
						other, _ = ParseSemantic("1.16.2+vmware.2-tkg.2.ab123fed")
						Expect(v.version).ToNot(BeNil())
						Expect(other.version).ToNot(BeNil())
						Expect(v.version.LessThan(other.version)).To(BeFalse())
						Expect(other.version.LessThan(v.version)).To(BeFalse())
					})

					It("should return true", func() {
						Expect(v.LessThan(other)).To(BeTrue())
					})
				})

				When("the v build metadata part is lower, comparing strings", func() {
					BeforeEach(func() {
						v, _ = ParseSemantic("1.16.2+vmware.2-tkg.2.covfefe")
						other, _ = ParseSemantic("1.16.2+vmware.2-tkg.2.covfefe1")
						Expect(v.version).ToNot(BeNil())
						Expect(other.version).ToNot(BeNil())
						Expect(v.version.LessThan(other.version)).To(BeFalse())
						Expect(other.version.LessThan(v.version)).To(BeFalse())
					})

					It("should return true", func() {
						Expect(v.LessThan(other)).To(BeTrue())
					})
				})
			})
		})
	})
})

var _ = Describe("BuildMetadata", func() {
	bms1 := "vmware.1-tkg.2.ab123fed"
	bms2 := "vmware.2-tkg.2.fed98bef"

	Context("ParseBuildMetadata()", func() {
		It("should parse the string into a slice of parts and separators", func() {
			Expect(ParseBuildMetadata(bms1)).To(Equal(BuildMetadata{"vmware", ".", "1", "-", "tkg", ".", "2", ".", "ab123fed"}))
			Expect(ParseBuildMetadata(bms2)).To(Equal(BuildMetadata{"vmware", ".", "2", "-", "tkg", ".", "2", ".", "fed98bef"}))
		})
	})

	Context(".LessThan()", func() {
		var x, y BuildMetadata
		BeforeEach(func() {
			x, y = nil, nil
		})

		It(fmt.Sprintf("'%s' should be less than '%s'", bms1, bms2), func() {
			bm1 := ParseBuildMetadata(bms1)
			bm2 := ParseBuildMetadata(bms2)

			Expect(bm1.LessThan(bm2)).To(BeTrue())
		})

		It("x is empty, y is empty, should return false", func() {
			Expect(x.LessThan(y)).To(BeFalse())
		})
		It("x is not empty, y is empty, should return false", func() {
			x = BuildMetadata{"vmware", ".", "1"}
			Expect(x.LessThan(y)).To(BeFalse())
		})
		It("x is empty, y is not empty, should return true", func() {
			y = BuildMetadata{"vmware", ".", "1"}
			Expect(x.LessThan(y)).To(BeTrue())
		})

		It("compare - < .", func() {
			Expect(strings.Compare("-", ".") < 0).To(BeTrue())
			Expect(strings.Compare("-", "123") < 0).To(BeTrue())
			Expect(strings.Compare(".", "123") < 0).To(BeTrue())
			Expect(strings.Compare("-", "abc") < 0).To(BeTrue())
			Expect(strings.Compare(".", "abc") < 0).To(BeTrue())
			Expect(strings.Compare(".", "-") < 0).To(BeFalse())
		})

		When("x is not empty, y is not empty", func() {
			When("x is a prefix of y", func() {
				BeforeEach(func() {
					x = BuildMetadata{"vmware", "."}
					y = BuildMetadata{"vmware", ".", "1"}
				})

				It("should return true", func() {
					Expect(x.LessThan(y)).To(BeTrue())
				})
			})
		})
	})
})

var _ = Describe("Prefixes()", func() {
	const vLabel = "v1.17.9---vmware.2-tkg.3"

	expectedPrefixLabels := labels.Set{
		"v1.17.9---vmware.2-tkg.3": "",
		"v1.17.9---vmware.2-tkg":   "",
		"v1.17.9---vmware.2":       "",
		"v1.17.9---vmware":         "",
		"v1.17.9":                  "",
		"v1.17":                    "",
		"v1":                       "",
	}

	It("should return all version prefixes of a version label", func() {
		Expect(Prefixes(vLabel)).To(Equal(expectedPrefixLabels))
	})
})

var _ = Describe("Label()", func() {
	It("Produces a label string for a version", func() {
		v := "1.22.0+vmware.1-tkg.3"
		Expect(Label(v)).To(Equal("v1.22.0---vmware.1-tkg.3"))
	})
})

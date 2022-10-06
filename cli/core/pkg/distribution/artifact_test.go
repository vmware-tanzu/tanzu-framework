// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package distribution

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestArtifact(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Distribution Suite")
}

var _ = Describe("Unit tests for distribution", func() {
	sampleArtifacts := Artifacts{}
	artifact1 := Artifact{
		Image:  "image1",
		URI:    "uri1",
		Digest: "digest1",
		OS:     "ubuntu",
		Arch:   "amd64",
	}

	artifact2 := Artifact{
		Image:  "image2",
		URI:    "uri2",
		Digest: "digest2",
		OS:     "ubuntu",
		Arch:   "amd64",
	}

	artifact3 := Artifact{
		Image:  "",
		URI:    "",
		Digest: "",
		OS:     "windows",
		Arch:   "amd64",
	}

	emptyImageArtifact := Artifact{
		Image:  "",
		URI:    "/tmp/tmp-local-artifact",
		Digest: "",
		OS:     "windows",
		Arch:   "amd64",
	}

	artifactList := ArtifactList{artifact1, artifact2, artifact3}
	testFetchArtifactList := ArtifactList{emptyImageArtifact}
	sampleArtifacts["1.0.0"] = artifactList
	sampleArtifacts["2.0.0"] = testFetchArtifactList

	var _ = Context("tests for the GetArtifact function", func() {
		var _ = It("test happy path", func() {
			artifact, err := sampleArtifacts.GetArtifact("1.0.0", "ubuntu", "amd64")
			expectedArtifact := artifact1

			Expect(err).ToNot(HaveOccurred())
			Expect(artifact).To(Equal(expectedArtifact))
			Expect(artifact.Image).To(Equal(expectedArtifact.Image))
			Expect(artifact.URI).To(Equal(expectedArtifact.URI))
		})

		var _ = It("when version does not exist in artifact keys", func() {
			artifact, err := sampleArtifacts.GetArtifact("2.0.0", "", "")
			expectedArtifact := Artifact{}
			Expect(err).To(HaveOccurred())
			Expect(artifact).To(Equal(expectedArtifact))
		})

		var _ = It("when artifactMap is nil", func() {
			var nilArtifactMap Artifacts
			expectedArtifact := Artifact{}

			artifact, err := nilArtifactMap.GetArtifact("1.0.0", "ubuntu", "amd64")
			Expect(err).To(HaveOccurred())
			Expect(artifact).To(Equal(expectedArtifact))
		})

	})

	var _ = Context("Unit tests for the Fetch function", func() {
		var tmpFileName string
		BeforeEach(func() {
			file, err := os.CreateTemp("/tmp", "local-artifact")
			Expect(err).ToNot(HaveOccurred())
			tmpFileName = file.Name()
		})

		AfterEach(func() {
			err := os.Remove(tmpFileName)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Image and URI are nil", func() {
			_, err := sampleArtifacts.Fetch("1.0.0", "windows", "amd64")
			Expect(err).To(HaveOccurred())
		})

		It("Image is nil", func() {
			testFetchArtifactList[0].URI = tmpFileName
			artifact, err := sampleArtifacts.Fetch("2.0.0", "windows", "amd64")
			Expect(err).ToNot(HaveOccurred())
			Expect(artifact).ToNot(BeNil())
		})
	})

	var _ = Context("Unit tests for getting the digest", func() {
		var _ = It("test happy path", func() {
			digest, err := sampleArtifacts.GetDigest("1.0.0", "ubuntu", "amd64")
			expectedDigest := artifact1.Digest
			Expect(err).ToNot(HaveOccurred())
			Expect(digest).To(Equal(expectedDigest))
		})

		var _ = It("test error in getting artifact", func() {
			_, err := sampleArtifacts.GetDigest("1.0.0", "linux", "amd64")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Unit tests for Describe Artifact", func() {
		It("Test Happy Path", func() {
			artifact, err := sampleArtifacts.DescribeArtifact("1.0.0", "ubuntu", "amd64")
			Expect(err).ToNot(HaveOccurred())
			Expect(artifact).To(Equal(artifact1))
		})

		It("Test with nil artifactMap", func() {
			var nilArtifactMap Artifacts
			expectedArtifact := Artifact{}

			artifact, err := nilArtifactMap.DescribeArtifact("1.0.0", "ubuntu", "amd64")
			Expect(err).To(HaveOccurred())
			Expect(artifact).To(Equal(expectedArtifact))
		})
	})
})

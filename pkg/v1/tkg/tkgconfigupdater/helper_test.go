// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"archive/zip"
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	sampleFilePath    string
	sampleFilePath2   string
	destDir           string
	baseDir           string
	incorrectFileName string
)

var _ = Describe("Tests while unzipping a file", func() {
	BeforeEach(func() {
		sampleFilePath = "/tmp/baz.zip"
		sampleFilePath2 = "/tmp/foobar.zip"
		contentFileName := "foo.txt"
		incorrectFileName = "../foobar.txt"
		baseDir, err = os.MkdirTemp("/tmp", "test-zip")
		if err != nil {
			return
		}

		err = setupSampleZipFile(sampleFilePath, contentFileName)
		if err != nil {
			return
		}
	})
	Context("Validating destDir path", func() {
		AfterEach(func() {
			os.RemoveAll(baseDir)
			os.Remove(sampleFilePath)
			os.Remove(sampleFilePath2)
			os.Remove("/tmp/foobar.txt")
		})
		It("when content file path is valid", func() {
			destDir = "/tmp/foo"
			err := unzip(sampleFilePath, destDir)
			Expect(err).To(BeNil())

			os.RemoveAll(destDir)
		})
		It("when content file path is invalid", func() {
			err := setupSampleZipFile(sampleFilePath2, incorrectFileName)
			Expect(err).To(BeNil())
			destDir = "/tmp/bar"
			err = unzip(sampleFilePath2, destDir)
			Expect(err).To(Not(BeNil()))
			Expect(err.Error()).To(ContainSubstring("filepath contains directory(parent)"))
		})
	})
})

func setupSampleZipFile(zipFileName, contentFileName string) error {
	nestedDir, err := os.Create(baseDir + "/" + contentFileName)
	if err != nil {
		return err
	}

	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	fileToZip, err := os.Open(nestedDir.Name())
	if err != nil {
		return err
	}

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = fileToZip.Name()

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

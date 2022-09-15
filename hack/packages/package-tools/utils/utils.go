// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GetProjectRootDir return project root directory
func GetProjectRootDir() (string, error) {
	cmdOut, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("couldn't get the project root dir: %w", err)
	}
	return strings.TrimSpace(string(cmdOut)), nil
}

// CreateDir creates a directory
func CreateDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("couldn't create directory %s: %w", path, err)
	}
	return nil
}

// RunMakeTarget runs a make target
func RunMakeTarget(path, target string, envArray ...string) error {
	cmd := exec.Command("make", "-C", path, target)
	var errBytes bytes.Buffer
	cmd.Stderr = &errBytes
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, envArray...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("couldn't run the make target %s: %s", target, errBytes.String())
	}
	return nil
}

// CreateTarball creates a tarball of contents of a path
func CreateTarball(tarballPath, tarballFileName, pathToContents string) error {
	if err := CreateDir(tarballPath); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(tarballPath, tarballFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("couldn't create tarball file %s: %w", tarballFileName, err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(pathToContents, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, pathToContents, "", -1), string(filepath.Separator))

		// write the header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tarWriter, f); err != nil {
			return err
		}

		f.Close()

		return nil
	})
}

// Untar function untars a tarball
func Untar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		// this is an action taken for gosec rule G305: File traversal when extracting zip/tar archive
		if err := sanitizeExtractPath(header.Name, dst); err != nil {
			return err
		}
		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name) //nolint: gosec

		// check the file type
		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			err := os.MkdirAll(filepath.Dir(target), 0755)
			if err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil { //nolint: gosec
				return err
			}

			f.Close()
		}
	}
}

// AfterString is for getting a substring after a string.
func AfterString(str, after string) string {
	pos := strings.LastIndex(str, after)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(after)
	if adjustedPos >= len(str) {
		return ""
	}
	return str[adjustedPos:]
}

// GetFormattedCurrentTime return time in RFC3339 format
func GetFormattedCurrentTime() string {
	t := time.Now()
	return t.Format(time.RFC3339)
}

// sanitizeExtractPath is validation for zip slip vulnerability https://snyk.io/research/zip-slip-vulnerability
func sanitizeExtractPath(destination, filePath string) error {
	destinationPath := filepath.Join(destination, filePath)
	if !strings.HasPrefix(destinationPath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return fmt.Errorf("illegal file path: %s", filePath)
	}
	return nil
}

// IsStringEmpty checks if a string is empty
func IsStringEmpty(str string) bool {
	return str == ""
}

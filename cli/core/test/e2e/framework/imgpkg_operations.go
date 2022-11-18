// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"strings"

	"github.com/pkg/errors"
)

type ImgpkgOps interface {
	PushBinary(filepath, registryBucketURL string) (registryURL string, err error)
	PushBundle(filepath, registryBucketURL string) (registryURL string, err error)
	PullBinary(url string, outputPath string) (stdOut string, err error)
	PullBundle(url string, outputPath string) (stdOut string, err error)
}

func NewImgpkgOps() ImgpkgOps {
	return &imgpkgOps{
		cmdExe: NewCmdOps(),
	}
}

type imgpkgOps struct {
	cmdExe CmdOps
}

func (ip *imgpkgOps) PushBinary(artifactImage, registryBucketURL string) (url string, err error) {
	cmd := "imgpkg " + "push " + "-i " + registryBucketURL + " -f " + artifactImage + " --json"
	return ip.pushAndProcessOutput(cmd)
}

func (ip *imgpkgOps) PushBundle(artifactFile, registryBucketURL string) (url string, err error) {
	cmd := "imgpkg " + "push " + "-b " + registryBucketURL + " -f " + artifactFile + " --json"
	return ip.pushAndProcessOutput(cmd)
}

func (ip *imgpkgOps) pushAndProcessOutput(cmd string) (url string, err error) {
	out, stdErr, err := ip.cmdExe.Exec(cmd)
	stdErrStr := stdErr.String()
	outStr := out.String()
	if err != nil {
		return "", errors.Wrap(err, stdErrStr)
	}
	pushStr := "\"Pushed '"
	i := strings.Index(outStr, pushStr)
	if i < 0 {
		return "", errors.New("invalid out from imgpkg push command")
	}
	remStr := outStr[(i + len(pushStr)):]
	j := strings.Index(remStr, "'")
	if j < 0 {
		return "", errors.New("invalid out from imgpkg push command")
	}
	registryBinaryUrl := remStr[:j]
	return registryBinaryUrl, nil
}

func (ip *imgpkgOps) PullBinary(registryURL, outputPath string) (stdOut string, err error) {
	cmd := "imgpkg " + " pull " + " -i " + registryURL + " -o " + outputPath + " --json "
	return ip.pullAndProcessOutput(cmd)
}
func (ip *imgpkgOps) PullBundle(registryURL, outputPath string) (stdOut string, err error) {
	cmd := "imgpkg " + " pull " + " -b " + registryURL + " -o " + outputPath + " --json "
	return ip.pullAndProcessOutput(cmd)
}

func (ip *imgpkgOps) pullAndProcessOutput(cmd string) (stdOut string, err error) {
	stdOutBuffer, _, err := ip.cmdExe.Exec(cmd)
	return stdOutBuffer.String(), err
}

// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package utils adding package comment to satisfy linters
package utils

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

// nolint:gosec
var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

var (
	numSplits          = 2
	privatePort uint16 = 6443
)

// ContainsString checks the string contains in string array
func ContainsString(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// ReplaceSpecialChars replaces special character in string
func ReplaceSpecialChars(str string) string {
	str = strings.ReplaceAll(str, ".", "-")
	return strings.ReplaceAll(str, "+", "-")
}

// ReplaceVersionInDockerImage replaces version in existing docker image with a new version.
// This simply replaces the tag in REGISTRY[:PORT]/REPO/MORE/PATH/ELEMENTS/IMAGE[:TAG]
func ReplaceVersionInDockerImage(image, newVersion string) (string, error) {
	if image == "" {
		return "", errors.New("cannot replace version on empty image string")
	}

	splitVersion := strings.Split(image, "/")
	// must have two slashes and three parts after split
	if len(splitVersion) <= numSplits {
		return "", errors.New("not enough slashes in docker image when replacing version")
	}

	// tag is always at the end
	tagSplit := strings.Split(splitVersion[len(splitVersion)-1], ":")
	if len(tagSplit) > numSplits {
		return "", errors.New("extra colons in docker image tag when replacing version, only two allowed")
	}

	// check that, if we have a tag, it doesn't have slashes
	splitTagSlashes := strings.Split(tagSplit[0], "/")
	if len(splitTagSlashes) > 1 {
		return "", errors.New("there cannot be slashes in the docker tag")
	}

	// first join image[:tag]
	splitVersion[len(splitVersion)-1] = strings.Join([]string{tagSplit[0], newVersion}, ":")

	// then rejoin everything else
	finalImage := strings.Join(splitVersion, "/")

	return finalImage, nil
}

// CompareVMwareVersionStrings compares vmware aware versions
func CompareVMwareVersionStrings(v1, v2 string) (int, error) {
	v1s := strings.Split(v1, "+vmware.")
	v2s := strings.Split(v2, "+vmware.")

	if len(v1s) <= 1 || len(v2s) <= 1 {
		return 0, errors.New("invalid version string")
	}

	v1Semver, err := semver.NewVersion(v1s[0])
	if err != nil {
		return 0, errors.Wrapf(err, "unable to parse %v", v1s[0])
	}
	v2Semver, err := semver.NewVersion(v2s[0])
	if err != nil {
		return 0, errors.Wrapf(err, "unable to parse %v", v2s[0])
	}

	compareResult := v1Semver.Compare(v2Semver)
	if compareResult != 0 {
		return compareResult, nil
	}

	v1VmwareBuildVersion, err1 := strconv.Atoi(v1s[1])
	v2VmwareBuildVersion, err2 := strconv.Atoi(v2s[1])
	if err1 != nil || err2 != nil {
		return 0, errors.New("invalid version string")
	}

	return v1VmwareBuildVersion - v2VmwareBuildVersion, nil
}

// CheckKubernetesUpgradeCompatibility checks if a tkg cluster with a k8s version can be upgraded to another version.
// Updrading operations is not supported if the gap between the minor versions is larger than 1.
// For example upgrading from v1.17.9 to v1.19.1 is not supported.
// The behavior of this function may be changed in the future, if the upstream supports upgrading between versions with larger gaps.
func CheckKubernetesUpgradeCompatibility(fromVersion, toVersion string) bool {
	v1Versions, err := utilversion.ParseSemantic(fromVersion)
	if err != nil {
		return false
	}

	v2Versions, err := utilversion.ParseSemantic(toVersion)
	if err != nil {
		return false
	}

	if v1Versions.Major() != v2Versions.Major() {
		return false
	}

	if minorGap := int(v2Versions.Minor()) - int(v1Versions.Minor()); minorGap > 1 || minorGap < 0 {
		return false
	} else if minorGap == 0 && v2Versions.Patch() < v1Versions.Patch() {
		return false
	}

	return true
}

// GenerateRandomID generates random string
func GenerateRandomID(length int, excludeCapitalLetters bool) string {
	b := make([]rune, length)
	runes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	if excludeCapitalLetters {
		runes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	}
	for i := range b {
		b[i] = runes[seededRand.Intn(len(runes))]
	}
	return string(b)
}

// ToSnakeCase converts string to SnakeCase with all upper case letters
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToUpper(snake)
}

// IsValidURL validates urls
func IsValidURL(s string) bool {
	if _, err := url.ParseRequestURI(s); err != nil {
		return false
	}

	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

// FixKubeConfigForMacEnvironment fix api server endpoint
func FixKubeConfigForMacEnvironment(dockerContext context.Context, cli *client.Client, kubeconfigBytes []byte) ([]byte, error) {
	log.V(6).Info("Mac and CAPD environment detected, fixing Kubeconfig")
	config, err := clientcmd.Load(kubeconfigBytes)
	if err != nil {
		return nil, errors.Wrap(err, "unable load kubeconfig")
	}

	containers, err := cli.ContainerList(dockerContext, types.ContainerListOptions{})
	if err != nil {
		log.V(3).Error(err, "unable get list of docker containers")
		return kubeconfigBytes, nil
	}

	clusterPortMap := map[string]uint16{}

	for i := 0; i < len(containers); i++ {
		var apiServerPublicPort uint16
		for _, portInfo := range containers[i].Ports {
			if portInfo.PrivatePort == privatePort {
				apiServerPublicPort = portInfo.PublicPort
			}
		}
		for _, name := range containers[i].Names {
			clusterPortMap[name] = apiServerPublicPort
		}
	}

	for clusterName, clusterConfig := range config.Clusters {
		lbContainerName := "/" + clusterName + "-lb"
		publicPort, exists := clusterPortMap[lbContainerName]
		if !exists {
			continue
		}
		clusterConfig.Server = fmt.Sprintf("https://127.0.0.1:%v", publicPort)
	}

	updatedKubeconfigBytes, err := clientcmd.Write(*config)
	if err != nil {
		log.V(3).Error(err, "unable write kubeconfig")
		return kubeconfigBytes, nil
	}

	return updatedKubeconfigBytes, nil
}

// DivideVPCCidr divide VPC cidr as per extendedBits and number of subnets needed
func DivideVPCCidr(cidrStr string, extendedBits, numSubnets int) ([]string, error) {
	results := make([]string, 0)
	_, baseSubnet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return nil, err
	}
	for i := 0; i < numSubnets; i++ {
		subNet, err := cidr.Subnet(baseSubnet, extendedBits, i)
		if err != nil {
			return nil, err
		}

		results = append(results, subNet.String())
	}

	return results, nil
}

// IsOnWindows returns true if running on a Windows machine.
func IsOnWindows() bool {
	return runtime.GOOS == "windows"
}

// GetTkrNameFromTkrVersion gets TKr name from TKr version
func GetTkrNameFromTkrVersion(tkrVersion string) string {
	strs := strings.Split(tkrVersion, "+")
	if len(strs) != numSplits {
		return tkrVersion
	}
	return strs[0] + "---" + strs[1]
}

// GetTKRVersionFromTKRName gets TKr version from TKr name
func GetTKRVersionFromTKRName(tkrName string) string {
	strs := strings.Split(tkrName, "---")
	if len(strs) != numSplits {
		return tkrName
	}
	return strs[0] + "+" + strs[1]
}

// GetTKGBoMTagFromFileName gets BOM Tag from filename
func GetTKGBoMTagFromFileName(fileName string) string {
	return strings.TrimSuffix(strings.TrimPrefix(fileName, "tkg-bom-"), ".yaml")
}

// CompareMajorMinorPatchVersion returns true if major/minor/patch parts of the versions match, else false
func CompareMajorMinorPatchVersion(version1, version2 string) bool {
	semVersion1, err := utilversion.ParseSemantic(version1)
	if err != nil {
		return false
	}
	semVersion2, err := utilversion.ParseSemantic(version2)
	if err != nil {
		return false
	}
	if semVersion1.Major() == semVersion2.Major() &&
		semVersion1.Minor() == semVersion2.Minor() &&
		semVersion1.Patch() == semVersion2.Patch() {
		return true
	}
	return false
}

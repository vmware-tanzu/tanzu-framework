// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Utils Suite")
}

var testingDir string

var _ = Describe("Common Utils", func() {
	var (
		err          error
		clientconfig *clientcmdapi.Config
		dockerClient *dockerclient.Client
		ctx          context.Context
	)

	BeforeSuite(createTempDirectory)
	AfterSuite(deleteTempDirectory)

	reInitialize := func() {
		clientconfig = &clientcmdapi.Config{}
		dockerClient = &dockerclient.Client{}
		ctx = context.Background()
	}

	Describe("Fix KubeConfig for Mac Environment", func() {
		BeforeEach(func() {
			reInitialize()
		})
		Context("When there is a matching cluster in the kubeconfig", func() {
			BeforeEach(func() {
				kubeConfigPath := getConfigFilePath("config6.yaml")
				kubeconfigBytes, _ := os.ReadFile(kubeConfigPath)
				lbContainer := types.Container{
					Names: []string{"/docker-mgmt-1-lb"},
					Ports: []types.Port{
						{
							IP:          "127.0.0.2",
							PrivatePort: 6443,
							PublicPort:  12345,
						},
					},
				}
				containerList := []types.Container{lbContainer}
				dockerClient, err = newListContainerMockDockerClient(containerList)
				var output []byte
				output, err = FixKubeConfigForMacEnvironment(ctx, dockerClient, kubeconfigBytes)
				Expect(err).NotTo(HaveOccurred())
				clientconfig, err = clientcmd.Load(output)
			})
			It("Should change the load balancer server socket", func() {
				Expect(err).NotTo(HaveOccurred())
				actualServer := clientconfig.Clusters["docker-mgmt-1"].Server
				Expect(actualServer).To(Equal("https://127.0.0.1:12345"))
			})
		})
	})
})

func getConfigFilePath(filename string) string {
	filePath := "../fakes/config/kubeconfig/" + filename
	f, _ := os.CreateTemp(testingDir, "kube")
	copyFile(filePath, f.Name())
	return f.Name()
}

func createTempDirectory() {
	testingDir, _ = os.MkdirTemp("", "cluster_client_test")
}

func deleteTempDirectory() {
	os.Remove(testingDir)
}

func copyFile(sourceFile, destFile string) {
	input, _ := os.ReadFile(sourceFile)
	_ = os.WriteFile(destFile, input, constants.ConfigFilePermissions)
}

// the following was brought from docker, this is how ListContainer is mocked
func newListContainerDoer(containers []types.Container) func(*http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		b, err := json.Marshal(containers)
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(b)),
		}, nil
	}
}

type transportFunc func(*http.Request) (*http.Response, error)

func (tf transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return tf(req)
}

func newMockHTTPClient(doer func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: transportFunc(doer),
	}
}

func newListContainerMockDockerClient(containers []types.Container) (*dockerclient.Client, error) {
	// we don't lint here because we never explicitly read the response body, we
	// create it and docker reads it.
	doer := newListContainerDoer(containers) //nolint
	mockHTTPClient := newMockHTTPClient(doer)
	dockerOpts := dockerclient.WithHTTPClient(mockHTTPClient)
	return dockerclient.NewClientWithOpts(dockerOpts)
}

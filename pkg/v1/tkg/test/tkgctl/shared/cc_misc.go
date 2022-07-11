// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/cluster-api/util"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

func E2ECCMiscSpec(context context.Context, inputGetter func() E2ECommonSpecInput) {
	var (
		err          error
		input        E2ECommonSpecInput
		tkgCtlClient tkgctl.TKGClient
		logsDir      string
		clusterName  string
		namespace    string
	)

	BeforeEach(func() {
		input = inputGetter()
		namespace = input.Namespace
		logsDir = filepath.Join(input.ArtifactsFolder, "logs")

		rand.Seed(time.Now().UnixNano())
		clusterName = input.E2EConfig.ClusterPrefix + "wc-" + util.RandomString(4) // nolint:gomnd

		tkgCtlClient, err = tkgctl.New(tkgctl.Options{
			ConfigDir: input.E2EConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, clusterName+".log"),
				Verbosity: input.E2EConfig.TkgCliLogLevel,
			},
		})

		Expect(err).To(BeNil())
	})

	It("should test misc clusterclass operations", func() {
		By("Running cluster create dry-run ")
		options := framework.CreateClusterOptions{
			ClusterName:  clusterName,
			Namespace:    namespace,
			Plan:         "devcc",
			CniType:      input.Cni,
			GenerateOnly: true,
		}

		if input.E2EConfig.InfrastructureName == "vsphere" {
			if endpointIP, ok := os.LookupEnv("CLUSTER_ENDPOINT_2"); ok {
				options.VsphereControlPlaneEndpoint = endpointIP
			}
		}

		clusterConfigFile, err := framework.GetTempClusterConfigFile(input.E2EConfig.TkgClusterConfigPath, &options)
		Expect(err).To(BeNil())
		defer os.Remove(clusterConfigFile)

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: clusterConfigFile,
			Edition:           "tkg",
			GenerateOnly:      true,
		})
		Expect(err).To(BeNil())

		outChan := make(chan string)
		go func() {
			var buf bytes.Buffer
			io.Copy(&buf, r)
			outChan <- buf.String()
		}()

		w.Close()
		os.Stdout = old

		out := <-outChan

		var clusterclass map[string]interface{}
		yaml.NewDecoder(bytes.NewBufferString(out)).Decode(&clusterclass)

		Expect(clusterclass["kind"]).To(Equal("ClusterClass"))

		if metadata, ok := clusterclass["metadata"].(map[string]interface{}); ok {
			Expect(metadata["name"]).To(Equal(fmt.Sprintf("tkg-%s-default", input.E2EConfig.InfrastructureName)))
		}

		By("running cluster create dry-run with a cc config file")
		ccConfig, err := os.CreateTemp("", "cc_config")
		Expect(err).ToNot(HaveOccurred())
		defer ccConfig.Close()

		io.Copy(ccConfig, bytes.NewBufferString(out))

		old = os.Stdout
		r, w, _ = os.Pipe()
		os.Stdout = w

		err = tkgCtlClient.CreateCluster(tkgctl.CreateClusterOptions{
			ClusterConfigFile: ccConfig.Name(),
			Edition:           "tkg",
			GenerateOnly:      true,
		})
		Expect(err).To(BeNil())

		outChan = make(chan string)
		go func() {
			var buf bytes.Buffer
			io.Copy(&buf, r)
			outChan <- buf.String()
		}()

		w.Close()
		os.Stdout = old

		outAgain := <-outChan

		Expect(out).To(Equal(outAgain))
	})
}

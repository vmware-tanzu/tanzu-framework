// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package shared

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type E2ECEIPSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
}

func E2ECEIPSpec(context context.Context, inputGetter func() E2ECEIPSpecInput) {
	var (
		err           error
		input         E2ECEIPSpecInput
		tkgCtlClient  tkgctl.TKGClient
		mcProxy       *framework.ClusterProxy
		logsDir       string
		mcContextName string
	)

	BeforeEach(func() {
		input = inputGetter()
		logsDir = filepath.Join(input.ArtifactsFolder, "logs")

		rand.Seed(time.Now().UnixNano())
		mcClusterName := input.E2EConfig.ManagementClusterName
		mcContextName = mcClusterName + "-admin@" + mcClusterName
		mcProxy = framework.NewClusterProxy(mcClusterName, "", mcContextName)

		rand.Seed(time.Now().UnixNano())

		tkgCtlClient, err = tkgctl.New(tkgctl.Options{
			ConfigDir: input.E2EConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, mcClusterName+".log"),
				Verbosity: input.E2EConfig.TkgCliLogLevel,
			},
		})

		Expect(err).To(BeNil())
	})

	Describe("should verify telemetry job urls for stage and prod", func() {
		It("should verify ceip opted out", func() {
			_, _ = GinkgoWriter.Write([]byte("Setting opt out status"))
			err := tkgCtlClient.SetCeip("false", "", "")
			Expect(err).ToNot(HaveOccurred())

			duration := 10 * time.Second
			time.Sleep(duration)

			optOutStatus, err := tkgCtlClient.GetCEIP()
			Expect(err).ToNot(HaveOccurred())
			Expect(optOutStatus.CeipStatus).To(Equal("Opt-out"))
		})

		It("should verify prod telemetry url added", func() {
			err = tkgCtlClient.SetCeip("true", "true", "")
			Expect(err).ToNot(HaveOccurred())

			cStatus, err := tkgCtlClient.GetCEIP()
			Expect(err).ToNot(HaveOccurred())
			Expect(cStatus.CeipStatus).To(Equal("Opt-in"))

			err = verifyTelemetryJobURL(context, "https://scapi.vmware.com", mcProxy)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should verify staging telemetry url added", func() {
			err = tkgCtlClient.SetCeip("true", "false", "")
			Expect(err).ToNot(HaveOccurred())

			cStatus, err := tkgCtlClient.GetCEIP()
			Expect(err).ToNot(HaveOccurred())
			Expect(cStatus.CeipStatus).To(Equal("Opt-in"))

			err = verifyTelemetryJobURL(context, "https://scapi-stg.vmware.com", mcProxy)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	AfterEach(func() {
	})
}

func verifyTelemetryJobURL(context context.Context, url string, mcProxy *framework.ClusterProxy) error {
	client := mcProxy.GetClient()
	cronJob := &v1beta1.CronJob{}

	_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("Context : %s \n", context)))
	err := client.Get(context, types.NamespacedName{Name: "tkg-telemetry", Namespace: "tkg-system-telemetry"}, cronJob)
	if err != nil {
		return err
	}

	container := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
	commands := container.Command

	for _, command := range commands {
		if strings.Contains(command, url) {
			return nil
		}
	}

	return errors.New("URL not found in the telemetry cron job")
}

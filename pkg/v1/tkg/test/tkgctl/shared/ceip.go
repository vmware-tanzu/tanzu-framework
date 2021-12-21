// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,stylecheck,nolintlint
package shared

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

const (
	telemetryNamespace = "tkg-system-telemetry"
	telemetryName      = "tkg-telemetry"
	phaseSucceeded     = "Succeeded"
	phaseRunning       = "Running"
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

	It("should verify CEIP opt-in and opt-out and verify telemetry job and pod are running", func() {
		By("should verify ceip opted out")
		err := tkgCtlClient.SetCeip("false", "", "")
		Expect(err).ToNot(HaveOccurred())

		duration := 5 * time.Second
		time.Sleep(duration)

		optOutStatus, err := tkgCtlClient.GetCEIP()
		Expect(err).ToNot(HaveOccurred())
		Expect(optOutStatus.CeipStatus).To(Equal("Opt-out"))

		By("should verify ceip opted in and prod telemetry url")
		err = tkgCtlClient.SetCeip("true", "true", "")
		Expect(err).ToNot(HaveOccurred())

		cStatus, err := tkgCtlClient.GetCEIP()
		Expect(err).ToNot(HaveOccurred())
		Expect(cStatus.CeipStatus).To(Equal("Opt-in"))

		err = verifyTelemetryJobURL(context, "https://scapi.vmware.com", mcProxy)
		Expect(err).ToNot(HaveOccurred())

		By("should verify ceip opted in and stage telemetry url")
		err = tkgCtlClient.SetCeip("true", "false", "")
		Expect(err).ToNot(HaveOccurred())

		cStatus, err = tkgCtlClient.GetCEIP()
		Expect(err).ToNot(HaveOccurred())
		Expect(cStatus.CeipStatus).To(Equal("Opt-in"))

		err = verifyTelemetryJobURL(context, "https://scapi-stg.vmware.com", mcProxy)
		Expect(err).ToNot(HaveOccurred())

		By("should verify telemetry job and pod are running")
		err = verifyTelemetryJobRunning(context, mcProxy)
		Expect(err).ToNot(HaveOccurred())
	})
}

func verifyTelemetryJobURL(context context.Context, url string, mcProxy *framework.ClusterProxy) error {
	client := mcProxy.GetClient()
	cronJob := &v1beta1.CronJob{}

	_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("Context : %s \n", context)))
	err := client.Get(context, types.NamespacedName{Name: telemetryName, Namespace: telemetryNamespace}, cronJob)
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

func verifyTelemetryJobRunning(context context.Context, mcProxy *framework.ClusterProxy) error {
	var (
		err          error
		selectors    = []crtclient.ListOption{crtclient.InNamespace(telemetryNamespace)}
		client       = mcProxy.GetClient()
		pollInterval = 30 * time.Second
		pollTimeout  = 90 * time.Second
	)

	scheme := mcProxy.GetScheme()
	batchv1.AddToScheme(scheme)

	cronJob := &v1beta1.CronJob{}
	if err = client.Get(context, types.NamespacedName{Name: telemetryName, Namespace: telemetryNamespace}, cronJob); err != nil {
		return err
	}

	// updating the telemetry cron job schedule to "* * * * *" so that the cronjob can be scheduled to run within the next 59 seconds
	cronJobToUpdate := cronJob.DeepCopy()
	cronJobToUpdate.Spec.Schedule = "* * * * *"
	if err = client.Update(context, cronJobToUpdate); err != nil {
		return err
	}

	// check to see if any telemetry job gets created within pollTimeout time interval
	jobs := &batchv1.JobList{}
	if err = wait.Poll(pollInterval, pollTimeout, func() (done bool, err error) {
		if err = client.List(context, jobs, selectors...); err != nil {
			if k8serr.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		if len(jobs.Items) == 0 {
			return false, nil
		}
		return true, nil
	}); err != nil {
		return err
	}

	// check to see if any telemetry pod gets created within pollTimeout time interval
	pods := &v1.PodList{}
	if err = wait.Poll(pollInterval, pollTimeout, func() (done bool, err error) {
		pods = &v1.PodList{}
		if err = client.List(context, pods, selectors...); err != nil {
			if k8serr.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		if len(pods.Items) == 0 {
			return false, nil
		}
		return true, nil
	}); err != nil {
		return err
	}

	// check to make sure that the telemetry pods are in either of "Running" or "Succeeded" phases
	if err = wait.Poll(pollInterval, pollTimeout, func() (done bool, err error) {
		for i := range pods.Items {
			po := &v1.Pod{}
			if err = client.Get(context, types.NamespacedName{Name: pods.Items[i].Name, Namespace: telemetryNamespace}, po); err != nil {
				if k8serr.IsNotFound(err) { // corner case: A pod in running state which has been kicked off by an older job, might have been terminated and not exist anymore
					continue
				}
				return false, err
			}
			if po.Status.Phase != phaseSucceeded && po.Status.Phase != phaseRunning {
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return err
	}

	// returning the telemetry cron job schedule back to "0 */6 * * *"
	cronJob = &v1beta1.CronJob{}
	if err = client.Get(context, types.NamespacedName{Name: telemetryName, Namespace: telemetryNamespace}, cronJob); err != nil {
		return err
	}
	cronJobToUpdate = cronJob.DeepCopy()
	cronJobToUpdate.Spec.Schedule = "0 */6 * * *"
	if err = client.Update(context, cronJobToUpdate); err != nil {
		return err
	}

	log.Info("successfully verified that telemetry job and pods are running")
	return nil
}

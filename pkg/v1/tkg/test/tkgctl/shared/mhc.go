// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,goconst,gocritic,golint,stylecheck,nolintlint
package shared

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/framework"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type E2EMhcSpecInput struct {
	E2EConfig       *framework.E2EConfig
	ArtifactsFolder string
	Cni             string
}

func E2EMhcSpec(context context.Context, inputGetter func() E2EMhcSpecInput) { //nolint:funlen
	var (
		input        E2EMhcSpecInput
		tkgCtlClient tkgctl.TKGClient
		logsDir      string
		clusterName  string
		namespace    string

		mcProxy *framework.ClusterProxy
		wcProxy *framework.ClusterProxy
	)

	BeforeEach(func() {
		var err error
		namespace = constants.DefaultNamespace
		input = inputGetter()
		logsDir = filepath.Join(input.ArtifactsFolder, "logs")

		mcClusterName := input.E2EConfig.ManagementClusterName
		mcContextName := mcClusterName + "-admin@" + mcClusterName
		mcProxy = framework.NewClusterProxy(mcClusterName, "", mcContextName)

		rand.Seed(time.Now().UnixNano())
		clusterName = input.E2EConfig.ClusterPrefix + "wc"

		tkgCtlClient, err = tkgctl.New(tkgctl.Options{
			ConfigDir: input.E2EConfig.TkgConfigDir,
			LogOptions: tkgctl.LoggingOptions{
				File:      filepath.Join(logsDir, clusterName+".log"),
				Verbosity: input.E2EConfig.TkgCliLogLevel,
			},
		})

		Expect(err).To(BeNil())

		By(fmt.Sprintf("Generating credentials for workload cluster %q", clusterName))
		err = tkgCtlClient.GetCredentials(tkgctl.GetWorkloadClusterCredentialsOptions{
			ClusterName: clusterName,
			Namespace:   namespace,
		})
		Expect(err).To(BeNil())

		wcContextName := clusterName + "-admin@" + clusterName
		wcProxy = framework.NewClusterProxy(clusterName, "", wcContextName)
	})

	It("mhc should remediate unhealthy machine", func() {
		// Validate MHC
		By(fmt.Sprintf("Getting MHC for cluster %q", clusterName))
		mhcList, err := tkgCtlClient.GetMachineHealthCheck(tkgctl.GetMachineHealthCheckOptions{
			ClusterName:            clusterName,
			MachineHealthCheckName: clusterName,
			Namespace:              namespace,
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(len(mhcList)).To(Equal(1))
		mhc := mhcList[0]
		Expect(mhc.Spec.ClusterName).To(Equal(clusterName))
		Expect(mhc.Name).To(Equal(clusterName))
		Expect(len(mhc.Spec.UnhealthyConditions)).To(Equal(2)) // nolint:gomnd

		// Delete MHC and verify if MHC is deleted
		By(fmt.Sprintf("Deleting MHC for cluster %q", clusterName))
		if tkgCtlClient == nil {
			_, _ = GinkgoWriter.Write([]byte("tkgCtlClient is nil"))
		}
		err = tkgCtlClient.DeleteMachineHealthCheck(tkgctl.DeleteMachineHealthCheckOptions{
			ClusterName:            clusterName,
			MachinehealthCheckName: clusterName,
			Namespace:              namespace,
			SkipPrompt:             true,
		})
		Expect(err).ToNot(HaveOccurred())

		mhcList, err = tkgCtlClient.GetMachineHealthCheck(tkgctl.GetMachineHealthCheckOptions{
			ClusterName:            clusterName,
			Namespace:              namespace,
			MachineHealthCheckName: clusterName,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(len(mhcList)).To(Equal(0))

		// Set MHC and verify if it is set
		By(fmt.Sprintf("Updating MHC for cluster %q", clusterName))
		err = tkgCtlClient.SetMachineHealthCheck(tkgctl.SetMachineHealthCheckOptions{
			ClusterName:            clusterName,
			Namespace:              namespace,
			MachineHealthCheckName: clusterName,
			UnhealthyConditions:    fmt.Sprintf("%s:%s:%s", string(corev1.NodeReady), string(corev1.ConditionFalse), "5m"),
		})
		Expect(err).ToNot(HaveOccurred())

		// Wait for Target in MHC status to get the machine name
		waitForMhcTarget(tkgCtlClient, clusterName, namespace)
		mhcList, err = tkgCtlClient.GetMachineHealthCheck(tkgctl.GetMachineHealthCheckOptions{
			ClusterName:            clusterName,
			Namespace:              namespace,
			MachineHealthCheckName: clusterName,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(len(mhcList)).To(Equal(1))

		mhc = mhcList[0]
		Expect(mhc.Spec.ClusterName).To(Equal(clusterName))
		Expect(mhc.Name).To(Equal(clusterName))
		Expect(len(mhc.Spec.UnhealthyConditions)).To(Equal(1))

		// Set machine to unhealthy and see if that machine is remediated
		Expect(len(mhc.Status.Targets)).To(Equal(1))
		machine := mhc.Status.Targets[0]
		By(fmt.Sprintf("Patching Node to make it fail the MHC %q", machine))
		_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("Context : %s \n", context)))
		patchNodeUnhealthy(context, wcProxy, machine, "", mcProxy)

		By("Waiting for the Node to be remediated")
		WaitForNodeRemediation(context, clusterName, "", mcProxy, wcProxy)
	})
}

func getMhcListForCluster(context context.Context, p *framework.ClusterProxy, clusterName string, namespace string) *clusterv1.MachineHealthCheckList {
	mhcList := &clusterv1.MachineHealthCheckList{}
	client := p.GetClient()
	err := client.List(context, mhcList, []crtclient.ListOption{
		crtclient.InNamespace(namespace),
		crtclient.MatchingLabels{
			clusterv1.ClusterLabelName: clusterName,
		},
	}...)
	Expect(err).To(BeNil(), "error fetching mhc for cluster "+clusterName)
	return mhcList
}

func patchNodeUnhealthy(context context.Context, p *framework.ClusterProxy, nodeName string, namespace string, m *framework.ClusterProxy) {
	client := p.GetClient()
	node, err := getNode(context, nodeName, namespace, p, m)
	Expect(err).ToNot(HaveOccurred())

	patchHelper, err := patch.NewHelper(node, client)
	Expect(err).ToNot(HaveOccurred())

	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == corev1.NodeReady {
			node.Status.Conditions[i] = corev1.NodeCondition{
				Type:               corev1.NodeReady,
				Status:             corev1.ConditionFalse,
				LastTransitionTime: metav1.Time{Time: time.Now()},
				LastHeartbeatTime:  metav1.Time{Time: time.Now()},
			}
		}
	}

	Expect(patchHelper.Patch(context, node)).To(Succeed())
}

func getNode(ctx context.Context, machineName string, namespace string, p *framework.ClusterProxy, m *framework.ClusterProxy) (*corev1.Node, error) {
	client := p.GetClient()
	mcClient := m.GetClient()
	node := &corev1.Node{}
	machine := &clusterv1.Machine{}

	_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("Searching for machine with name %s\n", machineName)))
	err := mcClient.Get(ctx, types.NamespacedName{Name: machineName, Namespace: constants.DefaultNamespace}, machine)
	if err != nil {
		return node, err
	}

	_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("Found machine with name %s\n", machineName)))
	_, _ = GinkgoWriter.Write([]byte("Details: \n"))

	nodeName := machine.Status.NodeRef.Name
	_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("Name of node : %s\n", nodeName)))
	if nodeName == "" {
		return node, errors.New("no node name present in machine status")
	}

	if nodeName != "" {
		err = client.Get(ctx, types.NamespacedName{Name: nodeName, Namespace: namespace}, node)
	}

	return node, err
}

func waitForMhcTarget(tkgctlClient tkgctl.TKGClient, clusterName string, namespace string) {
	Eventually(func() bool {
		_, _ = GinkgoWriter.Write([]byte("Waiting for target in MHC status\n"))
		mhcList, err := tkgctlClient.GetMachineHealthCheck(tkgctl.GetMachineHealthCheckOptions{
			ClusterName:            clusterName,
			Namespace:              namespace,
			MachineHealthCheckName: clusterName,
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(len(mhcList)).To(Equal(1))

		mhc := mhcList[0]
		Expect(mhc.Spec.ClusterName).To(Equal(clusterName))
		Expect(mhc.Name).To(Equal(clusterName))
		Expect(len(mhc.Spec.UnhealthyConditions)).To(Equal(1))

		return len(mhc.Status.Targets) == 1
	}, "10m", "30s").Should(BeTrue())
}

func WaitForNodeRemediation(ctx context.Context, clusterName string, namespace string, mcProxy *framework.ClusterProxy, wcProxy *framework.ClusterProxy) {
	Eventually(func() bool {
		_, _ = GinkgoWriter.Write([]byte("Waiting until the unhealthy node is remediated\n"))
		mhcList := getMhcListForCluster(ctx, mcProxy, clusterName, namespace)
		Expect(len(mhcList.Items)).To(Equal(2))
		mhc := mhcList.Items[0]

		for _, nodeName := range mhc.Status.Targets {
			node, err := getNode(ctx, nodeName, namespace, wcProxy, mcProxy)
			if err != nil && apierrors.IsNotFound(err) {
				continue
			}

			if err != nil {
				Fail(err.Error())
			}

			for _, condition := range node.Status.Conditions {
				if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionFalse {
					_, _ = GinkgoWriter.Write([]byte(fmt.Sprintf("Node %s - %s:%s\n", node.Name, string(condition.Type), string(condition.Status))))
					return false
				}
			}
		}

		return true
	}, "10m", "30s").Should(BeTrue())
}

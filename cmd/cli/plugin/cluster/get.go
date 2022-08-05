// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterctltree "sigs.k8s.io/cluster-api/cmd/clusterctl/client/tree"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

const noneTag = "<none>"

type getClustersOptions struct {
	namespace           string
	showOtherConditions string
	// Deprecated: Use showDetails instead.
	disableNoEcho bool
	// Deprecated: Use showGroupMembers instead.
	disableGroupObjects bool
	showDetails         bool
	showGroupMembers    bool
}

var cd = &getClustersOptions{}
var cmdOutput io.Writer

var getClustersCmd = &cobra.Command{
	Use:          "get CLUSTER_NAME",
	Short:        "Get details from a cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         get,
	SilenceUsage: true,
}

func init() {
	cmdOutput = getClustersCmd.OutOrStdout()

	getClustersCmd.Flags().StringVarP(&cd.namespace, "namespace", "n", "", "The namespace from which to get workload clusters. If not provided clusters from all namespaces will be returned")

	getClustersCmd.Flags().StringVar(&cd.showOtherConditions, "show-all-conditions", "", "List of comma separated kind or kind/name for which we should show all the object's conditions (all to show conditions for all the objects)")

	getClustersCmd.Flags().BoolVar(&cd.disableNoEcho, "disable-no-echo", false, "Disable hiding of a MachineInfrastructure and BootstrapConfig when ready condition is true or it has the Status, Severity and Reason of the machine's object")
	cli.DeprecateFlagWithAlternative(getClustersCmd, "disable-no-echo", "1.6.0", "--show-details")
	getClustersCmd.Flags().BoolVar(&cd.showDetails, "show-details", false, "Show details of MachineInfrastructure and BootstrapConfig when ready condition is true or it has the Status, Severity and Reason of the machine's object")

	getClustersCmd.Flags().BoolVar(&cd.disableGroupObjects, "disable-grouping", false, "Disable grouping machines when ready condition has the same Status, Severity and Reason")
	cli.DeprecateFlagWithAlternative(getClustersCmd, "disable-grouping", "1.6.0", "--show-group-members")
	getClustersCmd.Flags().BoolVar(&cd.showGroupMembers, "show-group-members", false, "Expand machine groups whose ready condition has the same Status, Severity and Reason")
}

func get(cmd *cobra.Command, args []string) error {
	f1 := cmd.Flags().Lookup("disable-no-echo")
	f1Changed := false
	if f1 != nil && f1.Changed {
		f1Changed = true
		cd.showDetails = cd.disableNoEcho
	}
	f2 := cmd.Flags().Lookup("show-details")
	f2Changed := false
	if f2 != nil && f2.Changed {
		f2Changed = true
	}
	if f1Changed && f2Changed {
		return fmt.Errorf("only one of --show-details or --disable-no-echo should be set")
	}

	f1 = cmd.Flags().Lookup("disable-grouping")
	f1Changed = false
	if f1 != nil && f1.Changed {
		f1Changed = true
		cd.showGroupMembers = cd.disableGroupObjects
	}
	f2 = cmd.Flags().Lookup("show-group-members")
	f2Changed = false
	if f2 != nil && f2.Changed {
		f2Changed = true
	}
	if f1Changed && f2Changed {
		return fmt.Errorf("only one of --show-group-members or --disable-grouping should be set")
	}

	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting cluster details with a global server is not implemented yet")
	}
	return getCluster(server, args[0])
}

//nolint:whitespace,gocritic
func getCluster(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	// Output the Status
	describeClusterOptions := tkgctl.DescribeTKGClustersOptions{
		ClusterName:         clusterName,
		Namespace:           cd.namespace,
		ShowOtherConditions: cd.showOtherConditions,
		ShowDetails:         cd.showDetails,
		ShowGroupMembers:    cd.showGroupMembers,
	}

	results, err := tkgctlClient.DescribeCluster(describeClusterOptions)
	if err != nil {
		return err
	}

	t := component.NewOutputWriter(cmdOutput, "table", "NAME", "NAMESPACE", "STATUS", "CONTROLPLANE", "WORKERS", "KUBERNETES", "ROLES", "TKR")

	cl := results.ClusterInfo
	clusterRoles := noneTag
	if len(cl.Roles) != 0 {
		clusterRoles = strings.Join(cl.Roles, ",")
	}
	if cl.Name != "" {
		t.AddRow(cl.Name, cl.Namespace, cl.Status, cl.ControlPlaneCount, cl.WorkerCount, cl.K8sVersion, clusterRoles, cl.TKR)
	}

	t.Render()

	if results.Objs != nil && results.Cluster != nil {
		log.Infof("\n\nDetails:\n\n")
		treeView(results.Objs, results.Cluster)
	} else {
		// printing the below at log level 1, so that if users want to know why the tree view is not available(for TKGS) it could provide insights
		log.V(1).Infof("\n\n  Warning! Unable to get cluster ObjectTree/cluster objects, so detailed(tree) view of cluster resources is not available!!\n\n")
	}

	// If it is a Management Cluster, output the providers
	if results.InstalledProviders != nil {
		log.Infof("\n\nProviders:\n\n")
		p := component.NewOutputWriter(cmdOutput, "table", "NAMESPACE", "NAME", "TYPE", "PROVIDERNAME", "VERSION", "WATCHNAMESPACE")
		for _, installedProvider := range results.InstalledProviders.Items {
			p.AddRow(installedProvider.Namespace, installedProvider.Name, installedProvider.Type, installedProvider.ProviderName, installedProvider.Version, installedProvider.WatchedNamespace)
		}
		p.Render()
	}

	return nil

}

const (
	firstElemPrefix = `├─`
	lastElemPrefix  = `└─`
	pipe            = `│ `
	doubleSpace     = "  "
)

var (
	gray   = color.New(color.FgHiBlack)
	red    = color.New(color.FgRed)
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	white  = color.New(color.FgWhite)
	cyan   = color.New(color.FgCyan)
)

// treeView prints object hierarchy to out stream.
func treeView(objs *clusterctltree.ObjectTree, obj client.Object) {
	tbl := uitable.New()
	tbl.Separator = "  "
	tbl.AddRow("NAME", "READY", "SEVERITY", "REASON", "SINCE", "MESSAGE")
	treeViewInner("", tbl, objs, obj)
	fmt.Fprintln(color.Output, tbl)
}

type conditions struct {
	readyColor *color.Color
	age        string
	status     string
	severity   string
	reason     string
	message    string
}

func getCond(c *clusterv1.Condition) conditions {
	maxMessage := 100
	v := conditions{}
	if c == nil {
		return v
	}

	switch c.Status {
	case corev1.ConditionTrue:
		v.readyColor = green
	case corev1.ConditionFalse, corev1.ConditionUnknown:
		switch c.Severity {
		case clusterv1.ConditionSeverityError:
			v.readyColor = red
		case clusterv1.ConditionSeverityWarning:
			v.readyColor = yellow
		default:
			v.readyColor = white
		}
	default:
		v.readyColor = gray
	}

	v.status = string(c.Status)
	v.severity = string(c.Severity)
	v.reason = c.Reason
	v.message = c.Message
	if len(v.message) > maxMessage {
		v.message = fmt.Sprintf("%s ...", v.message[:100])
	}
	v.age = duration.HumanDuration(time.Since(c.LastTransitionTime.Time))

	return v
}

func treeViewInner(prefix string, tbl *uitable.Table, objs *clusterctltree.ObjectTree, obj client.Object) {
	v := conditions{}
	v.readyColor = gray
	minDelim := 2

	ready := clusterctltree.GetReadyCondition(obj)
	name := getName(obj)
	if ready != nil {
		v = getCond(ready)
	}

	if clusterctltree.IsGroupObject(obj) {
		name = white.Add(color.Bold).Sprintf(name)
		items := strings.Split(clusterctltree.GetGroupItems(obj), clusterctltree.GroupItemsSeparator)
		if len(items) <= minDelim {
			v.message = gray.Sprintf("See %s", strings.Join(items, clusterctltree.GroupItemsSeparator))
		} else {
			v.message = gray.Sprintf("See %s, ...", strings.Join(items[:2], clusterctltree.GroupItemsSeparator))
		}
	}
	if !obj.GetDeletionTimestamp().IsZero() {
		name = fmt.Sprintf("%s %s", red.Sprintf("!! DELETED !!"), name)
	}

	tbl.AddRow(
		fmt.Sprintf("%s%s", gray.Sprint(printPrefix(prefix)), name),
		v.readyColor.Sprint(v.status),
		v.readyColor.Sprint(v.severity),
		v.readyColor.Sprint(v.reason),
		v.age,
		v.message)

	chs := objs.GetObjectsByParent(obj.GetUID())

	if clusterctltree.IsShowConditionsObject(obj) {
		otherConditions := clusterctltree.GetOtherConditions(obj)
		for i := range otherConditions {
			cond := otherConditions[i]

			p := ""
			filler := strings.Repeat(" ", 10)
			siblingsPipe := doubleSpace
			if len(chs) > 0 {
				siblingsPipe = pipe
			}
			switch i {
			case len(otherConditions) - 1:
				p = prefix + siblingsPipe + filler + lastElemPrefix
			default:
				p = prefix + siblingsPipe + filler + firstElemPrefix
			}

			v = getCond(cond)
			tbl.AddRow(
				fmt.Sprintf("%s%s", gray.Sprint(printPrefix(p)), cyan.Sprint(cond.Type)),
				v.readyColor.Sprint(v.status),
				v.readyColor.Sprint(v.severity),
				v.readyColor.Sprint(v.reason),
				v.age,
				v.message)
		}
	}

	sort.Slice(chs, func(i, j int) bool {
		return getName(chs[i]) < getName(chs[j])
	})

	for i, child := range chs {
		switch i {
		case len(chs) - 1:
			treeViewInner(prefix+lastElemPrefix, tbl, objs, child)
		default:
			treeViewInner(prefix+firstElemPrefix, tbl, objs, child)
		}
	}
}

func getName(obj client.Object) string {
	if clusterctltree.IsGroupObject(obj) {
		items := strings.Split(clusterctltree.GetGroupItems(obj), clusterctltree.GroupItemsSeparator)
		return fmt.Sprintf("%d Machines...", len(items))
	}

	if clusterctltree.IsVirtualObject(obj) {
		return obj.GetName()
	}

	objName := fmt.Sprintf("%s/%s",
		obj.GetObjectKind().GroupVersionKind().Kind,
		color.New(color.Bold).Sprint(obj.GetName()))

	name := objName
	if objectPrefix := clusterctltree.GetMetaName(obj); objectPrefix != "" {
		name = fmt.Sprintf("%s - %s", objectPrefix, gray.Sprintf(name))
	}
	return name
}

func printPrefix(p string) string {
	if strings.HasSuffix(p, firstElemPrefix) {
		p = strings.Replace(p, firstElemPrefix, pipe, strings.Count(p, firstElemPrefix)-1)
	} else {
		p = strings.ReplaceAll(p, firstElemPrefix, pipe)
	}

	if strings.HasSuffix(p, lastElemPrefix) {
		p = strings.Replace(p, lastElemPrefix, strings.Repeat(" ", len([]rune(lastElemPrefix))), strings.Count(p, lastElemPrefix)-1)
	} else {
		p = strings.ReplaceAll(p, lastElemPrefix, strings.Repeat(" ", len([]rune(lastElemPrefix))))
	}
	return p
}

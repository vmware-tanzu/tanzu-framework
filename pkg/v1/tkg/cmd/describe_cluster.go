// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fabriziopandini/capi-conditions/cmd/kubectl-capi-tree/status"
	"github.com/fatih/color"
	"github.com/gosuri/uitable"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

const (
	// ClusterRoleNone defines cluster role none
	ClusterRoleNone = "<none>"
	// Whilespace defines whitespace
	Whilespace = " "
	// TableSeparator defines table separator
	TableSeparator = "  "
	// MessageLengthLimit defines limit on the length of the message displayed
	MessageLengthLimit = 100
	// GroupItemLimit defines limit on number of groups displayed
	GroupItemLimit = 2
)

type describeClustersOptions struct {
	namespace           string
	showOtherConditions string
	disableNoEcho       bool
	disableGroupObjects bool
}

var cd = &describeClustersOptions{}

var describeClustersCmd = &cobra.Command{
	Use:  "cluster CLUSTER_NAME",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runDescribeCluster(args[0])

		verifyCommandError(err)
	},
}

func init() {
	describeClustersCmd.Flags().StringVarP(&cd.namespace, "namespace", "n", "", "The namespace from which to get workload clusters. If not provided clusters from all namespaces will be returned")

	describeClustersCmd.Flags().StringVar(&cd.showOtherConditions, "show-all-conditions", "", " list of comma separated kind or kind/name for which we should show all the object's conditions (all to show conditions for all the objects)")
	describeClustersCmd.Flags().BoolVar(&cd.disableNoEcho, "disable-no-echo", false, "Disable hiding of a MachineInfrastructure and BootstrapConfig when ready condition is true or it has the Status, Severity and Reason of the machine's object")
	describeClustersCmd.Flags().BoolVar(&cd.disableGroupObjects, "disable-grouping", false, "Disable grouping machines when ready condition has the same Status, Severity and Reason")

	describeCmd.AddCommand(describeClustersCmd)
}

func runDescribeCluster(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	// Output the Status
	describeClusterOptions := tkgctl.DescribeTKGClustersOptions{
		ClusterName:         clusterName,
		Namespace:           cd.namespace,
		ShowOtherConditions: cd.showOtherConditions,
		DisableNoEcho:       cd.disableNoEcho,
		DisableGroupObjects: cd.disableGroupObjects,
	}

	results, err := tkgClient.DescribeCluster(describeClusterOptions)
	if err != nil {
		return err
	}

	t := utils.CreateTableWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"NAME", "NAMESPACE", "STATUS", "CONTROLPLANE", "WORKERS", "KUBERNETES", "ROLES"})

	cl := results.ClusterInfo
	clusterRoles := ClusterRoleNone
	if len(cl.Roles) != 0 {
		clusterRoles = strings.Join(cl.Roles, ",")
	}
	t.AppendRow(table.Row{cl.Name, cl.Namespace, cl.Status, cl.ControlPlaneCount, cl.WorkerCount, cl.K8sVersion, clusterRoles})

	t.Render()
	log.Infof("\n\nDetails:\n\n")
	treeView(results.Objs, results.Cluster)

	// If it is a Management Cluster, output the providers
	if results.InstalledProviders != nil {
		log.Infof("\n\nProviders:\n\n")
		p := utils.CreateTableWriter()
		p.SetOutputMirror(os.Stdout)
		p.AppendHeader(table.Row{"NAMESPACE", "NAME", "TYPE", "PROVIDERNAME", "VERSION", "WATCHNAMESPACE"})
		for i := range results.InstalledProviders.Items {
			p.AppendRow(table.Row{results.InstalledProviders.Items[i].Namespace, results.InstalledProviders.Items[i].Name, results.InstalledProviders.Items[i].Type, results.InstalledProviders.Items[i].ProviderName, results.InstalledProviders.Items[i].Version, results.InstalledProviders.Items[i].WatchedNamespace})
		}
		p.Render()
	}

	return nil
}

const (
	firstElemPrefix = `├─`
	lastElemPrefix  = `└─`
	pipe            = `│ `
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
func treeView(objs *status.ObjectTree, obj controllerutil.Object) {
	tbl := uitable.New()
	tbl.Separator = TableSeparator
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
	if len(v.message) > MessageLengthLimit {
		v.message = fmt.Sprintf("%s ...", v.message[:MessageLengthLimit])
	}
	v.age = duration.HumanDuration(time.Since(c.LastTransitionTime.Time))

	return v
}

func treeViewInner(prefix string, tbl *uitable.Table, objs *status.ObjectTree, obj controllerutil.Object) {
	v := conditions{}
	v.readyColor = gray

	ready := status.GetReadyCondition(obj)
	name := getName(obj)
	if ready != nil {
		v = getCond(ready)
	}

	if status.IsGroupObject(obj) {
		name = white.Add(color.Bold).Sprintf(name)
		items := strings.Split(status.GetGroupItems(obj), status.GroupItemsSeparator)
		if len(items) <= GroupItemLimit {
			v.message = gray.Sprintf("See %s", strings.Join(items, status.GroupItemsSeparator))
		} else {
			v.message = gray.Sprintf("See %s, ...", strings.Join(items[:GroupItemLimit], status.GroupItemsSeparator))
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

	if status.IsShowConditionsObject(obj) {
		otherConditions := status.GetOtherConditions(obj)
		for i := range otherConditions {
			cond := otherConditions[i]

			p := ""
			filler := strings.Repeat(Whilespace, 10)
			siblingsPipe := TableSeparator
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

func getName(obj controllerutil.Object) string {
	if status.IsGroupObject(obj) {
		items := strings.Split(status.GetGroupItems(obj), status.GroupItemsSeparator)
		return fmt.Sprintf("%d Machines...", len(items))
	}

	if status.IsVirtualObject(obj) {
		return obj.GetName()
	}

	objName := fmt.Sprintf("%s/%s",
		obj.GetObjectKind().GroupVersionKind().Kind,
		color.New(color.Bold).Sprint(obj.GetName()))

	name := objName
	if objectPrefix := status.GetMetaName(obj); objectPrefix != "" {
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
		p = strings.Replace(p, lastElemPrefix, strings.Repeat(Whilespace, len([]rune(lastElemPrefix))), strings.Count(p, lastElemPrefix)-1)
	} else {
		p = strings.ReplaceAll(p, lastElemPrefix, strings.Repeat(Whilespace, len([]rune(lastElemPrefix))))
	}
	return p
}

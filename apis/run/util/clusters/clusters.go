// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusters

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/sets"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// AvailableUpgrades returns the available upgrade versions of the cluster
func AvailableUpgrades(cluster *v1beta1.Cluster) sets.StringSet {
	updatesMsg := ""
	if condition := conditions.Get(cluster, runv1.ConditionUpdatesAvailable); condition != nil && condition.Status == v1.ConditionTrue {
		updatesMsg = condition.Message
	}
	return updatesFromConditionMessage(updatesMsg)
}

func updatesFromConditionMessage(updatesMsg string) sets.StringSet {
	if updatesMsg == "" {
		return sets.Strings()
	}
	// Example for message - [<tkr-version-1> <tkr-version-2>]"
	updates := strings.Split(strings.TrimRight(strings.TrimLeft(updatesMsg, "["), "]"), " ")
	return sets.Strings(updates...)
}

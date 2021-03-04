// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"os"
	"testing"

	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/image"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/types"
)

func TestRegistryPullBOM(t *testing.T) {
	if integTest := os.Getenv("TKR_INTEG_TEST"); integTest != "true" {
		return
	}
	bomRegistry := os.Getenv("TEST_BOM_IMAGE_PATH")
	if bomRegistry == "" {
		t.Fatal("TEST_BOM_IMAGE_PATH must be set")
	}

	ro := ctlimg.RegistryOpts{}

	reg := New(&ro)

	tags, err := reg.ListImageTags(bomRegistry)
	if err != nil {
		t.Fatalf("error listing tags should not occurs %s", err.Error())
	}

	t.Log(tags[0])

	content, err := reg.GetFile(bomRegistry, tags[0], "")
	if err != nil {
		t.Fatalf("error getting image content should not occurs %s", err.Error())
	}

	bom, err := types.NewBom(content)
	if err != nil {
		t.Fatalf("error parsing bom content should not occurs %s", err.Error())
	}

	t.Log(bom.GetComponent("kubernetes"))
}

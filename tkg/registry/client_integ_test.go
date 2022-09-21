// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/registry"
)

func TestRegistryPullBOM(t *testing.T) {
	if integTest := os.Getenv("TKR_INTEG_TEST"); integTest != "true" {
		return
	}
	bomRegistry := os.Getenv("TEST_BOM_IMAGE_PATH")
	if bomRegistry == "" {
		t.Fatal("TEST_BOM_IMAGE_PATH must be set")
	}

	numTags, err := strconv.Atoi(os.Getenv("TEST_BOM_IMAGE_NUM"))
	if err != nil {
		t.Fatalf("wrong TEST_BOM_IMAGE_NUM, %s", err.Error())
	}

	ro := ctlimg.Opts{}

	reg, err := New(&ro)
	if err != nil {
		t.Fatalf("error creating registry client")
	}

	tags, err := reg.ListImageTags(bomRegistry)
	if err != nil {
		t.Fatalf("error listing tags should not occurs %s", err.Error())
	}

	if numTags != len(tags) {
		t.Fatal("number of tags does not match")
	}

	t.Log(tags[0])

	content, err := reg.GetFile(fmt.Sprintf("%s:%s", bomRegistry, tags[0]), "")
	if err != nil {
		t.Fatalf("error getting image content should not occurs %s", err.Error())
	}

	bom, err := NewBom(content)
	if err != nil {
		t.Fatalf("error parsing bom content should not occurs %s", err.Error())
	}

	t.Log(bom.GetComponent("kubernetes"))
}

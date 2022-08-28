// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package tkgconfigreaderwriter_test

import (
	"testing"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
)

func Test_New_Success(t *testing.T) {
	client, err := tkgconfigreaderwriter.New("../fakes/config/config.yaml")
	if err != nil {
		t.Errorf("Expected success instead got error:%s", err.Error())
	}

	if client == nil {
		t.Error("Expected new client, got nil")
	}
}

func Test_NewReaderWriterFromConfigFile_Success(t *testing.T) {
	readerWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile("../fakes/config/kubeconfig/config1.yaml", "../fakes/config/config.yaml")
	if err != nil {
		t.Errorf("Expected success, got error: %s", err.Error())
	}

	if readerWriter == nil {
		t.Error("readerWriter was nil, was expecting a valid readerWriter")
	}
}

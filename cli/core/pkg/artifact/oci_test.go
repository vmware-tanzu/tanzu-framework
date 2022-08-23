// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package artifact

import (
	"testing"
)

func TestOCIArtifactWhenMultipleFilesFound(t *testing.T) {
	expectedImageName := "foo"
	artifact := NewOCIArtifact(expectedImageName)
	o, _ := artifact.(*OCIArtifact)
	o.getFilesMapFromImage = func(s string) (map[string][]byte, error) {
		if s != expectedImageName {
			t.Fatalf("Unexpected image in call to get files map. Expected '%s', got '%s'", expectedImageName, s)
		}

		// Return a fake map with 2 files.
		return map[string][]byte{
			"file1": nil,
			"file2": nil,
		}, nil
	}

	data, err := o.Fetch()

	if len(data) != 0 {
		t.Fatalf("Expected to receive a nil map, got '%+v'", data)
	}

	expectedErrorMessage := "oci artifact image for plugin is required to have only 1 file, but found 2"
	if err.Error() != expectedErrorMessage {
		t.Fatalf("Did not receive the expected error message. Expected '%s', got '%s'", expectedErrorMessage, err.Error())
	}
}

// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"fmt"
	"os"
	"testing"
)

func TestIsTTYEnabled(t *testing.T) {
	envTests := []struct {
		key, val string
		want     bool
	}{
		{"TANZU_CLI_NO_COLOR", "1", false},
		{"NO_COLOR", "1", false},
		{"TERM", "dumb", false},
		{"TERM", "duMb", false},
	}

	for _, tt := range envTests {
		testName := fmt.Sprintf("(%v,%v):%v", tt.key, tt.val, tt.want)
		t.Run(testName, func(t *testing.T) {
			err := os.Setenv(tt.key, tt.val)
			if err != nil {
				return
			}
			ans := IsTTYEnabled()
			if ans != tt.want {
				t.Errorf("got %v : want %v", ans, tt.want)
			}
		})
	}
}

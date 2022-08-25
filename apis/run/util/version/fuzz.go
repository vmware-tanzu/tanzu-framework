// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"fmt"
	"strings"

	fuzz "github.com/google/gofuzz"
	"k8s.io/apimachinery/pkg/util/rand"
)

// Fuzz fuzzes the passed version.
func Fuzz(v *Version, _ fuzz.Continue) {
	major := rand.Intn(3)
	minor := rand.Intn(50)
	patch := rand.Intn(30)
	var ss []string
	bmCount := rand.Intn(3)
	for i := 0; i < bmCount; i++ {
		ss = append(ss, fmt.Sprintf("%s.%v", rand.String(rand.IntnRange(3, 10)), rand.IntnRange(1, 5)))
	}
	buildMeta := strings.Join(ss, "-")
	vString := strings.TrimSuffix(fmt.Sprintf("%v.%v.%v+%s", major, minor, patch, buildMeta), "+")
	vNew, _ := ParseSemantic(vString)
	*v = *vNew
}

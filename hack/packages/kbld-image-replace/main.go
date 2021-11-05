// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package main is a utility program to add a newImage path to an existing image in kbld-config.yaml.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	kbld "github.com/k14s/kbld/pkg/kbld/config"
	"sigs.k8s.io/yaml"
)

func main() {
	file := flag.String("kbld-config", "../../../packages/kbld-config.yaml", "Path to kbld-config.yaml file")
	flag.Parse()

	if len(os.Args) != 3 {
		log.Fatal("need image and newImage arguments")
	}

	image := os.Args[1]
	newImage := os.Args[2]

	file = &[]string{filepath.Clean(*file)}[0]

	data, err := os.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}

	config := &kbld.Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		log.Fatal(err)
	}

	found := false
	for i := range config.Overrides {
		if config.Overrides[i].Image == image {
			config.Overrides[i].NewImage = newImage
			found = true
			break
		}
	}

	if !found {
		log.Fatal(fmt.Sprintf("image %q not found in kbld config", image))
	}

	data, err = yaml.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}

	if err = os.WriteFile(*file, data, 0644); err != nil {
		log.Fatal(err)
	}
}

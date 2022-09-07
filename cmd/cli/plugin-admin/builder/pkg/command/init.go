// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/aunum/log"

	"github.com/vmware-tanzu/tanzu-framework/plugin-admin/builder/pkg/template"
)

const (
	github = "github"
)

func Initialize(name, repoType string, dryRun bool) error {
	data := struct {
		RepositoryName string
	}{
		RepositoryName: name,
	}
	targets := template.DefaultInitTargets
	if strings.EqualFold(repoType, github) {
		targets = append(targets, template.GitHubCI)
	} else {
		targets = append(targets, template.GitLabCI)
	}
	for _, target := range targets {
		err := target.Run(name, data, dryRun)
		if err != nil {
			return err
		}
	}

	if dryRun {
		return nil
	}

	c := exec.Command("git", "init", name)
	b, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s -- %s", err, string(b))
	}
	log.Success("successfully created repository")
	return nil
}

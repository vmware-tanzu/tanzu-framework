// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ssh

import (
	"strings"
	"golang.org/x/crypto/ssh/agent"
)

type client struct {
	sshAgent agent.Agent
}


func (c *client) KeysAsString() (string, error) {
	keys, err := c.sshAgent.List()
	if err != nil {
		return "", err
	}
	var keyStrings []string
	for _, key := range keys {
		keyStrings = append(keyStrings, string(key.Marshal()))
	}
	allKeysAsString := strings.Join(keyStrings, "\n")
	return allKeysAsString, nil
}


// New creates an Oracle client
func New() (_ Client, err error) {
	c := &client{}
	conn, err := newConn()
	if err != nil {
		return nil, err
	}
	c.sshAgent = agent.NewClient(conn)

	return c, nil
}

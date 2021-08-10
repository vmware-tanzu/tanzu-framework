// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"strings"
)

// Schema represents any openapi schema that may exist on a cluster
func Schema(name, partialSchema string) *QueryPartialSchema {
	return &QueryPartialSchema{
		schema:   partialSchema,
		presence: true,
		name:     name,
	}
}

// QueryPartialSchema allows for matching a partial schema
// currently this is quite naive string comparison but will be updated fpr better validation
type QueryPartialSchema struct {
	schema   string
	name     string
	presence bool
}

// Name is the name of the query.
func (q *QueryPartialSchema) Name() string {
	return q.name
}

// Run the partial query match
func (q *QueryPartialSchema) Run(config *clusterQueryClientConfig) (bool, error) {
	doc, err := config.discoveryClientset.OpenAPISchema()
	if err != nil {
		return false, err
	}
	// TODO: fix this with better validation means
	if strings.Contains(doc.String(), q.schema) {
		return true, nil
	}
	return false, nil
}

// QueryFailure exposes detail on the query failure for consumers to parse
type QueryFailure struct {
	Target   QueryTarget
	Presence bool
}

// Reason returns  the query failure, of it failed
// todo: this should be a results{} struct
func (q *QueryPartialSchema) Reason() string {
	return fmt.Sprintf("method=partial-schema name=%s status=unmatched presence=%t", q.name, q.presence)
}

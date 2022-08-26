// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ini

// Walk will traverse the AST using the v, the Visitor.
func Walk(tree []AST, v Visitor) error {
	for i := range tree {
		switch tree[i].Kind {
		case ASTKindExpr,
			ASTKindExprStatement:

			if err := v.VisitExpr(&tree[i]); err != nil {
				return err
			}
		case ASTKindStatement,
			ASTKindCompletedSectionStatement,
			ASTKindNestedSectionStatement,
			ASTKindCompletedNestedSectionStatement:

			if err := v.VisitStatement(&tree[i]); err != nil {
				return err
			}
		}
	}

	return nil
}

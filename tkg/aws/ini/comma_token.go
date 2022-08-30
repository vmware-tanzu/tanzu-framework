// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ini

var commaRunes = []rune(",")

func isComma(b rune) bool {
	return b == ','
}

func newCommaToken() Token {
	return newToken(TokenComma, commaRunes, NoneType)
}

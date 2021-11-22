// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

import (
	"strings"
	"testing"
)

func shouldContain(t *testing.T, result, expected string) {
	if !strings.Contains(result, expected) {
		t.Fatalf("Result \"%s\" was expected to contain \"%s\" (but doesn't)", result, expected)
	}
}

func shouldBeBlank(t *testing.T, result, description string) {
	if result != "" {
		t.Fatalf("Test \"%s\" expected a blank result, but got \"%s\"", description, result)
	}
}

func TestMapToConfigString(t *testing.T) {
	target := map[string]string{"key1": "value1", "key2": "value2"}
	result := mapToConfigString(target)
	shouldContain(t, result, "key1:value1,")
	shouldContain(t, result, "key2:value2,")

	result = mapToConfigString(nil)
	shouldBeBlank(t, result, "sending nil")

	result = mapToConfigString(map[string]string{})
	shouldBeBlank(t, result, "sending empty map")
}

func shouldBeEmpty(t *testing.T, description string, result map[string]string) {
	if len(result) > 0 {
		t.Fatalf("Test \"%s\" expected an empty map to result, but got \"%v\"", description, result)
	}
}

func shouldHaveOneEntry(t *testing.T, description string, result map[string]string, key, value string) {
	if len(result) != 1 {
		t.Fatalf("Test \"%s\" expected a single-entry map to result, but got \"%v\"", description, result)
	} else if result[key] != value {
		t.Fatalf("Test \"%s\" expected a result of map[%s:%s], but got \"%v\"", description, key, value, result)
	}
}

func shouldHaveTwoEntries(t *testing.T, description string, result map[string]string, key1 string, value1 string, key2 string, value2 string) {
	if len(result) != 2 {
		t.Fatalf("Test \"%s\" expected a double-entry map to result, but got \"%v\"", description, result)
	} else if result[key1] != value1 || result[key2] != value2 {
		t.Fatalf("Test \"%s\" expected a result of map[%s:%s %s:%s], but got \"%v\"", description, key1, value1, key2, value2, result)
	}
}

func TestConfigStringToMap(t *testing.T) {
	target := ""
	result := configStringToMap(target)
	shouldBeEmpty(t, "empty string", result)

	target = "malformedKeyValue"
	result = configStringToMap(target)
	shouldBeEmpty(t, "string of "+target, result)

	target = "too:many:colons"
	result = configStringToMap(target)
	shouldBeEmpty(t, "string of "+target, result)

	target = "key1:value1"
	result = configStringToMap(target)
	shouldHaveOneEntry(t, "string of "+target, result, "key1", "value1")

	target = "key:value,"
	result = configStringToMap(target)
	shouldHaveOneEntry(t, "string of "+target, result, "key", "value")

	target = "key1:value1,malformedKeyValuePair"
	result = configStringToMap(target)
	shouldHaveOneEntry(t, "string of "+target, result, "key1", "value1")

	target = "malformedKeyValuePair,key1:value1,"
	result = configStringToMap(target)
	shouldHaveOneEntry(t, "string of "+target, result, "key1", "value1")

	target = "key1:value1,,,,"
	result = configStringToMap(target)
	shouldHaveOneEntry(t, "string of "+target, result, "key1", "value1")

	target = "key1:value1,key2:value2,"
	result = configStringToMap(target)
	shouldHaveTwoEntries(t, "string "+target, result, "key1", "value1", "key2", "value2")
}

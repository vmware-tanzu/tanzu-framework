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
	shouldContain(t, result, "key1:value1")
	shouldContain(t, result, "key2:value2")
	if strings.LastIndex(result, ",") == len(result)-1 {
		t.Fatalf("Test with two keys detected ending comma: \"%s\"", result)
	}

	result = mapToConfigString(nil)
	shouldBeBlank(t, result, "sending nil")

	result = mapToConfigString(map[string]string{})
	shouldBeBlank(t, result, "sending empty map")

	result = mapToConfigString(map[string]string{"": "blank", "foo": "bar"})
	if result != "foo:bar" {
		t.Fatalf("Test with blank key expected to be ignored, but got \"%s\"", result)
	}
}

func shouldBeEmpty(t *testing.T, target string, result map[string]string) {
	if len(result) > 0 {
		t.Fatalf("Test sending string \"%s\" expected an empty map to result, but got \"%v\"", target, result)
	}
}

func shouldHaveOneEntry(t *testing.T, target string, result map[string]string, key, value string) {
	if len(result) != 1 {
		t.Fatalf("Test sending string \"%s\" expected a single-entry map to result, but got \"%v\"", target, result)
	} else if result[key] != value {
		t.Fatalf("Test sending string \"%s\" expected a result of map[%s:%s], but got \"%v\"", target, key, value, result)
	}
}

func shouldHaveTwoEntries(t *testing.T, target string, result map[string]string, key1 string, value1 string, key2 string, value2 string) {
	if len(result) != 2 {
		t.Fatalf("Test sending string \"%s\" expected a double-entry map to result, but got \"%v\"", target, result)
	} else if result[key1] != value1 || result[key2] != value2 {
		t.Fatalf("Test sending string \"%s\" expected a result of map[%s:%s %s:%s], but got \"%v\"", target, key1, value1, key2, value2, result)
	}
}

func TestConfigStringToMap(t *testing.T) {
	target := ""
	result := configStringToMap(target)
	shouldBeEmpty(t, "", result)

	target = "malformedKeyValue"
	result = configStringToMap(target)
	shouldBeEmpty(t, target, result)

	target = "too:many:colons"
	result = configStringToMap(target)
	shouldBeEmpty(t, target, result)

	target = "key1:value1"
	result = configStringToMap(target)
	shouldHaveOneEntry(t, target, result, "key1", "value1")

	target = "key:value,"
	result = configStringToMap(target)
	shouldHaveOneEntry(t, target, result, "key", "value")

	target = "key1:value1,malformedKeyValuePair"
	result = configStringToMap(target)
	shouldHaveOneEntry(t, target, result, "key1", "value1")

	target = "malformedKeyValuePair,key1:value1,"
	result = configStringToMap(target)
	shouldHaveOneEntry(t, target, result, "key1", "value1")

	target = "key1:value1,,,,"
	result = configStringToMap(target)
	shouldHaveOneEntry(t, target, result, "key1", "value1")

	target = "key1:value1,key2:value2,"
	result = configStringToMap(target)
	shouldHaveTwoEntries(t, target, result, "key1", "value1", "key2", "value2")

	target = "key1:value1,key2:value2"
	result = configStringToMap(target)
	shouldHaveTwoEntries(t, target, result, "key1", "value1", "key2", "value2")
}

type testStruct struct {
	Foo string
}

func TestGetFieldFromConfig(t *testing.T) {
	fieldValue := "bar"
	target := testStruct{Foo: fieldValue}
	fieldName := "Foo"
	result := getFieldFromConfig(target, fieldName)
	if result != fieldValue {
		t.Fatalf("getFieldFromConfig(%v, '%s') expected return of '%s' but got '%s'", target, fieldName, fieldValue, result)
	}

	unknownField := "foogle"
	result = getFieldFromConfig(target, unknownField)
	if result != "" {
		t.Fatalf("getFieldFromConfig(%v, %s) expected return of '' but got '%s'", target, fieldName, result)
	}
}

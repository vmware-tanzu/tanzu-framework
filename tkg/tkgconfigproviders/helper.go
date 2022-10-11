// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

import "strings"

// mapToConfigString returns a string with key:value pairs separated by ,
// For example: "key1:value1,key2:value2"
// An empty map returns an empty string
// An empty key is ignored
func mapToConfigString(data map[string]string) string {
	result := ""
	for key, value := range data {
		if len(key) > 0 {
			result += key + ":" + value + ","
		}
	}
	result = strings.TrimSuffix(result, ",")
	return result
}

// configStringToMap parses a data string into a map
// The expected format is: "key1:value1,key2:value2"
// Blank values are ignored; any key-value with no colon (or multiple colons) is ignored;
// An empty data string is allowed; an empty map will be returned
// A string ending with a comma is allowed ("key1:value1,key2:value2,")
func configStringToMap(data string) map[string]string {
	result := make(map[string]string)

	entryArray := strings.Split(data, ",")
	for _, entry := range entryArray {
		keyValue := strings.Split(entry, ":")
		if len(keyValue) == 2 {
			result[keyValue[0]] = keyValue[1]
		}
	}
	return result
}

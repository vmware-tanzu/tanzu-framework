// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package openapischema

import (
	"bytes"
	"fmt"
	"os"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"gopkg.in/yaml.v3"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apimachinery/pkg/util/json"
)

var _ = Describe("SchemaDefault", func() {
	Context("When default openapi schema for contour is specified", func() {
		It("SchemaDefault() should return a map object with defaults populated", func() {
			contourSchema, err := os.Open("testdata/contourschema.yaml")
			Expect(err).ToNot(HaveOccurred())
			defer contourSchema.Close()

			contourschema, err := os.ReadFile(contourSchema.Name())
			Expect(err).ToNot(HaveOccurred())

			defaultValues, err := os.Open("testdata/contour-data-values.yaml")
			Expect(err).ToNot(HaveOccurred())
			defer defaultValues.Close()

			expectedDefault, err := os.ReadFile(defaultValues.Name())
			Expect(err).ToNot(HaveOccurred())

			schemadefaulted, err := SchemaDefault(contourschema)
			Expect(err).ToNot(HaveOccurred())

			gotMap := make(map[string]interface{})

			err = yaml.Unmarshal(schemadefaulted, gotMap)
			Expect(err).ToNot(HaveOccurred())

			expectedMap := make(map[string]interface{})
			err = yaml.Unmarshal(expectedDefault, expectedMap)
			Expect(err).ToNot(HaveOccurred())

			if !reflect.DeepEqual(expectedMap, gotMap) {
				Fail(fmt.Sprintf("Expected: \n %s \n\n Got: \n %s", expectedDefault, schemadefaulted))
			}

		})

	})
	Context("When default openapi schema for grafana is specified", func() {
		It("SchemaDefault() should return a map object with defaults populated", func() {
			contourSchema, err := os.Open("testdata/grafanaschema.yaml")
			Expect(err).ToNot(HaveOccurred())
			defer contourSchema.Close()

			contourschema, err := os.ReadFile(contourSchema.Name())
			Expect(err).ToNot(HaveOccurred())

			defaultValues, err := os.Open("testdata/grafana-data-values.yaml")
			Expect(err).ToNot(HaveOccurred())
			defer defaultValues.Close()

			expectedDefault, err := os.ReadFile(defaultValues.Name())
			Expect(err).ToNot(HaveOccurred())

			schemadefaulted, err := SchemaDefault(contourschema)
			Expect(err).ToNot(HaveOccurred())

			gotMap := make(map[string]interface{})

			err = yaml.Unmarshal(schemadefaulted, gotMap)
			Expect(err).ToNot(HaveOccurred())

			expectedMap := make(map[string]interface{})
			err = yaml.Unmarshal(expectedDefault, expectedMap)
			Expect(err).ToNot(HaveOccurred())

			if !reflect.DeepEqual(expectedMap, gotMap) {
				Fail(fmt.Sprintf("Expected: \n %s \n\n Got: \n %s", expectedDefault, schemadefaulted))
			}

		})

	})

})

// schemaDefault tests adopted from k8s.io/apiextensions-apiserver/pkg/apiserver/schema with modifications
var _ = Describe("schemaDefault", func() {

	type TestCase struct {
		JSON     string
		Schema   *structuralschema.Structural
		Expected string
	}

	DescribeTable("with",
		func(t TestCase) {
			var in interface{}
			err := json.Unmarshal([]byte(t.JSON), &in)
			Expect(err).ToNot(HaveOccurred())

			var expected interface{}
			err = json.Unmarshal([]byte(t.Expected), &expected)
			Expect(err).ToNot(HaveOccurred())

			schemaDefault(in, t.Schema)
			if !reflect.DeepEqual(in, expected) {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetIndent("", "  ")
				err := enc.Encode(in)
				Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("unexpected result mashalling error: %v", err))
				Fail(fmt.Sprintf("expected: %s\ngot: %s", t.Expected, buf.String()))
			}
		},

		Entry("empty schema and null object returns null", TestCase{
			JSON:     "null",
			Schema:   nil,
			Expected: "null",
		}),
		Entry("unrelated schema for a scalar returns it back", TestCase{
			JSON:     "4",
			Expected: "4",
			Schema: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					Default: structuralschema.JSON{Object: "foo"},
				},
			},
		}),
		Entry("unrelated schema for a scalar array returns it back", TestCase{
			JSON:     "[1,2]",
			Expected: "[1,2]",
			Schema: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					Default: structuralschema.JSON{Object: "foo"},
				},
			},
		}),
		Entry("schema for object array returns array with missing defaults populated", TestCase{
			JSON:     `[{"a":1},{"b":1},{"c":1}]`,
			Expected: `[{"a":1,"b":"B","c":"C"},{"a":"A","b":1,"c":"C"},{"a":"A","b":"B","c":1}]`,
			Schema: &structuralschema.Structural{
				Items: &structuralschema.Structural{
					Properties: map[string]structuralschema.Structural{
						"a": {
							Generic: structuralschema.Generic{
								Default: structuralschema.JSON{Object: "A"},
							},
						},
						"b": {
							Generic: structuralschema.Generic{
								Default: structuralschema.JSON{Object: "B"},
							},
						},
						"c": {
							Generic: structuralschema.Generic{
								Default: structuralschema.JSON{Object: "C"},
							},
						},
					},
				},
			},
		}),
		Entry("schema for object, array, object and additional properties returns object with missing defaults populated", TestCase{
			JSON:     `{"array":[{"a":1},{"b":2}],"object":{"a":1},"additionalProperties":{"x":{"a":1},"y":{"b":2}}}`,
			Expected: `{"array":[{"a":1,"b":"B"},{"a":"A","b":2}],"object":{"a":1,"b":"O"},"additionalProperties":{"x":{"a":1,"b":"beta"},"y":{"a":"alpha","b":2}},"foo":"bar"}`,
			Schema: &structuralschema.Structural{
				Properties: map[string]structuralschema.Structural{
					"array": {
						Items: &structuralschema.Structural{
							Properties: map[string]structuralschema.Structural{
								"a": {
									Generic: structuralschema.Generic{
										Default: structuralschema.JSON{Object: "A"},
									},
								},
								"b": {
									Generic: structuralschema.Generic{
										Default: structuralschema.JSON{Object: "B"},
									},
								},
							},
						},
					},
					"object": {
						Properties: map[string]structuralschema.Structural{
							"a": {
								Generic: structuralschema.Generic{
									Default: structuralschema.JSON{Object: "N"},
								},
							},
							"b": {
								Generic: structuralschema.Generic{
									Default: structuralschema.JSON{Object: "O"},
								},
							},
						},
					},
					"additionalProperties": {
						Generic: structuralschema.Generic{
							AdditionalProperties: &structuralschema.StructuralOrBool{
								Structural: &structuralschema.Structural{
									Properties: map[string]structuralschema.Structural{
										"a": {
											Generic: structuralschema.Generic{
												Default: structuralschema.JSON{Object: "alpha"},
											},
										},
										"b": {
											Generic: structuralschema.Generic{
												Default: structuralschema.JSON{Object: "beta"},
											},
										},
									},
								},
							},
						},
					},
					"foo": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{Object: "bar"},
						},
					},
				},
			},
		}),
		Entry("schema with empty and null array returns array with missing defaults populated", TestCase{
			JSON:     `[{},{"a":1},{"a":0},{"a":0.0},{"a":""},{"a":null},{"a":[]},{"a":{}}]`,
			Expected: `[{"a":"A"},{"a":1},{"a":0},{"a":0.0},{"a":""},{"a":"A"},{"a":[]},{"a":{}}]`,
			Schema: &structuralschema.Structural{
				Items: &structuralschema.Structural{
					Properties: map[string]structuralschema.Structural{
						"a": {
							Generic: structuralschema.Generic{
								Default: structuralschema.JSON{Object: "A"},
							},
						},
					},
				},
			},
		}),
		Entry("schema with null in a nullable list returns null", TestCase{
			JSON:     `[null]`,
			Expected: `[null]`,
			Schema: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					Nullable: true,
				},
				Items: &structuralschema.Structural{
					Properties: map[string]structuralschema.Structural{
						"a": {
							Generic: structuralschema.Generic{
								Default: structuralschema.JSON{Object: "A"},
							},
						},
					},
				},
			},
		}),
		Entry("schema with null in a non-nullable list returns list with default", TestCase{
			JSON:     `[null]`,
			Expected: `["A"]`,
			Schema: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					Nullable: false,
				},
				Items: &structuralschema.Structural{
					Generic: structuralschema.Generic{
						Default: structuralschema.JSON{Object: "A"},
					},
				},
			},
		}),
		Entry("schema with null in a nullable object returns it as is", TestCase{
			JSON:     `{"a": null}`,
			Expected: `{"a": null}`,
			Schema: &structuralschema.Structural{
				Generic: structuralschema.Generic{},
				Properties: map[string]structuralschema.Structural{
					"a": {
						Generic: structuralschema.Generic{
							Nullable: true,
							Default:  structuralschema.JSON{Object: "A"},
						},
					},
				},
			},
		}),
		Entry("schema with null in a non-nullable object returns it with default populated", TestCase{
			JSON:     `{"a": null}`,
			Expected: `{"a": "A"}`,
			Schema: &structuralschema.Structural{
				Generic: structuralschema.Generic{},
				Properties: map[string]structuralschema.Structural{
					"a": {
						Generic: structuralschema.Generic{
							Nullable: false,
							Default:  structuralschema.JSON{Object: "A"},
						},
					},
				},
			},
		}),
		Entry("schema with nullable object and additionalProperties returns it as is", TestCase{
			JSON:     `{"a": null}`,
			Expected: `{"a": null}`,
			Schema: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					AdditionalProperties: &structuralschema.StructuralOrBool{
						Structural: &structuralschema.Structural{
							Generic: structuralschema.Generic{
								Nullable: true,
								Default:  structuralschema.JSON{Object: "A"},
							},
						},
					},
				},
			},
		}),
		Entry("schema with non-nullable object and additionalProperties returns it with default populated", TestCase{
			JSON:     `{"a": null}`,
			Expected: `{"a": "A"}`,
			Schema: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					AdditionalProperties: &structuralschema.StructuralOrBool{
						Structural: &structuralschema.Structural{
							Generic: structuralschema.Generic{
								Nullable: false,
								Default:  structuralschema.JSON{Object: "A"},
							},
						},
					},
				},
			},
		}),
		Entry("schema with null and additionalProperties with unknown field returns it as is", TestCase{
			JSON:     `{"a": null}`,
			Expected: `{"a": null}`,
			Schema: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					AdditionalProperties: &structuralschema.StructuralOrBool{
						Bool: true,
					},
				},
			},
		}),
	)

})

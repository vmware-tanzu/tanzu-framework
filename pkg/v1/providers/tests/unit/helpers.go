package unit

import (
	"io"
	"strings"

	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

type yttValues map[string]interface{}

func (v yttValues) toReader() io.Reader {
	return strings.NewReader(createDataValues(v))
}

func (v yttValues) Set(key string, value interface{}) {
	v[key] = value
}

func (v yttValues) Delete(key string) {
	delete(v, key)
}

func (v yttValues) DeepCopy() yttValues {
	other := make(yttValues)
	for key, value := range v {
		other[key] = value
	}
	return other
}

func assertNotFound(docs []string, err error) {
	Expect(err).NotTo(HaveOccurred())
	Expect(docs).To(HaveLen(0))
}

func assertFoundOne(docs []string, err error) {
	Expect(err).NotTo(HaveOccurred())
	Expect(docs).To(HaveLen(1))
}

func createDataValues(values map[string]interface{}) string {
	dataValues := "#@data/values\n---\n"
	bytes, err := yaml.Marshal(values)
	if err != nil {
		return ""
	}
	valuesStr := string(bytes)
	valuesStr = strings.ReplaceAll(valuesStr, "\"true\"", "true")
	valuesStr = strings.ReplaceAll(valuesStr, "\"false\"", "false")
	return dataValues + valuesStr
}

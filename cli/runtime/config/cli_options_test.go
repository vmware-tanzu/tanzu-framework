package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestGetEdition(t *testing.T) {

	// setup
	func() {
		LocalDirName = fmt.Sprintf(".tanzu-test")
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tests := []struct {
		name   string
		in     *configapi.ClientConfig
		out    string
		errStr string
	}{
		{
			name: "success k8s",
			in: &configapi.ClientConfig{
				ClientOptions: &configapi.ClientOptions{
					Env: map[string]string{
						"test": "test",
					},
				},
			},
			out: "test",
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := StoreClientConfig(spec.in)
			if err != nil {
				fmt.Printf("StoreClientConfigV2 errors: %v\n", err)
			}

			c, err := GetEnv("test")
			if err != nil {
				fmt.Printf("errors: %v\n", err)
			}

			assert.Equal(t, spec.out, c)
			assert.NoError(t, err)

		})
	}
}

func TestSetEdition(t *testing.T) {

	// setup
	func() {
		LocalDirName = fmt.Sprintf(".tanzu-test")

		// err := StoreClientConfig(&configapi.ClientConfig{})
		// if err != nil {
		// 	fmt.Printf("StoreClientConfig errors: %v\n", err)
		// }
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tests := []struct {
		name    string
		value   string
		persist bool
	}{
		{
			name:    "should persist tanzu when empty client config",
			value:   "tanzu",
			persist: true,
		},
		{
			name:    "should update and persist update-tanzu",
			value:   "update-tanzu",
			persist: true,
		},
		{
			name:    "should not persist same value update-tanzu",
			value:   "update-tanzu",
			persist: false,
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {

			persist, err := SetEdition(spec.value)
			assert.NoError(t, err)
			assert.Equal(t, spec.persist, persist)

			c, err := GetEdition()

			assert.Equal(t, spec.value, c)
			assert.NoError(t, err)

		})
	}
}

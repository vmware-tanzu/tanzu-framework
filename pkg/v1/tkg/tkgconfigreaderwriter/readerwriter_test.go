// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigreaderwriter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

func Test_viperReader_Get(t *testing.T) {
	os.Setenv("FOO", "foo")

	tkgConfigFile := "../fakes/config/config.yaml"

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Read from env",
			args: args{
				key: "FOO",
			},
			want:    "foo",
			wantErr: false,
		},
		{
			name: "Read from file",
			args: args{
				key: "BAR",
			},
			want:    "bar",
			wantErr: false,
		},
		{
			name: "Read from file",
			args: args{
				key: "AWS_B64ENCODED_CREDENTIALS",
			},
			want:    "XXXXXXXX",
			wantErr: false,
		},
		{
			name: "Fails if missing",
			args: args{
				key: "BAZ",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &tkgConfigReaderWriter{}

			err := v.Init(tkgConfigFile)
			if err != nil {
				t.Fatalf("Init() error = %v", err)
			}

			got, err := v.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_viperReader_Set(t *testing.T) {
	dir, err := os.MkdirTemp("", "tkg")
	if err != nil {
		t.Fatalf("os.MkdirTemp() error = %v", err)
	}
	defer os.RemoveAll(dir)

	os.Setenv("FOO", "foo")

	tkgConfigFile := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(tkgConfigFile, []byte("bar: bar"), constants.ConfigFilePermissions); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "",
			args: args{
				key:   "FOO",
				value: "bar",
			},
			want: "bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &tkgConfigReaderWriter{}

			err := v.Init(tkgConfigFile)
			if err != nil {
				t.Fatalf("Init() error = %v", err)
			}

			v.Set(tt.args.key, tt.args.value)

			got, err := v.Get(tt.args.key)
			if err != nil {
				t.Errorf("Get() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v (Set() did not worked)", got, tt.want)
			}
		})
	}
}

func Test_MergeInConfig_Success(t *testing.T) {
	configReaderWriter := newTKGConfigReaderWriter()
	err := configReaderWriter.Init("../fakes/config/config.yaml")
	if err != nil {
		t.Errorf("Expected init success, instead got error: %s", err.Error())
	}
	err = configReaderWriter.MergeInConfig("../fakes/config/config2.yaml")

	if err != nil {
		t.Errorf("Failed merging in config with error %s", err.Error())
	}
}

func Test_MergeInConfig_MissingFile(t *testing.T) {
	configReaderWriter := newTKGConfigReaderWriter()
	err := configReaderWriter.Init("../fakes/config/config.yaml")
	if err != nil {
		t.Errorf("Expected init success, instead got error: %s", err.Error())
	}
	err = configReaderWriter.MergeInConfig("../fakes/config/config1.yaml")

	if err == nil {
		t.Error("Expected error retrieving fakes/config/config1.yaml")
	}
}

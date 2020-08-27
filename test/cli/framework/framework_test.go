package test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanCommand(t *testing.T) {
	for _, test := range []struct {
		name    string
		command string
		res     []string
	}{
		{
			name:    "cli name exists",
			command: "tmc cluster create",
			res:     []string{"cluster", "create"},
		},
		{
			name:    "cli name does not exists",
			command: "cluster create",
			res:     []string{"cluster", "create"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := cleanCommand(test.command)
			require.EqualValues(t, test.res, r)
		})
	}
}

func TestContainsString(t *testing.T) {
	for _, test := range []struct {
		name      string
		stdOut    []byte
		contains  string
		shouldErr bool
	}{
		{
			name:      "basic",
			stdOut:    []byte(`\x1b[2mℹ\x1b[0m using template \"default\"\n\x1b[32m✔\x1b[0m cluster \"test-1571609912\" created successfully \n`),
			contains:  "created successfully",
			shouldErr: false,
		},
		{
			name:      "basic should fail",
			stdOut:    []byte(`a foo bar`),
			contains:  "baz",
			shouldErr: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var b bytes.Buffer
			b.Write(test.stdOut)
			err := ContainsString(b, test.contains)
			if test.shouldErr {
				require.NotNil(t, err, "error should not be nil")
			} else {
				require.Nil(t, err, "error should be nil")
			}
		})
	}
}

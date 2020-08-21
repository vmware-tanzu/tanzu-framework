package cli

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestCreateFlagDescriptor(t *testing.T) {
	type args struct {
		flag *pflag.Flag
	}
	tests := []struct {
		name string
		args args
		want FlagDescriptor
	}{
		{
			name: "flag with no shorthand",
			args: args{
				&pflag.Flag{
					Name:  "flag",
					Usage: "usage",
				},
			},
			want: FlagDescriptor{
				Name:        "--flag",
				Description: "usage",
			},
		},
		{
			name: "flag with shorthand",
			args: args{
				&pflag.Flag{
					Name:      "flag",
					Shorthand: "f",
					Usage:     "usage",
				},
			},
			want: FlagDescriptor{
				Name:        "-f, --flag",
				Description: "usage",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateFlagDescriptor(tt.args.flag)
			require.Equal(t, tt.want, got)
		})
	}
}

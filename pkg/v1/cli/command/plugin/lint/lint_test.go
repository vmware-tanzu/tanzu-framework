// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestTKGTerms(t *testing.T) {
	type test struct {
		cmd            *cobra.Command
		subcmd         *cobra.Command
		terms          *tanzuTerms
		expectedOutput string
		expectedLints  int
	}

	tests := []test{
		{
			cmd: &cobra.Command{
				Use: "RANDOM",
			},
			terms: &tanzuTerms{
				Nouns: []string{"noun"},
			},
			expectedOutput: "unknown top-level term RANDOM, expected standard noun",
			expectedLints:  1,
		},
		{
			cmd: &cobra.Command{
				Use: "noun",
			},
			subcmd: &cobra.Command{
				Use: "RANDOM",
			},
			terms: &tanzuTerms{
				Nouns: []string{"noun"},
				Verbs: []string{"verbin"},
			},
			expectedOutput: "unknown subcommand term RANDOM, expected standard term",
			expectedLints:  1,
		},
		{
			cmd: &cobra.Command{
				Use: "realcmd",
			},
			subcmd: &cobra.Command{
				Use: "realsubcmd",
			},
			terms: &tanzuTerms{
				Nouns: []string{"realcmd"},
				Verbs: []string{"realsubcmd"},
			},
			expectedLints: 0,
		},
	}

	for _, tt := range tests {
		if tt.subcmd != nil {
			tt.cmd.AddCommand(tt.subcmd)
		}
		l := TKGTerms{
			cmd:   tt.cmd,
			nouns: tt.terms.Nouns,
			verbs: tt.terms.Verbs,
		}
		res := l.Execute()
		require.Equal(t, tt.expectedLints, len(*res))
		if tt.expectedLints == 0 {
			continue
		}
		require.Contains(t, (*res)[tt.cmd.Use], tt.expectedOutput)
	}
}

func TestTKGFlags(t *testing.T) {
	type testFlag struct {
		value string
		short string
	}

	var bogus string
	type test struct {
		cmd            *cobra.Command
		flag           *testFlag
		terms          *tanzuTerms
		expectedOutput string
		expectedLints  int
	}

	tests := []test{
		{
			cmd: &cobra.Command{
				Use: "RANDOM",
			},
			flag: &testFlag{
				value: "invalid",
				short: "i",
			},
			terms: &tanzuTerms{
				CmdFlags: []string{"valid"},
			},
			expectedOutput: "unexpected flag invalid, expected standard flag",
			expectedLints:  1,
		}, {
			cmd: &cobra.Command{
				Use: "RANDOM",
			},
			flag: &testFlag{
				value: "valid",
				short: "v",
			},
			terms: &tanzuTerms{
				CmdFlags: []string{"valid"},
			},
			expectedLints: 0,
		},
	}

	for _, tt := range tests {
		if tt.flag != nil {
			tt.cmd.Flags().StringVarP(&bogus, tt.flag.value, "", tt.flag.short, "")
		}
		r := make(Results)
		l := TKGFlags{
			cmd:      tt.cmd,
			cmdFlags: tt.terms.CmdFlags,
			results:  &r,
		}
		res := l.Execute()
		require.Equal(t, tt.expectedLints, len(*res))
		if tt.expectedLints == 0 {
			continue
		}
		require.Contains(t, (*res)[tt.cmd.Use], tt.expectedOutput)
	}
}

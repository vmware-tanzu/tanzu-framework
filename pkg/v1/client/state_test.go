package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestState(t *testing.T) {
	stateTest := State{
		"test": true,
		"foo":  "bar",
	}
	err := SetState(stateTest)
	require.NoError(t, err)

	state, err := GetState()
	require.NoError(t, err)
	require.EqualValues(t, stateTest, state)

	err = DeleteState("test")
	require.NoError(t, err)

	stateTestN := State{
		"foo": "bar",
	}
	state, err = GetState()
	require.NoError(t, err)
	require.Equal(t, stateTestN, state)
}

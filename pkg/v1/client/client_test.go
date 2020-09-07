package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	s, err := GetCurrentServer()
	require.NoError(t, err)

	if s.IsGlobal() {
		
	}
}

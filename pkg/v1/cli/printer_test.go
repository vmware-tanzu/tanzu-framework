package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTableWriter(t *testing.T) {
	tab := NewTableWriter("a", "b", "c")
	require.NotNil(t, tab)
}

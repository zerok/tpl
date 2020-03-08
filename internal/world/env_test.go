package world

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	w := New(&Options{})

	t.Run("single-env", func(t *testing.T) {
		os.Setenv("TEST", "hello world")
		out := requireRender(t, w, `{{ index .Env "TEST" }}`)
		require.Equal(t, "hello world", out)
	})
}

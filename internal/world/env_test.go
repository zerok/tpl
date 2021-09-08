package world

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	w := New(context.Background(), &Options{})

	t.Run("single-env", func(t *testing.T) {
		os.Setenv("TEST", "hello world")
		out := requireRender(t, w, `{{ index .Env "TEST" }}`)
		require.Equal(t, "hello world", out)
	})
}

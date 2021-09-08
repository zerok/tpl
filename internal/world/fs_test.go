package world

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFSExists(t *testing.T) {
	w := New(context.Background(), &Options{})

	t.Run("existing-file", func(t *testing.T) {
		out := requireRender(t, w, `{{ if .FS.Exists "world.go" }}exists{{ else }}not found{{ end }}`)
		require.Equal(t, "exists", out)
	})

	t.Run("not-existing-file", func(t *testing.T) {
		out := requireRender(t, w, `{{ if .FS.Exists "missing-file.go" }}exists{{ else }}not found{{ end }}`)
		require.Equal(t, "not found", out)
	})
}

func TestReadFile(t *testing.T) {
	w := New(context.Background(), &Options{})

	t.Run("existing-file", func(t *testing.T) {
		out := requireRender(t, w, `{{ .FS.ReadFile "world.go" }}`)
		require.Contains(t, out, "type World")
	})

	t.Run("not-existing-file", func(t *testing.T) {
		requireError(t, w, `{{ .FS.ReadFile "i-dont-exist.go" }}`)
	})
}

func requireRender(t *testing.T, w *World, tpl string) string {
	var out bytes.Buffer
	in := bytes.NewBufferString(tpl)
	err := w.Render(&out, in)
	require.NoError(t, err)
	return out.String()
}

func requireError(t *testing.T, w *World, tpl string) {
	var out bytes.Buffer
	in := bytes.NewBufferString(tpl)
	err := w.Render(&out, in)
	require.Error(t, err)
}

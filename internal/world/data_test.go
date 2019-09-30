package world_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/tpl/internal/world"
)

func TestData(t *testing.T) {
	w := world.New(nil)
	data, err := world.LoadData([]string{
		"items=test.yaml",
	}, "../../testdata")
	require.NoError(t, err)
	require.NotNil(t, data)
	w.Data = data
	tmpl := bytes.NewBufferString("{{ range .Data.items }}> {{ . }}\n{{ end }}")
	var out bytes.Buffer
	err = w.Render(&out, tmpl)
	require.NoError(t, err)
	require.Equal(t, "> 1\n> 2\n> 3\n", out.String())
}

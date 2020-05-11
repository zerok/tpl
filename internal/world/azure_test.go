package world_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/tpl/internal/world"
)

func TestAzureSecret(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		w := world.New(nil)
		var out bytes.Buffer
		in := bytes.NewBufferString(`{{ azure "secret/path" "value" }}`)
		err := w.Render(&out, in)
		t.Logf("[[[ %s ]]]", out.String())
		require.Error(t, err)
	})
	t.Run("invalid-token", func(t *testing.T) {
		w := world.New(nil)
		var out bytes.Buffer
		in := bytes.NewBufferString(`{{ .Azure.Secret "secrets--path" | jsonToMap | jmsepathValue "nested.json" }}`)
		err := w.Render(&out, in)
		require.Error(t, err)
		require.NotContains(t, err.Error(), "invalid memory")
	})
}

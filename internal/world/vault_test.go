package world_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zerok/tpl/internal/world"
)

func TestVaultSecret(t *testing.T) {
	w := world.New(context.Background(), nil)
	os.Setenv("VAULT_ADDR", "http://127.0.0.1:54000")
	os.Setenv("VAULT_TOKEN", "")
	var out bytes.Buffer
	in := bytes.NewBufferString("{{ .Vault.Secret \"secret/path\" \"value\" }}")
	err := w.Render(&out, in)
	t.Logf("[[[ %s ]]]", out.String())
	require.Error(t, err)
}

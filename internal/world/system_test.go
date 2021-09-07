package world_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/zerok/tpl/internal/world"
)

// TestSystemShellOutput checks that ShellOutput moves the rendering into an
// error state if the executed command fails.
func TestSystemShellOutput(t *testing.T) {
	w := world.New(context.Background(), &world.Options{
		Insecure: true,
	})
	tests := []struct {
		input   string
		output  string
		errored bool
	}{
		{
			input:   "{{ .System.ShellOutput \"i-dont-exist\" }}",
			output:  "",
			errored: true,
		},
		{
			input:   "{{ .System.ShellOutput \"echo hello\" }}",
			output:  "hello\n",
			errored: false,
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			var out bytes.Buffer
			in := bytes.NewBufferString(test.input)
			err := w.Render(&out, in)
			if test.errored {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.output, out.String())
			}
		})

	}
}

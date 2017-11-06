package world

import (
	"bytes"
	"os"
	"testing"
)

// TestWorldRenderingDelims checks that the delimiters used by the Go template
// engine can be overriden for the given world in order to make working within
// things like JSON easier.
func TestWorldRenderingDelims(t *testing.T) {
	w := New(&Options{
		LeftDelim:  "<",
		RightDelim: ">",
	})
	var out bytes.Buffer
	os.Setenv("TEST", "yes")
	in := bytes.NewBufferString(`< index .Env "TEST" >`)
	err := w.Render(&out, in)
	if err != nil {
		t.Fatalf("rendering shouldn't have resulted in an error. Got %s instead", err.Error())
	}
	if out.String() != "yes" {
		t.Fatalf("Unexpected output: %v", out.String())
	}
}

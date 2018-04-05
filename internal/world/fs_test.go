package world

import (
	"bytes"
	"testing"
)

func TestFSExists(t *testing.T) {
	w := New(&Options{})
	var out bytes.Buffer
	in := bytes.NewBufferString(`{{ if .FS.Exists "world.go" }}exists{{ else }}not found{{ end }}`)
	err := w.Render(&out, in)
	if err != nil {
		t.Fatalf("rendering shouldn't have resulted in an error. Got %s instead", err.Error())
	}
	if out.String() != "exists" {
		t.Fatalf("Unexpected output: %v", out.String())
	}
}

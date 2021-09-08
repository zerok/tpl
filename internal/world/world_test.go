package world

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/jmespath/go-jmespath"
)

// TestWorldRenderingDelims checks that the delimiters used by the Go template
// engine can be overriden for the given world in order to make working within
// things like JSON easier.
func TestWorldRenderingDelims(t *testing.T) {
	w := New(context.Background(), &Options{
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

func TestJsonToMap(t *testing.T) {
	w := New(context.Background(), &Options{})
	data, err := w.jsonToMap(`{ "test": { "key": "value", "key2": "value2" } }`)
	if err != nil {
		t.Fatalf("jsonToMap shouldn't have resulted in an error. Got %s instead", err.Error())
	}
	value, err := jmespath.Search("test.key", data)
	if err != nil {
		t.Fatalf("jmespath shouldn't have resulted in an error. Got %s instead", err.Error())
	}
	if value != "value" {
		t.Fatalf("Unexpected output: %v", value)
	}
}

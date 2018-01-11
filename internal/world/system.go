package world

import (
	"bytes"
	"os/exec"
	"runtime"
)

// System acts as a container for various info points about the system tpl
// is executed on.
type System struct {
	world *World

	// OS represents the name of the operating system as exposed by
	// runtime.GOOS.
	OS string

	// Arch represents the name of the current system architecture as exposed
	// by runtime.GOARCH.
	Arch string
}

// System allows access to various pieces of information regarding the system
// tpl is executed on.
func (w *World) System() *System {
	return &System{
		world: w,
		OS:    runtime.GOOS,
		Arch:  runtime.GOARCH,
	}
}

// ShellOutput starts a shell and executed the given command in it. Note that
// this feature is locked behind the --insecure flag.
func (sys *System) ShellOutput(cmd string) (string, error) {
	if !sys.world.insecure {
		return "", ErrInsecureRequired
	}
	var output bytes.Buffer
	c := exec.Command("/bin/bash", "-c", cmd)
	c.Stdout = &output
	if err := c.Run(); err != nil {
		return "", err
	}
	return output.String(), nil
}

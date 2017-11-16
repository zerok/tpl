package world

import "runtime"

// System acts as a container for various info points about the system tpl
// is executed on.
type System struct {

	// OS represents the name of the operating system as exposed by
	// runtime.GOOS.
	OS string

	// Arch represents the name of the current system architecture as exposed
	// by runtime.GOARCH.
	Arch string
}

// System allows access to various pieces of information regarding the system
// tpl is executed on.
func (w *World) System() System {
	return System{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

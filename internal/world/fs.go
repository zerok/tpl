package world

import "os"

type FS struct {
}

// Exists checks if a given path exists and returns true if it does.
func (fs *FS) Exists(fpath string) bool {
	_, err := os.Stat(fpath)
	if err != nil {
		return false
	}
	return true
}

package world

import (
	"io/ioutil"
	"os"
)

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

// ReadFile returns the content of the given file as string.
func (fs *FS) ReadFile(fpath string) (string, error) {
	fp, err := os.Open(fpath)
	if err != nil {
		return "", err
	}
	defer fp.Close()
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

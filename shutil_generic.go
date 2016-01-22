// +build !linux !cgo

package shutil

import (
	"os"
)

func clonefile(fdst *os.File, fsrc *os.File) (bool, error) {
	return false, nil
}

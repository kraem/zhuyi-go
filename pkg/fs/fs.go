package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func HavePermissions(path string) error {
	ap, err := filepath.Abs(path)
	if err != nil {
		err := fmt.Errorf("%s (%s)", err.Error(), path)
		return err
	}
	if err := checkPermissions(ap); err != nil {
		err := fmt.Errorf("%s (%s)", err.Error(), ap)
		return err
	}
	return nil
}

func checkPermissions(path string) error { return unix.Access(path, unix.W_OK) }

func AppendTrailingSlash(s string) string {
	if string(s[len(s)-1:]) != "/" {
		s = s + "/"
	}
	return s
}

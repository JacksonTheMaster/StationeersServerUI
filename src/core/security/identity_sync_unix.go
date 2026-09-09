//go:build !windows

package security

import "os"

func syncIdentityDirectory(directory string) error {
	dir, err := os.Open(directory)
	if err != nil {
		return err
	}
	err = dir.Sync()
	closeErr := dir.Close()
	if err != nil {
		return err
	}
	return closeErr
}

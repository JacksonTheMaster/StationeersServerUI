//go:build !windows

package security

import "os"

func replaceIdentityFile(source, destination string) error {
	return os.Rename(source, destination)
}

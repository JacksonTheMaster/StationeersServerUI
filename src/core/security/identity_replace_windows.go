//go:build windows

package security

import "golang.org/x/sys/windows"

func replaceIdentityFile(source, destination string) error {
	return windows.Rename(source, destination)
}

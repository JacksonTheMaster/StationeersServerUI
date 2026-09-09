//go:build windows

package security

// Windows does not support syncing directory handles. The identity file itself
// has already been flushed before the atomic rename.
func syncIdentityDirectory(string) error {
	return nil
}

// +build !windows

package config

// Install does nothing when build is not for windows
func Install(installdir string, cli bool) error {
	return nil
}

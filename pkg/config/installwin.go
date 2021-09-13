//go:build windows
// +build windows

package config

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"golang.org/x/sys/windows/registry"
)

// Install installs on windows
func Install(installdir string, cli bool) error {
	var message string

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	if installdir != "" {
		installdir, err = filepath.Abs(installdir)
		if err != nil {
			return err
		}
	}

	if installdir == "" {
		installdir = filepath.Join(home, "AO")
	}

	if !exists(installdir) {
		err = os.Mkdir(installdir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	if strings.Contains(exe, installdir) {
		message = "Cant use AO Install on the installed version"
	} else {

		err = copyInstall(exe, filepath.Join(installdir, "ao.exe"))
		if err != nil {
			message += err.Error() + "\n"
		} else {
			reboot, err := prependPath(installdir)
			if err != nil {
				message += err.Error() + "\n"
			}
			message += "AO installed to " + installdir + "\n"
			if reboot {
				message += "You need to logout and login again for the changes to take effect.\n"
			}
		}
	}
	if cli {
		fmt.Println(message)
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println(message + "\nPress Enter to close")
		reader.ReadString('\n')
	}
	return nil
}

func exists(filePath string) (exists bool) {
	exists = true

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exists = false
	}

	return
}

func prependPath(value string) (bool, error) {
	var key string = "Path"
	existingValue, err := getRegistryKey(key)
	if err != nil {
		return false, err
	}
	if !strings.Contains(existingValue, value) {
		setRegistryKey(key, value+";"+existingValue)
		return true, nil
	}
	return false, nil
}

func setRegistryKey(key string, value string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.WRITE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	err = k.SetStringValue(key, value)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func getRegistryKey(key string) (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	s, _, err := k.GetStringValue(key)
	if err != nil {
		log.Fatal(err)
	}
	return s, nil
}

func copyInstall(source string, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

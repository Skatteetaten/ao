// +build windows

package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/windows/registry"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs the current Windows executable to a folder in the users HOME directory.",
	Long: `Installs the current Windows executable to a folder in the users HOME directory.

Normally, after downloading AO, a file called ao.exe will exists in the 
users Download folder.

The Install command is meant to be run on the downloaded ao.exe by 
opening a command prompt, navigating to the Download folder, and 
issue the ao install command.

This will copy the ao.exe file to a new directory AO in the users 
HOME folder.  It is possible to override this by using the 
--installdir option.

After copying, the command will check if the install directory 
is contained in the users PATH; if not, the PATH will be updated 
and a message written prompting the user to log out and in again
to make the PATH change take effect.`,
	RunE: Install,
}

var installdir string

func init() {
	if runtime.GOOS == "windows" {
		RootCmd.AddCommand(installCmd)
		installCmd.Flags().StringVarP(&installdir, "installdir", "", "", "Override default install location")
	}
}

func Install(cmd *cobra.Command, args []string) error {
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
		err = errors.New("Cant use AO Install on the installed version")
		return err
	}

	err = copyInstall(exe, filepath.Join(installdir, "ao.exe"))
	if err != nil {
		return err
	}

	reboot, err := prependPath(installdir)
	if err != nil {
		return err
	}
	if reboot {
		fmt.Println("You need to logout and login again for the changes to take effect")
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

// +build windows

package cmd

import (
	"runtime"

	"github.com/skatteetaten/ao/pkg/config"
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
	return config.Install(installdir, true)
}

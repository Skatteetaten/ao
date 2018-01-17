package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version of the ao client",
	RunE:  Version,
}

func init() {
	RootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVarP(&flagJSON, "json", "", false, "output version as json")
}

func Version(cmd *cobra.Command, args []string) error {

	if !flagJSON {
		fmt.Println("AO version " + config.Version)
		fmt.Println("Build time " + config.BuildStamp)
		fmt.Println("OS: " + runtime.GOOS)
		executable, err := os.Executable()
		if err == nil {
			fmt.Println("Executable: " + executable)
		}
		return nil
	}

	data, err := json.MarshalIndent(config.DefaultAOVersion, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}

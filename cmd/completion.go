package cmd

import (
	"fmt"

	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion file",
	Long: `This command generates a bash script file that provides bash completion.
Bash completion allows you to press the Tab key to complete keywords.
After running this command, a file called ao.sh will exist in your home directory.
By typing

	source ./ao.sh

you will have bash completion in ao

To persist this across login sessions, please update your .bashrc file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := RootCmd.GenBashCompletionFile("ao.sh"); err == nil {
			wd, _ := os.Getwd()
			fmt.Println("Bash completion file created at", wd+"/ao.sh")
		} else {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(completionCmd)
}

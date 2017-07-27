package cmd

import (
	"fmt"
	"github.com/skatteetaten/aoc/pkg/deletecmd"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var deleteCmd = &cobra.Command{
	Use:   "delete vault <vaultname> | secret <vaultname> <secretname>",
	Short: "Delete a resource",
	Long:  `Delete a resource from the repository.`,

	Run: func(cmd *cobra.Command, args []string) {
		var deletecmdObject deletecmd.DeletecmdClass
		output, err := deletecmdObject.DeleteObject(args, &persistentOptions)
		if err != nil {
			l := log.New(os.Stderr, "", 0)
			l.Println(err.Error())
			os.Exit(-1)
		} else {
			if output != "" {
				fmt.Println(output)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}

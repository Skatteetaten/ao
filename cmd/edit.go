package cmd

import (
	"fmt"

	pkgEditCmd "github.com/skatteetaten/ao/pkg/editcmd"
	"github.com/spf13/cobra"
	"github.com/stromland/coprompt"
)

var editcmdObject = &pkgEditCmd.EditcmdClass{
	Configuration: config,
}

var editCmd = &cobra.Command{
	Use:   "edit [env/]file",
	Short: "Edit a single configuration file or a secret in a vault",
	Long:  `Edit a single configuration file or a secret in a vault.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			cmd.Usage()
			return
		}

		if output, err := editcmdObject.FuzzyEditFile(args); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var editFileCmd = &cobra.Command{
	Use:   "file [env/]<filename>",
	Short: "Edit a single configuration file",
	Annotations: map[string]string{
		coprompt.CALLBACK_ANNOTATION: "GetFiles",
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			return
		}

		if output, err := editcmdObject.FuzzyEditFile(args); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var editSecretCmd = &cobra.Command{
	Use:   "secret <vaultname> <secretname>",
	Short: "Edit a secret in a vault",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			return
		}

		if output, err := editcmdObject.EditSecret(args[0], args[1]); err == nil {
			fmt.Println(output)
		} else {
			fmt.Println(err)
		}
	},
}

var editVaultCmd = &cobra.Command{
	Use:   "vault <vaultname>",
	Short: "Edit a vault",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}

		if output, err := editcmdObject.EditVault(args[0]); err == nil {
			fmt.Print(output)
		} else {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editFileCmd)
	editCmd.AddCommand(editSecretCmd)
	editCmd.AddCommand(editVaultCmd)
}

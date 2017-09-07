package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/auroraconfig"

	pkgEditCmd "github.com/skatteetaten/ao/pkg/editcmd"
	"github.com/spf13/cobra"
)

var editcmdObject = &pkgEditCmd.EditcmdClass{
	Configuration: config,
}

var editCmd = &cobra.Command{
	Use:   "edit [env/]file",
	Short: "Edit a single file in the AuroraConfig repository, or a secret in a vault",
	Long: `Edit a single file in the AuroraConfig repository, or a secret in a vault.
The file can be specified using unique shortened name, so given that the file superapp-test/about.json exists, then the command

	ao edit test/about

will edit this file, if there is no other file matching the same shortening.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			cmd.Usage()
			return
		}

		if output, err := editcmdObject.FuzzyEditFile(args); err == nil {
			fmt.Println(output)
			auroraconfig.UpdateLocalRepository(config.GetAffiliation(), aoConfig)
		} else {
			fmt.Println(err)
		}
	},
}

var editFileCmd = &cobra.Command{
	Use:   "file [env/]<filename>",
	Short: "Edit a single configuration file",
	Long: `Edit a single configuration file or a secret in a vault.
The file can be specified using unique shortened name, so given that the file superapp-test/about.json exists, then the command

	ao edit test/about

will edit this file, if there is no other file matching the same shortening.`,
	Annotations: map[string]string{
		CallbackAnnotation: "GetFiles",
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			return
		}

		if output, err := editcmdObject.FuzzyEditFile(args); err == nil {
			fmt.Println(output)
			auroraconfig.UpdateLocalRepository(config.GetAffiliation(), aoConfig)
		} else {
			fmt.Println(err)
		}
	},
}

var editSecretCmd = &cobra.Command{
	Use:   "secret <vaultname> <secretname>",
	Short: "Edit a secret in a vault",
	Long: `This command will edit the content of the given secret in a vault.
If the given vault does not exist, it will be created.
If the given secret does not exist in the vault, it will be created.`,
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
	Long: `This command will edit the content of the given vault.
The editor will present a JSON view of the vault.
The secrets will be presented as Base64 encoded strings.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}

		if output, err := editcmdObject.EditVault(args[0]); err == nil {
			fmt.Println(output)
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
	editVaultCmd.Hidden = true
}

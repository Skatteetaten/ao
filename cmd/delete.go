package cmd

import (
	"fmt"

	"github.com/skatteetaten/ao/pkg/command"
	pkgDelteCmd "github.com/skatteetaten/ao/pkg/deletecmd"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
)

var forceFlag bool
var deletecmdObject = &pkgDelteCmd.DeletecmdClass{
	Configuration: oldConfig,
	Force:         forceFlag,
}

var deleteCmd = &cobra.Command{
	Use:   "delete app <appname> | env <envname> | deployment <envname> <appname> | file <filename> | vault <vaultname>",
	Short: "Delete a resource",
	Long: `Delete a resource from the AuroraConfig repository.
Deleting an app will delete the app from all environments it is deployed to.  If this leaves any environment emtpy, the command will also delete the about.json file in the env folder.
Deleting an environment will delete all the applications in the given env.  If any application is not deployed in another env, the root app.json file is deleted as well.
Deleting a deployment will delete a specific app from a specific environment.  If the app does not exist in another environment, the root app.json file is deleted as well.  If no other apps are deployed in the given environment, the about.json file are deleted as well

Deleting a specific file will only remove the given filename.  None of the chekcs for related files as done with the delete app, delete env or delete deployment will we executed.

The delete file, vault or secret commands will not ask for any confirmation, but the delete app, env and deployment will ask for confirmation for every file deleted.  It is possible to skip a single delete by pressing N,
or to cancel all deletions by pressing C.

Specifying the force flag will suppress the confirmation prompts, and delete all matching files.
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var deleteAppCmd = &cobra.Command{
	Use:   "app <appname>",
	Short: "Delete application",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			return
		}

		err := command.DeleteFilesFor(fuzzy.APP_FILTER, args[0], DefaultApiClient)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Delete success")
		}
	},
}

var deleteEnvCmd = &cobra.Command{
	Use:   "env <envname>",
	Short: "Delete environment",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			fmt.Println(cmd.UseLine())
			return
		}

		err := command.DeleteFilesFor(fuzzy.ENV_FILTER, args[0], DefaultApiClient)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Delete success")
		}
	},
}

var deleteFileCmd = &cobra.Command{
	Use:   "file <filename>",
	Short: "Delete file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println(cmd.UseLine())
			return
		}

		files, err := command.MultiSelectFile(args[0], DefaultApiClient)
		if err != nil {
			fmt.Println(err)
			return
		}

		command.DefaultTablePrinter(command.GetFilesTable(files))
		message := fmt.Sprintf("Do you want to delete %d file(s)?", len(files))
		shouldDelete := prompt.Confirm(message)

		if !shouldDelete {
			return
		}

		err = command.DeleteFiles(files, DefaultApiClient)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Delete success")
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteAppCmd)
	deleteCmd.AddCommand(deleteEnvCmd)
	deleteCmd.AddCommand(deleteFileCmd)

	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "ignore nonexistent files and arguments, never prompt")
}

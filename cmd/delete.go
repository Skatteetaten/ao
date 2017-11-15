package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
	"io"
	"sort"
)

var forceFlag bool

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
	Annotations: map[string]string{"type": "remote"},
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

		err := DeleteFilesFor(fuzzy.APP_FILTER, args[0], DefaultApiClient, cmd.OutOrStdout())
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

		err := DeleteFilesFor(fuzzy.ENV_FILTER, args[0], DefaultApiClient, cmd.OutOrStdout())
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

		files, err := MultiSelectFile(args[0], DefaultApiClient)
		if err != nil {
			fmt.Println(err)
			return
		}

		DefaultTablePrinter(GetFilesTable(files), cmd.OutOrStdout())
		message := fmt.Sprintf("Do you want to delete %d file(s)?", len(files))
		shouldDelete := prompt.Confirm(message)

		if !shouldDelete {
			return
		}

		err = DeleteFiles(files, DefaultApiClient)
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

func MultiSelectFile(search string, api *client.ApiClient) ([]string, error) {
	var files []string
	fileNames, err := api.GetFileNames()
	if err != nil {
		return files, err
	}

	options := fuzzy.SearchForFile(search, fileNames)

	if len(options) > 1 {
		message := fmt.Sprintf("Matched %d files. Which file do you want?", len(options))
		files = prompt.MultiSelect(message, options)
	} else if len(options) == 1 {
		files = []string{options[0]}
	}

	if len(files) == 0 {
		return files, errors.New("No file to edit")
	}

	return files, nil
}

func DeleteFilesFor(mode fuzzy.FilterMode, search string, api *client.ApiClient, out io.Writer) error {

	fileNames, err := api.GetFileNames()
	if err != nil {
		return err
	}

	files, err := findAllFiles(mode, search, fileNames)
	if err != nil {
		return err
	}

	table := GetFilesTable(files)
	DefaultTablePrinter(table, out)
	message := fmt.Sprintf("Do you want to delete %s?", search)
	deleteAll := prompt.Confirm(message)

	if !deleteAll {
		return errors.New("Delete aborted")
	}

	return DeleteFiles(files, api)
}

func DeleteFiles(files []string, api *client.ApiClient) error {

	ac, err := api.GetAuroraConfig()
	if err != nil {
		return err
	}

	for _, file := range files {
		delete(ac.Files, file)
	}

	res, err := api.SaveAuroraConfig(ac)
	if err != nil {
		return err
	}

	if res != nil {
		return errors.New(res.String())
	}

	return nil
}

func findAllFiles(mode fuzzy.FilterMode, search string, fileNames client.FileNames) ([]string, error) {

	matches := fuzzy.FindAllDeploysFor(mode, search, fileNames.GetDeployments())

	if len(matches) == 0 {
		return nil, errors.New("No matches")
	}

	if mode == fuzzy.APP_FILTER {
		matches = append(matches, search)
	} else {
		matches = append(matches, search+"/about")
	}

	sort.Strings(matches)

	var files []string
	for _, m := range matches {
		files = append(files, m+".json")
	}

	return files, nil
}

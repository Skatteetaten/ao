package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/cmd/common"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
	"io"
	"sort"
)

var forceFlag bool

var (
	deleteCmd = &cobra.Command{
		Use:         "delete",
		Short:       "Delete AuroraConfig files",
		Annotations: map[string]string{"type": "remote"},
	}

	deleteAppCmd = &cobra.Command{
		Use:   "app <appname>",
		Short: "Delete application",
		RunE:  DeleteApplication,
	}

	deleteEnvCmd = &cobra.Command{
		Use:   "env <envname>",
		Short: "Delete environment",
		RunE:  DeleteEnvironment,
	}

	deleteFileCmd = &cobra.Command{
		Use:   "file <filename>",
		Short: "Delete file",
		RunE:  DeleteFile,
	}
)

func init() {
	RootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteAppCmd)
	deleteCmd.AddCommand(deleteEnvCmd)
	deleteCmd.AddCommand(deleteFileCmd)

	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "ignore nonexistent files and arguments, never prompt")
}

func DeleteApplication(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Usage()
	}

	err := deleteFilesFor(fuzzy.APP_FILTER, args[0], DefaultApiClient, cmd.OutOrStdout())
	if err != nil {
		return err
	}
	cmd.Println("Delete success")
	return nil
}

func DeleteEnvironment(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Usage()
	}

	err := deleteFilesFor(fuzzy.ENV_FILTER, args[0], DefaultApiClient, cmd.OutOrStdout())
	if err != nil {
		return err
	}
	cmd.Println("Delete success")
	return nil
}

func DeleteFile(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Usage()
	}

	var files []string
	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	options := fuzzy.SearchForFile(args[0], fileNames)

	if len(options) > 1 {
		message := fmt.Sprintf("Matched %d files. Which file do you want?", len(options))
		files = prompt.MultiSelect(message, options)
	} else if len(options) == 1 {
		files = []string{options[0]}
	}

	if len(files) == 0 {
		return errors.New("No file to edit")
	}

	table := common.GetFilesTable(files)
	common.DefaultTablePrinter(table, cmd.OutOrStdout())
	message := fmt.Sprintf("Do you want to delete %d file(s)?", len(files))
	shouldDelete := prompt.Confirm(message)

	if !shouldDelete {
		return nil
	}

	err = deleteFiles(files, DefaultApiClient)
	if err != nil {
		return err
	}

	cmd.Println("Delete success")
	return nil
}

func deleteFilesFor(mode fuzzy.FilterMode, search string, api *client.ApiClient, out io.Writer) error {

	fileNames, err := api.GetFileNames()
	if err != nil {
		return err
	}

	matches := fuzzy.FindAllDeploysFor(mode, search, fileNames.GetDeployments())

	if len(matches) == 0 {
		return errors.New("No matches")
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

	table := common.GetFilesTable(files)
	common.DefaultTablePrinter(table, out)
	message := fmt.Sprintf("Do you want to delete %s?", search)
	deleteAll := prompt.Confirm(message)

	if !deleteAll {
		return errors.New("Delete aborted")
	}

	return deleteFiles(files, api)
}

func deleteFiles(files []string, api *client.ApiClient) error {

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

package cmd

import (
	"fmt"

	"io"
	"sort"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"github.com/spf13/cobra"
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
		Short: "Delete application files from current AuroraConfig, does not affect deployed applications",
		RunE:  DeleteApplication,
	}

	deleteEnvCmd = &cobra.Command{
		Use:   "env <envname>",
		Short: "Delete all files for a single environment from current AuroraConfig, does not affect deployed applications",
		RunE:  DeleteEnvironment,
	}

	deleteFileCmd = &cobra.Command{
		Use:   "file <filename>",
		Short: "Delete a single file from current AuroraConfig",
		RunE:  DeleteFile,
	}
)

func init() {
	RootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteAppCmd)
	deleteCmd.AddCommand(deleteEnvCmd)
	deleteCmd.AddCommand(deleteFileCmd)

	deleteCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "no interactive prompt")
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

	fileNames, err := DefaultApiClient.GetFileNames()
	if err != nil {
		return err
	}

	search := args[0]
	if len(args) == 2 {
		search = fmt.Sprintf("%s/%s", args[0], args[1])
	}

	matches := fuzzy.FindMatches(search, fileNames, true)

	var fileName string
	if len(matches) > 1 {
		return errors.Errorf("Search matched than one file. Search must be more specific.\n%v", matches)
	} else if len(matches) < 1 {
		return errors.New("No file to delete")
	} else {
		fileName = matches[0]
	}

	message := fmt.Sprintf("Do you want to delete %s?", fileName)
	shouldDelete := prompt.Confirm(message, false)

	if !shouldDelete {
		return nil
	}

	err = deleteFiles([]string{fileName}, DefaultApiClient)
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

	matches := fuzzy.FindAllDeploysFor(mode, search, fileNames.GetApplicationIds())

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

	header, rows := GetFilesTable(files)
	DefaultTablePrinter(header, rows, out)
	message := fmt.Sprintf("Do you want to delete %d files?", len(files))

	if shouldDelete := prompt.Confirm(message, false); !shouldDelete {
		return errors.New("delete aborted")
	}

	return deleteFiles(files, api)
}

// TODO: This should be api feature
func deleteFiles(files []string, api *client.ApiClient) error {

	// ac, err := api.GetAuroraConfig()
	// if err != nil {
	// 	return err
	// }

	// for _, file := range files {
	// 	delete(ac.Files, file)
	// }

	// res, err := api.SaveAuroraConfig(ac)
	// if err != nil {
	// 	return err
	// }
	// if res != nil {
	// 	return errors.New(res.String())
	// }

	return nil
}

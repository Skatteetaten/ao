package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
	"sort"
)

func DeleteFilesFor(mode fuzzy.FilterMode, search string, api *client.ApiClient) error {

	fileNames, err := api.GetFileNames()
	if err != nil {
		return err
	}

	files, err := findAllFiles(mode, search, fileNames)
	if err != nil {
		return err
	}

	table := GetFilesTable(files)
	DefaultTablePrinter(table)
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

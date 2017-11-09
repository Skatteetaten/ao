package command

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/fuzzy"
	"github.com/skatteetaten/ao/pkg/prompt"
)

func SelectFile(search string, api *client.ApiClient) (string, error) {
	var fileName string
	fileNames, err := api.GetFileNames()
	if err != nil {
		return fileName, err
	}

	options, err := fuzzy.SearchForFile(search, fileNames)
	if err != nil {
		return fileName, err
	}

	if len(options) > 1 {
		message := fmt.Sprintf("Matched %d files. Which file do you want?", len(options))
		fileName = prompt.Select(message, options)
	} else if len(options) == 1 {
		fileName = options[0]
	}

	if fileName == "" {
		return fileName, errors.New("No files")
	}

	return fileName, nil
}

func MultiSelectFile(search string, api *client.ApiClient) ([]string, error) {
	var files []string
	fileNames, err := api.GetFileNames()
	if err != nil {
		return files, err
	}

	options, err := fuzzy.SearchForFile(search, fileNames)
	if err != nil {
		return files, err
	}

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

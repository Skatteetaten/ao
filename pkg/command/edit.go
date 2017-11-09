package command

import (
	"encoding/json"
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/skatteetaten/ao/pkg/editor"
)

func EditFile(fileName string, api *client.ApiClient) (string, error) {

	file, err := api.GetAuroraConfigFile(fileName)
	if err != nil {
		return "", err
	}

	onSave := func(modified string) ([]string, error) {
		file.Contents = json.RawMessage(modified)
		res, err := api.PutAuroraConfigFile(file)
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res.GetAllErrors(), nil
		}
		return nil, nil
	}

	status, err := editor.Edit(string(file.Contents), file.Name, onSave)
	if err != nil {
		return "", err
	}

	return status, nil
}

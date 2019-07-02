package service

import (
	"errors"
	"fmt"
	"io"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/client"
)

// Get application list
func GetApplications(apiClient client.AuroraConfigClient, pattern, version string, excludes []string, out io.Writer) ([]string, error) {
	filenames, err := apiClient.GetFileNames()
	if err != nil {
		return nil, err
	}

	applications, err := auroraconfig.GetApplicationRefs(filenames, pattern, excludes)
	if err != nil {
		return nil, err
	}

	if version != "" && len(applications) != 0 {
		if len(applications) > 1 {
			return nil, errors.New("Deploy with version does only support one application")
		}

		fileName, err := filenames.Find(applications[0])
		if err != nil {
			return nil, err
		}

		err = updateVersion(apiClient, version, fileName, out)
		if err != nil {
			return nil, err
		}
	}

	return applications, nil
}

// SetValue updates single Aurora Config value
func SetValue(apiClient client.AuroraConfigClient, name, path, value string) (string, error) {
	fileNames, err := apiClient.GetFileNames()
	if err != nil {
		return "", err
	}

	fileName, err := fileNames.Find(name)
	if err != nil {
		return "", err
	}

	op := client.JsonPatchOp{
		OP:    "add",
		Path:  path,
		Value: value,
	}

	if err = op.Validate(); err != nil {
		return "", err
	}

	if err = apiClient.PatchAuroraConfigFile(fileName, op); err != nil {
		return "", err
	}

	return fileName, nil
}

func updateVersion(apiClient client.AuroraConfigClient, version, fileName string, out io.Writer) error {
	path, value := "/version", version

	fileName, err := SetValue(apiClient, fileName, path, value)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "%s has been updated with %s %s\n", fileName, path, value)

	return nil
}

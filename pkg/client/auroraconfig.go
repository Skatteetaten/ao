package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
)

type AuroraConfigClient interface {
	Doer
	GetFileNames() (auroraconfig.FileNames, error)
	GetAuroraConfig() (*auroraconfig.AuroraConfig, error)
	GetAuroraConfigNames() (*auroraconfig.AuroraConfigNames, error)
	PutAuroraConfig(endpoint string, ac *auroraconfig.AuroraConfig) error
	ValidateAuroraConfig(ac *auroraconfig.AuroraConfig, fullValidation bool) (string, error)
	PatchAuroraConfigFile(fileName string, operation auroraconfig.JsonPatchOp) error
	GetAuroraConfigFile(fileName string) (*auroraconfig.AuroraConfigFile, string, error)
	PutAuroraConfigFile(file *auroraconfig.AuroraConfigFile, eTag string) error
}

type (
	auroraConfigFilePayload struct {
		Content string `json:"content"`
	}
)

func (api *ApiClient) GetFileNames() (auroraconfig.FileNames, error) {
	endpoint := fmt.Sprintf("/auroraconfig/%s/filenames", api.Affiliation)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var fileNames auroraconfig.FileNames
	err = response.ParseItems(&fileNames)
	if err != nil {
		return nil, err
	}

	return fileNames, nil
}

func (api *ApiClient) GetAuroraConfig() (*auroraconfig.AuroraConfig, error) {
	endpoint := fmt.Sprintf("/auroraconfig/%s", api.Affiliation)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var ac auroraconfig.AuroraConfig
	err = response.ParseFirstItem(&ac)
	if err != nil {
		return nil, errors.Wrap(err, "aurora config")
	}

	return &ac, nil
}

func (api *ApiClient) GetAuroraConfigNames() (*auroraconfig.AuroraConfigNames, error) {
	endpoint := fmt.Sprintf("/auroraconfignames")

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var acn auroraconfig.AuroraConfigNames
	err = response.ParseItems(&acn)
	if err != nil {
		return nil, errors.Wrap(err, "aurora config names")
	}
	return &acn, nil
}

func (api *ApiClient) PutAuroraConfig(endpoint string, ac *auroraconfig.AuroraConfig) error {

	payload, err := json.Marshal(ac)
	if err != nil {
		return err
	}

	response, err := api.Do(http.MethodPut, endpoint, payload)
	if err != nil {
		return err
	}

	if !response.Success {
		return response.Error()
	}

	return nil
}

func (api *ApiClient) ValidateAuroraConfig(ac *auroraconfig.AuroraConfig, fullValidation bool) (string, error) {
	resourceValidation := "false"
	if fullValidation {
		resourceValidation = "true"
	}
	endpoint := fmt.Sprintf("/auroraconfig/%s/validate?resourceValidation=%s", api.Affiliation, resourceValidation)

	payload, err := json.Marshal(ac)
	if err != nil {
		return "", err
	}

	response, err := api.Do(http.MethodPut, endpoint, payload)
	if err != nil {
		return "", err
	}

	if !response.Success {
		return "", response.Error()
	}

	//for validation you can also have warnings
	warnings, err := response.toErrorResponse()
	if err != nil {
		return "", err
	}

	if warnings != nil {
		return warnings.formatWarnings(), nil
	}

	return "", nil
}

func (api *ApiClient) GetAuroraConfigFile(fileName string) (*auroraconfig.AuroraConfigFile, string, error) {
	endpoint := fmt.Sprintf("/auroraconfig/%s/%s", api.Affiliation, fileName)

	bundle, err := api.DoWithHeader(http.MethodGet, endpoint, nil, nil)
	if err != nil || bundle == nil {
		return nil, "", err
	}

	if !bundle.BooberResponse.Success {
		return nil, "", errors.New("Failed getting file " + fileName)
	}

	var file auroraconfig.AuroraConfigFile
	err = bundle.BooberResponse.ParseFirstItem(&file)
	if err != nil {
		return nil, "", errors.Wrap(err, "aurora config file")
	}

	eTag := bundle.HttpResponse.Header.Get("ETag")

	return &file, eTag, nil
}

func (api *ApiClient) PatchAuroraConfigFile(fileName string, operation auroraconfig.JsonPatchOp) error {
	endpoint := fmt.Sprintf("/auroraconfig/%s/%s", api.Affiliation, fileName)

	_, _, err := api.GetAuroraConfigFile(fileName)
	if err != nil {
		return err
	}

	op, err := json.Marshal([]auroraconfig.JsonPatchOp{operation})
	if err != nil {
		return err
	}

	payload := auroraConfigFilePayload{
		Content: string(op),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	response, err := api.Do(http.MethodPatch, endpoint, data)
	if err != nil {
		return err
	}

	if !response.Success {
		return response.Error()
	}

	return nil
}

func (api *ApiClient) PutAuroraConfigFile(file *auroraconfig.AuroraConfigFile, eTag string) error {
	endpoint := fmt.Sprintf("/auroraconfig/%s/%s", api.Affiliation, file.Name)

	payload := auroraConfigFilePayload{
		Content: string(file.Contents),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var header map[string]string
	if eTag != "" {
		header = map[string]string{
			"If-Match": eTag,
		}
	}

	bundle, err := api.DoWithHeader(http.MethodPut, endpoint, header, data)
	if err != nil || bundle == nil {
		return err
	}

	if !bundle.BooberResponse.Success {
		return bundle.BooberResponse.Error()
	}

	return nil
}

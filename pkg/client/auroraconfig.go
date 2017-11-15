package client

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/collections"
	"net/http"
	"strings"
)

type (
	FileNames []string

	AuroraConfig struct {
		Files    map[string]json.RawMessage `json:"files"`
		Versions map[string]string          `json:"versions"`
	}

	AuroraConfigFile struct {
		Name     string          `json:"name"`
		Version  string          `json:"version"`
		Override bool            `json:"override"`
		Contents json.RawMessage `json:"contents"`
	}

	JsonPatchOp struct {
		OP    string      `json:"op"`
		Path  string      `json:"path"`
		Value interface{} `json:"value"`
	}

	AuroraConfigFilePayload struct {
		Version          string `json:"version"`
		ValidateVersions bool   `json:"validateVersions"`
		Content          string `json:"content"`
	}
)

func NewAuroraConfig() *AuroraConfig {
	return &AuroraConfig{
		Files:    make(map[string]json.RawMessage),
		Versions: make(map[string]string),
	}
}

func (api *ApiClient) GetFileNames() (FileNames, error) {
	endpoint := fmt.Sprintf("/auroraconfig/%s/filenames", api.Affiliation)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var fileNames FileNames
	err = response.ParseItems(&fileNames)
	if err != nil {
		return nil, err
	}

	return fileNames, nil
}

func (api *ApiClient) GetAuroraConfig() (*AuroraConfig, error) {
	endpoint := fmt.Sprintf("/auroraconfig/%s", api.Affiliation)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var ac AuroraConfig
	err = response.ParseFirstItem(&ac)
	if err != nil {
		return nil, errors.Wrap(err, "aurora config")
	}

	return &ac, nil
}

func (api *ApiClient) PutAuroraConfig(endpoint string, ac *AuroraConfig) (*ErrorResponse, error) {

	payload, err := json.Marshal(ac)
	if err != nil {
		return nil, err
	}

	response, err := api.Do(http.MethodPut, endpoint, payload)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return response.ToErrorResponse()
	}

	return nil, nil
}

func (api *ApiClient) SaveAuroraConfig(ac *AuroraConfig) (*ErrorResponse, error) {
	endpoint := fmt.Sprintf("/auroraconfig/%s", api.Affiliation)
	return api.PutAuroraConfig(endpoint, ac)
}

func (api *ApiClient) ValidateAuroraConfig(ac *AuroraConfig) (*ErrorResponse, error) {
	endpoint := fmt.Sprintf("/auroraconfig/%s/validate", api.Affiliation)
	return api.PutAuroraConfig(endpoint, ac)
}

func (api *ApiClient) GetAuroraConfigFile(fileName string) (*AuroraConfigFile, error) {
	endpoint := fmt.Sprintf("/auroraconfigfile/%s/%s", api.Affiliation, fileName)

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, errors.New("Failed getting file " + fileName)
	}

	var file AuroraConfigFile
	err = response.ParseFirstItem(&file)
	if err != nil {
		return nil, errors.Wrap(err, "aurora config file")
	}

	return &file, nil
}

func (api *ApiClient) PatchAuroraConfigFile(fileName string, operation JsonPatchOp) error {
	endpoint := fmt.Sprintf("/auroraconfigfile/%s/%s/", api.Affiliation, fileName)

	file, err := api.GetAuroraConfigFile(fileName)
	if err != nil {
		return err
	}

	op, err := json.Marshal([]JsonPatchOp{operation})
	if err != nil {
		return err
	}

	payload := AuroraConfigFilePayload{
		Version:          file.Version,
		ValidateVersions: true,
		Content:          string(op),
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
		return errors.New(response.Message)
	}

	return nil
}

func (api *ApiClient) PutAuroraConfigFile(file *AuroraConfigFile) (*ErrorResponse, error) {
	endpoint := fmt.Sprintf("/auroraconfigfile/%s/%s", api.Affiliation, file.Name)

	payload := AuroraConfigFilePayload{
		Version:          file.Version,
		Content:          string(file.Contents),
		ValidateVersions: true,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	response, err := api.Do(http.MethodPut, endpoint, data)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return response.ToErrorResponse()
	}

	return nil, nil
}

func (f *AuroraConfigFile) ToPrettyJson() string {

	data, err := json.MarshalIndent(f.Contents, "", "  ")
	if err != nil {
		return ""
	}

	return string(data)
}

func (f FileNames) GetDeployments() []string {
	var filteredFiles []string
	for _, file := range f {
		if strings.ContainsRune(file, '/') && !strings.Contains(file, "about") {
			filteredFiles = append(filteredFiles, strings.TrimSuffix(file, ".json"))
		}
	}
	return filteredFiles
}

func (f FileNames) GetApplications() []string {
	unique := collections.NewStringSet()
	for _, file := range f {
		if !strings.ContainsRune(file, '/') && !strings.Contains(file, "about") {
			unique.Add(strings.TrimSuffix(file, ".json"))
		}
	}
	return unique.All()
}

func (f FileNames) GetEnvironments() []string {
	unique := collections.NewStringSet()
	for _, file := range f {
		if strings.ContainsRune(file, '/') && !strings.Contains(file, "about") {
			split := strings.Split(file, "/")
			unique.Add(split[0])
		}
	}
	return unique.All()
}

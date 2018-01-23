package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/collections"
)

var (
	ErrJsonPathPrefix = errors.New("json path must start with /")
)

type (
	FileNames []string

	AuroraConfigNames []string

	AuroraConfig struct {
		Name  string             `json:"name"`
		Files []AuroraConfigFile `json:"files"`
	}

	AuroraConfigFile struct {
		Name     string `json:"name"`
		Contents string `json:"contents"`
	}

	JsonPatchOp struct {
		OP    string      `json:"op"`
		Path  string      `json:"path"`
		Value interface{} `json:"value"`
	}

	auroraConfigFilePayload struct {
		Content string `json:"content"`
	}
)

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

func (api *ApiClient) GetAuroraConfigNames() (*AuroraConfigNames, error) {
	endpoint := fmt.Sprintf("/auroraconfignames")

	response, err := api.Do(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var acn AuroraConfigNames
	err = response.ParseItems(&acn)
	if err != nil {
		return nil, errors.Wrap(err, "aurora config names")
	}
	return &acn, nil
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

func (api *ApiClient) ValidateAuroraConfig(ac *AuroraConfig, fullValidation bool) (*ErrorResponse, error) {
	resourceValidation := "false"
	if fullValidation {
		resourceValidation = "true"
	}
	endpoint := fmt.Sprintf("/auroraconfig/%s/validate?resourceValidation=%s", api.Affiliation, resourceValidation)
	return api.PutAuroraConfig(endpoint, ac)
}

func (api *ApiClient) GetAuroraConfigFile(fileName string) (*AuroraConfigFile, string, error) {
	endpoint := fmt.Sprintf("/auroraconfig/%s/%s", api.Affiliation, fileName)

	bundle, err := api.DoWithHeader(http.MethodGet, endpoint, nil, nil)
	if err != nil || bundle == nil {
		return nil, "", err
	}

	if !bundle.BooberResponse.Success {
		return nil, "", errors.New("Failed getting file " + fileName)
	}

	var file AuroraConfigFile
	err = bundle.BooberResponse.ParseFirstItem(&file)
	if err != nil {
		return nil, "", errors.Wrap(err, "aurora config file")
	}

	eTag := bundle.HttpResponse.Header.Get("ETag")

	return &file, eTag, nil
}

func (api *ApiClient) PatchAuroraConfigFile(fileName string, operation JsonPatchOp) (*ErrorResponse, error) {
	endpoint := fmt.Sprintf("/auroraconfig/%s/%s/", api.Affiliation, fileName)

	_, _, err := api.GetAuroraConfigFile(fileName)
	if err != nil {
		return nil, err
	}

	op, err := json.Marshal([]JsonPatchOp{operation})
	if err != nil {
		return nil, err
	}

	payload := auroraConfigFilePayload{
		Content: string(op),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	response, err := api.Do(http.MethodPatch, endpoint, data)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return response.ToErrorResponse()
	}

	return nil, nil
}

func (api *ApiClient) PutAuroraConfigFile(file *AuroraConfigFile, eTag string) (*ErrorResponse, error) {
	endpoint := fmt.Sprintf("/auroraconfig/%s/%s", api.Affiliation, file.Name)

	payload := auroraConfigFilePayload{
		Content: string(file.Contents),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var header map[string]string
	if eTag != "" {
		header = map[string]string{
			"If-Match": eTag,
		}
	}

	bundle, err := api.DoWithHeader(http.MethodPut, endpoint, header, data)
	if err != nil || bundle == nil {
		return nil, err
	}

	if !bundle.BooberResponse.Success {
		return bundle.BooberResponse.ToErrorResponse()
	}

	return nil, nil
}

func (f *AuroraConfigFile) ToPrettyJson() string {

	var out map[string]interface{}
	err := json.Unmarshal([]byte(f.Contents), &out)
	if err != nil {
		return ""
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return ""
	}

	return string(data)
}

func (f FileNames) GetApplicationIds() []string {
	var filteredFiles []string
	for _, file := range f {
		if strings.ContainsRune(file, '/') && !strings.Contains(file, "about") {
			filteredFiles = append(filteredFiles, strings.TrimSuffix(file, ".json"))
		}
	}
	sort.Strings(filteredFiles)
	return filteredFiles
}

func (f FileNames) GetApplications() []string {
	unique := collections.NewStringSet()
	for _, file := range f {
		if !strings.ContainsRune(file, '/') && !strings.Contains(file, "about") {
			unique.Add(strings.TrimSuffix(file, ".json"))
		}
	}
	filteredFiles := unique.All()
	sort.Strings(filteredFiles)
	return filteredFiles
}

func (f FileNames) GetEnvironments() []string {
	unique := collections.NewStringSet()
	for _, file := range f {
		if strings.ContainsRune(file, '/') && !strings.Contains(file, "about") {
			split := strings.Split(file, "/")
			unique.Add(split[0])
		}
	}
	filteredFiles := unique.All()
	sort.Strings(filteredFiles)
	return filteredFiles
}

func (op JsonPatchOp) Validate() error {
	if !strings.HasPrefix(op.Path, "/") {
		return ErrJsonPathPrefix
	}
	return nil
}

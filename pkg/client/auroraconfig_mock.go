package client

import (
	"github.com/pkg/errors"
)

// AuroraConfigClientMock is a base mock type
type AuroraConfigClientMock struct {
	APIClientMock
	files []string
}

// NewAuroraConfigClientMock returns a new AurorConfigClientMock
func NewAuroraConfigClientMock(fileNames FileNames) *AuroraConfigClientMock {
	return &AuroraConfigClientMock{files: fileNames}
}

// GetFileNames default mock implementation
func (api *AuroraConfigClientMock) GetFileNames() (FileNames, error) {
	return api.files, nil
}

// GetAuroraConfig default mock implementation
func (api *AuroraConfigClientMock) GetAuroraConfig() (*AuroraConfig, error) {
	return nil, errors.New("Not implemented")
}

// GetAuroraConfigNames default mock implementation
func (api *AuroraConfigClientMock) GetAuroraConfigNames() (*AuroraConfigNames, error) {
	return nil, errors.New("Not implemented")
}

// PutAuroraConfig default mock implementation
func (api *AuroraConfigClientMock) PutAuroraConfig(endpoint string, ac *AuroraConfig) error {
	return errors.New("Not implemented")
}

// ValidateAuroraConfig default mock implementation
func (api *AuroraConfigClientMock) ValidateAuroraConfig(ac *AuroraConfig, fullValidation bool) error {
	return errors.New("Not implemented")
}

// PatchAuroraConfigFile default mock implementation
func (api *AuroraConfigClientMock) PatchAuroraConfigFile(fileName string, operation JsonPatchOp) error {
	return errors.New("Not implemented")
}

// GetAuroraConfigFile default mock implementation
func (api *AuroraConfigClientMock) GetAuroraConfigFile(fileName string) (*AuroraConfigFile, string, error) {
	return nil, "", errors.New("Not implemented")
}

// PutAuroraConfigFile default mock implementation
func (api *AuroraConfigClientMock) PutAuroraConfigFile(file *AuroraConfigFile, eTag string) error {
	return errors.New("Not implemented")
}

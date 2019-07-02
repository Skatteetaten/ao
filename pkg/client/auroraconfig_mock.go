package client

import (
	"github.com/pkg/errors"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
)

// AuroraConfigClientMock is a base mock type
type AuroraConfigClientMock struct {
	APIClientMock
	files []string
}

// NewAuroraConfigClientMock returns a new AurorConfigClientMock
func NewAuroraConfigClientMock(fileNames auroraconfig.FileNames) *AuroraConfigClientMock {
	return &AuroraConfigClientMock{files: fileNames}
}

// GetFileNames default mock implementation
func (api *AuroraConfigClientMock) GetFileNames() (auroraconfig.FileNames, error) {
	return api.files, nil
}

// GetAuroraConfig default mock implementation
func (api *AuroraConfigClientMock) GetAuroraConfig() (*auroraconfig.AuroraConfig, error) {
	return nil, errors.New("Not implemented")
}

// GetAuroraConfigNames default mock implementation
func (api *AuroraConfigClientMock) GetAuroraConfigNames() (*auroraconfig.AuroraConfigNames, error) {
	return nil, errors.New("Not implemented")
}

// PutAuroraConfig default mock implementation
func (api *AuroraConfigClientMock) PutAuroraConfig(endpoint string, ac *auroraconfig.AuroraConfig) error {
	return errors.New("Not implemented")
}

// ValidateAuroraConfig default mock implementation
func (api *AuroraConfigClientMock) ValidateAuroraConfig(ac *auroraconfig.AuroraConfig, fullValidation bool) error {
	return errors.New("Not implemented")
}

// PatchAuroraConfigFile default mock implementation
func (api *AuroraConfigClientMock) PatchAuroraConfigFile(fileName string, operation auroraconfig.JsonPatchOp) error {
	return errors.New("Not implemented")
}

// GetAuroraConfigFile default mock implementation
func (api *AuroraConfigClientMock) GetAuroraConfigFile(fileName string) (*auroraconfig.AuroraConfigFile, string, error) {
	return nil, "", errors.New("Not implemented")
}

// PutAuroraConfigFile default mock implementation
func (api *AuroraConfigClientMock) PutAuroraConfigFile(file *auroraconfig.AuroraConfigFile, eTag string) error {
	return errors.New("Not implemented")
}

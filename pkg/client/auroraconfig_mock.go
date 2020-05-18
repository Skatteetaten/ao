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
func (api *AuroraConfigClientMock) GetAuroraConfigNames() (*auroraconfig.Names, error) {
	return nil, errors.New("Not implemented")
}

// PutAuroraConfig default mock implementation
func (api *AuroraConfigClientMock) PutAuroraConfig(endpoint string, payload []byte) (string, error) {
	return "", errors.New("Not implemented")
}

// ValidateAuroraConfig default mock implementation
func (api *AuroraConfigClientMock) ValidateAuroraConfig(ac *auroraconfig.AuroraConfig, fullValidation bool) (string, error) {
	return "", errors.New("Not implemented")
}

// GetAuroraConfigFile default mock implementation
func (api *AuroraConfigClientMock) GetAuroraConfigFile(fileName string) (*auroraconfig.File, string, error) {
	return nil, "", errors.New("Not implemented")
}

// PutAuroraConfigFile default mock implementation
func (api *AuroraConfigClientMock) PutAuroraConfigFile(file *auroraconfig.File, eTag string) error {
	return errors.New("Not implemented")
}

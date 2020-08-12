package config

import (
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

func (ao *AOConfig) GetAzureToken() (string, error) {
	deviceConfig := auth.NewDeviceFlowConfig(ao.AzureClientID, ao.AzureTenantID)

	spt, err := deviceConfig.ServicePrincipalToken()
	if err != nil {
		return "", err
	}

	return spt.Token().AccessToken, nil
}

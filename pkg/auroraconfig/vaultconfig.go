package auroraconfig

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

type Vault struct {
	Name        string                      `json:"name"`
	Permissions serverapi.PermissionsStruct `json:"permissions,omitempty"`
	Secrets     map[string]string           `json:"secrets"`
	Versions    map[string]string           `json:"versions,omitempty"`
}

func GetAllVaults(outputFolder string, configuration *configuration.ConfigurationClass) (output string, err error) {
	vaults, err := GetVaultsArray(configuration)
	if err != nil {
		return
	}

	if outputFolder != "" {
		if fileutil.IsLegalFileFolder(outputFolder) == fileutil.SpecIllegal {
			err = errors.New("Illegal file/folder")
			return "", err

		}
	}

	var newline = ""
	for vaultIndex := range vaults {
		content, err := json.Marshal(vaults[vaultIndex])
		contentStr := jsonutil.PrettyPrintJson(string(content))
		output += newline + contentStr
		newline = "\n"
		if outputFolder != "" {
			filename := vaults[vaultIndex].Name + ".json"
			err = fileutil.WriteFile(outputFolder, filename, contentStr)
			if err != nil {
				return "", err
			}
		}
	}

	return output, err
}

func PutVault(vaultname string, vault Vault, version string, configuration *configuration.ConfigurationClass) (validationMessages string, err error) {
	var apiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/vault/"

	content, err := json.Marshal(vault)

	return putContent(apiEndpoint, string(content), version, configuration)

}

func GetVault(vaultname string, configuration *configuration.ConfigurationClass) (vault Vault, err error) {

	vaults, err := GetVaultsArray(configuration)

	if err != nil {
		return vault, err
	}
	for _, vault := range vaults {
		if vault.Name == vaultname {
			return vault, nil
		}
	}

	err = errors.New(vaultname + ": No such vault")
	return vault, err

	/*var apiEndpoint string = "/affiliation/" + configuration.GetAffiliation() + "/vault/" + vaultname

	response, err := serverapi.CallApi(http.MethodGet, apiEndpoint, "", configuration)
	if err != nil {
		output, err := serverapi.ResponsItems2MessageString(response)
		if err != nil {
			return vault, err
		}
		err = errors.New(output)
		return vault, err

	}
	vault, err = serverapi.ResponseItems2Vault(response)

	return vault, err*/
}

func GetVaults(configuration *configuration.ConfigurationClass) (output string, err error) {
	var apiEndpoint string = "/affiliation/" + configuration.GetAffiliation() + "/vault"
	response, err := serverapi.CallApi(http.MethodGet, apiEndpoint, "", configuration)
	if err != nil {
		if !response.Success {
			output, err := serverapi.ResponsItems2MessageString(response)
			if err != nil {
				return output, err
			}
			err = errors.New(output)
			return output, err

		}

		return output, err

	}

	output, err = serverapi.ResponseItems2Vaults(response)

	return output, err
}

func GetVaultRequest(configuration *configuration.ConfigurationClass) (request *serverapi.Request) {
	request = new(serverapi.Request)
	request.Method = http.MethodGet
	request.ApiEndpoint = "/affiliation/" + configuration.GetAffiliation() + "/vault"
	return request
}

func ResponseItems2VaultsArray(response serverapi.Response) (vaults []Vault, err error) {
	vaults = make([]Vault, len(response.Items))

	for item := range response.Items {
		err = json.Unmarshal([]byte(response.Items[item]), &vaults[item])
		if err != nil {
			return
		}
	}
	return
}

func GetVaultsArray(configuration *configuration.ConfigurationClass) (vaults []Vault, err error) {
	var apiEndpoint string = "/affiliation/" + configuration.GetAffiliation() + "/vault"
	response, err := serverapi.CallApi(http.MethodGet, apiEndpoint, "", configuration)
	if err != nil {
		if !response.Success {
			output, err := serverapi.ResponsItems2MessageString(response)
			if err != nil {
				return vaults, err
			}
			err = errors.New(output)
			return vaults, err

		}

	}

	vaults, err = ResponseItems2VaultsArray(response)

	return vaults, err
}

func GetSecret(vaultName string, secretName string, configuration *configuration.ConfigurationClass) (output string, version string, err error) {
	var vaults []Vault
	vaults, err = GetVaultsArray(configuration)

	for vaultindex := range vaults {
		if vaults[vaultindex].Name == vaultName {
			decodedSecret, _ := base64.StdEncoding.DecodeString(vaults[vaultindex].Secrets[secretName])
			output = string(decodedSecret)
			version = vaults[vaultindex].Versions[secretName]
		}
	}
	return
}

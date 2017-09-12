package vault

import (
	"encoding/base64"
	"io/ioutil"
	"testing"
)

const expectedFolderCount = 2

func TestCountFolders(t *testing.T) {
	folderCount, err := countFolders("Testfiles/CountTest")
	if err != nil {
		t.Errorf("Error in folderCount: %v", err.Error())
	}
	if folderCount != expectedFolderCount {
		t.Errorf("folderCount returned unexpected result, expected %v, got %v", expectedFolderCount, folderCount)
	}
}

const expectedVaultArraySize = 3

func TestVaultsFolder2VaultsArray(t *testing.T) {
	vaults, err := vaultsFolder2VaultsArray("Testfiles/ImportTest")
	if err != nil {
		t.Errorf("Error in TestImport: %v", err.Error())
	}
	if len(vaults) != expectedVaultArraySize {
		t.Errorf("vaultsFolder2VaultsArray returned unexpected result, expected %v, got %v.  Vaultname[0] = %v", expectedVaultArraySize, len(vaults), vaults[0].Name)
	}

}

func TestSecretFolder2Vault(t *testing.T) {
	vault, err := secretsFolder2Vault("Testfiles/ImportTest/Vault1")
	if err != nil {
		t.Errorf("Error in TestSecretFolder: %v", err.Error())
	}
	if vault.Name != "Vault1" {
		t.Errorf("secretsFolder2Vault returned unexpected result for name, expected %v, got %v", "Vault1", vault.Name)
	}

	secretContent, err := ioutil.ReadFile("Testfiles/ImportTest/Vault1/latest.properties")
	if err != nil {
		t.Errorf("Error in reading testfile: %v", err.Error())
	}
	secretContent64 := base64.StdEncoding.EncodeToString(secretContent)
	if vault.Secrets["latest.properties"] != secretContent64 {
		t.Errorf("secretsFolder2Vault returned unexpected result for content, expected %v, got %v", secretContent64, vault.Secrets["latest.properties"])
	}

}

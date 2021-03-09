package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiClient_GetVaults(t *testing.T) {
	t.Run("Should get vaults", func(t *testing.T) {
		fileName := "get_vaults_response"
		responseBody := ReadTestFile(fileName)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef("", ts.URL, "test", "sales", "")
		vaults, err := api.GetVaults()

		assert.NoError(t, err)
		assert.Len(t, vaults, 4)
	})
}

func TestApiClient_GetVault(t *testing.T) {
	t.Run("Should get console vault", func(t *testing.T) {
		fileName := "get_vault_console_response"
		responseBody := ReadTestFile(fileName)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef(ts.URL, "", "test", affiliation, "")
		vault, err := api.GetVault("console")
		assert.NoError(t, err)

		assert.Equal(t, "console", vault.Name)
		assert.Len(t, vault.Secrets, 1)
	})
}

func TestApiClient_DeleteVault(t *testing.T) {
	t.Run("Should delete vault", func(t *testing.T) {
		response := []byte("{\"data\":{\"deleteVault\":{\"affiliationName\":\"paas\",\"vaultName\":\"my_test_vault\"}}}")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.DeleteVault("my_test_vault")
		assert.NoError(t, err)
	})
	t.Run("Should fail to delete vault because it does not exist", func(t *testing.T) {
		response := []byte("{\"errors\":[{\"message\":\"Vault not found name=my_test_vault.\",\"locations\":[{\"line\":2,\"column\":3}],\"path\":[\"deleteVault\"],\"extensions\":{\"errorMessage\":\"Vault not found name=my_test_vault.\",\"Korrelasjonsid\":\"\"}}]}")
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.DeleteVault("my_test_vault")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Vault not found name=my_test_vault")
		assert.Contains(t, err.Error(), api.Korrelasjonsid)
	})
}

func TestAPIClient_CreateVault(t *testing.T) {
	t.Run("Should create vault", func(t *testing.T) {
		responseFileName := "createvault_success_response"
		response := ReadTestFile(responseFileName)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		newVault := AuroraSecretVault{
			Name:        "test-vault",
			Permissions: []string{"utv"},
			Secrets: Secrets{
				"latest.properties": "YWJjMTIz",
			},
		}

		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.CreateVault(newVault)

		assert.NoError(t, err)
	})
	t.Run("Should create a vault successfully", func(t *testing.T) {
		response := []byte("{\"data\":{\"createVault\":{\"hasAccess\":true,\"name\":\"my_test_vault\",\"permissions\":[\"APP_PaaS_utv\"]}}}")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		newVault := NewAuroraSecretVault("my_test_vault")
		newVault.Secrets = Secrets{
			"latest.properties": "YWJjMTIz",
		}
		newVault.Permissions = []string{"utv_permission"}
		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.CreateVault(*newVault)

		assert.NoError(t, err)
	})
	t.Run("Should fail to create existing vault", func(t *testing.T) {
		response := []byte("{\"errors\":[{\"message\":\"Vault with vault name my_test_vault already exists.\",\"locations\":[{\"line\":2,\"column\":3}],\"path\":[\"createVault\"],\"extensions\":{\"errorMessage\":\"Vault with vault name my_test_vault already exists.\"}}]}")
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		newVault := NewAuroraSecretVault("my_test_vault")
		newVault.Secrets = Secrets{
			"latest.properties": "YWJjMTIz",
		}
		newVault.Permissions = []string{"utv_permission"}
		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.CreateVault(*newVault)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Vault with vault name my_test_vault already exists")
		assert.Contains(t, err.Error(), api.Korrelasjonsid)
	})
}

func TestApiClient_SaveVault(t *testing.T) {
	t.Run("Should not save vault with no permissions", func(t *testing.T) {

		vault := NewAuroraSecretVault("foo")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef(ts.URL, "", "test", affiliation, "")
		err := api.SaveVault(*vault)
		assert.Error(t, err, "SaveVault should return error when there are no permissions")
	})

	t.Run("Should save vault", func(t *testing.T) {

		vault := NewAuroraSecretVault("foo")
		vault.Permissions = append(vault.Permissions, "someTestGroup")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef(ts.URL, "", "test", affiliation, "")
		err := api.SaveVault(*vault)
		assert.NoError(t, err)
	})
}

func TestAPIClient_RenameVault(t *testing.T) {
	t.Run("Should rename vault OK", func(t *testing.T) {
		response := []byte(`{"data":{"renameVault":{"name":"newname"}}}`)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.RenameVault("oldname", "newname")

		assert.NoError(t, err)
	})
	t.Run("Should fail to rename vault that does not exist", func(t *testing.T) {
		response := []byte(`{"errors":[{"message":"Vault not found name=nonexistingvaultname.","locations":[{"line":3,"column":6}],"path":["renameVault"],"extensions":{"errorMessage":"Vault not found name=nonexistingvaultname.","Korrelasjonsid":"0a3e68bd-8a59-4980-ab1a-c254d0f3f9cd"}}]}`)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.RenameVault("nonexistingvaultname", "newname")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Vault not found name=nonexistingvaultname")
		assert.Contains(t, err.Error(), api.Korrelasjonsid)
	})
	t.Run("Should fail to rename vault to a name that already exists", func(t *testing.T) {
		response := []byte(`{"errors":[{"message":"Vault with vault name existingvaultname already exists.","locations":[{"line":3,"column":6}],"path":["renameVault"],"extensions":{"errorMessage":"Vault with vault name existingvaultname already exists.","Korrelasjonsid":"6e20fe0c-c9de-4552-9332-cd47e4a7219a"}}]}`)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.RenameVault("oldvaultname", "existingvaultname")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Vault with vault name existingvaultname already exists")
		assert.Contains(t, err.Error(), api.Korrelasjonsid)
	})
}

func TestApiClient_UpdateSecretFile(t *testing.T) {
	t.Run("Should update secret file", func(t *testing.T) {

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef(ts.URL, "", "test", affiliation, "")
		err := api.UpdateSecretFile("console", "latest.properties", "", []byte("Rk9PPVRFU1QK"))
		assert.NoError(t, err)
	})
}

func TestAddPermission(t *testing.T) {
	secrets := Secrets{
		"latest.properties": "Rk9PPVRFU1QK",
	}
	secret, err := secrets.GetSecret("latest.properties")
	assert.NoError(t, err)
	assert.Equal(t, secret, "FOO=TEST\n")

	_, err = secrets.GetSecret("latest.properties2")
	assert.Error(t, err)

	secrets.AddSecret("latest.properties2", "FOO=TEST2")
	secret, err = secrets.GetSecret("latest.properties2")
	assert.NoError(t, err)
	assert.Equal(t, secret, "FOO=TEST2")

	secrets.RemoveSecret("latest.properties2")
	_, err = secrets.GetSecret("latest.properties2")
	assert.Error(t, err)
}

func TestAPIClient_AddPermissions(t *testing.T) {
	t.Run("Should add a permission", func(t *testing.T) {
		response := []byte("{\"data\":{\"addVaultPermissions\":{\"hasAccess\":true,\"name\":\"my_test_vault\",\"permissions\":[\"existingpermission\",\"permission\"]}}}")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.AddPermissions("my_test_vault", []string{"permission"})
		assert.NoError(t, err)
	})
	t.Run("Should fail to add a permission because it exist", func(t *testing.T) {
		response := []byte("{\"errors\":[{\"message\":\"Permission [permission] already exists for vault with vault name my_test_vault.\",\"locations\":[{\"line\":2,\"column\":3}],\"path\":[\"addVaultPermissions\"],\"extensions\":{\"errorMessage\":\"Permission [permission] already exists for vault with vault name my_test_vault.\"}}]}")
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.AddPermissions("my_test_vault", []string{"permission"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Permission [permission] already exists for vault with vault name my_test_vault.")
		assert.Contains(t, err.Error(), api.Korrelasjonsid)
	})
}

func TestAPIClient_RemovePermissions(t *testing.T) {
	t.Run("Should remove a permission", func(t *testing.T) {
		response := []byte("{\"data\":{\"removeVaultPermissions\":{\"hasAccess\":true,\"name\":\"my_test_vault\",\"permissions\":[\"existingpermission\"]}}}")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.RemovePermissions("my_test_vault", []string{"permission"})
		assert.NoError(t, err)
	})
	t.Run("Should fail to remove a permission because it is not found", func(t *testing.T) {
		response := []byte("{\"errors\":[{\"message\":\"Permission [permission] does not exist on vault with vault name my_test_vault.\",\"locations\":[{\"line\":2,\"column\":3}],\"path\":[\"removeVaultPermissions\"],\"extensions\":{\"errorMessage\":\"Permission [permission] does not exist on vault with vault name my_test_vault.\"}}]}")
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.RemovePermissions("my_test_vault", []string{"permission"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Permission [permission] does not exist on vault with vault name my_test_vault.")
		assert.Contains(t, err.Error(), api.Korrelasjonsid)
	})
}

func TestAPIClient_AddSecrets(t *testing.T) {
	t.Run("Should add a secret", func(t *testing.T) {
		response := []byte(`{"data":{"addVaultSecrets":{"name":"my_test_vault","secrets":[{"name":"latest.properties"},{"name":"secret.txt"}]}}}`)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()
		var secrets []Secret
		secret := NewSecret("secret.txt", "VGhpcyBpcyBhIHNlY3JldA==")
		secrets = append(secrets, secret)
		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.AddSecrets("my_test_vault", secrets)
		assert.NoError(t, err)
	})
	t.Run("Should fail to add a secret because it exist", func(t *testing.T) {
		response := []byte(`{"errors":[{"message":"Secret [secret.txt] already exists for vault with vault name my_test_vault.","locations":[{"line":2,"column":3}],"path":["addVaultSecrets"],"extensions":{"errorMessage":"Secret [secret.txt] already exists for vault with vault name my_test_vault."}}]}`)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()
		var secrets []Secret
		secret := NewSecret("secret.txt", "VGhpcyBpcyBhIHNlY3JldA==")
		secrets = append(secrets, secret)
		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.AddSecrets("my_test_vault", secrets)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Secret [secret.txt] already exists for vault with vault name my_test_vault")
		assert.Contains(t, err.Error(), api.Korrelasjonsid)
	})
}

func TestAPIClient_RemoveSecrets(t *testing.T) {
	t.Run("Should remove a secret", func(t *testing.T) {
		response := []byte(`{"data":{"removeVaultSecrets":{"name":"my_test_vault","secrets":[{"name":"latest.properties"}]}}}`)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		secretNames := []string{"secret.txt"}
		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.RemoveSecrets("my_test_vault", secretNames)
		assert.NoError(t, err)
	})
	t.Run("Should fail to remove a secret because it does not exist", func(t *testing.T) {
		response := []byte(`{"errors":[{"message":"Secret [secret.txt] does not exist on vault with vault name my_test_vault.","locations":[{"line":2,"column":3}],"path":["removeVaultSecrets"],"extensions":{"errorMessage":"Secret [secret.txt] does not exist on vault with vault name my_test_vault."}}]}`)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))
		defer ts.Close()

		secretNames := []string{"secret.txt"}
		api := NewAPIClientDefaultRef("", ts.URL, "test", affiliation, "")
		err := api.RemoveSecrets("my_test_vault", secretNames)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Secret [secret.txt] does not exist on vault with vault name my_test_vault")
		assert.Contains(t, err.Error(), api.Korrelasjonsid)
	})
}

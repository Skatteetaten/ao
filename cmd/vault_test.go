package cmd

import (
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var (
	vaultTestFolder = "test_files/vault"
	secretFile      = "latest.properties"
	permissionsFile = ".permissions"
)

func TestCreateVault(t *testing.T) {

}

func Test_collectSecrets(t *testing.T) {
	secret := path.Join(vaultTestFolder, secretFile)

	t.Run("should add secret latest.properties from given file to vault 'test'", func(t *testing.T) {
		vault := client.NewAuroraSecretVault("test")

		err := collectSecrets(secret, vault, true)
		assert.NoError(t, err)

		secret, err := vault.Secrets.GetSecret(secretFile)
		assert.NoError(t, err)
		assert.Equal(t, "FOO=BAR\nBAZ=FOOBAR", secret)
	})

	t.Run("should add secret and permission from given folder to vault 'test'", func(t *testing.T) {
		vault := client.NewAuroraSecretVault("test")

		err := collectSecrets(vaultTestFolder, vault, true)
		assert.NoError(t, err)

		secret, err := vault.Secrets.GetSecret(secretFile)
		assert.NoError(t, err)
		assert.Equal(t, "FOO=BAR\nBAZ=FOOBAR", secret)
		assert.Equal(t, []string{"test_group"}, vault.Permissions.GetGroups())
	})
}

func Test_readPermissionFile(t *testing.T) {

	t.Run("should get groups from .permissions file", func(t *testing.T) {
		groups, err := readPermissionFile(path.Join(vaultTestFolder, permissionsFile))

		assert.NoError(t, err)
		assert.Len(t, groups, 1)
		assert.Equal(t, groups[0], "test_group")
	})

	t.Run("should return error when .permissions has no groups", func(t *testing.T) {
		tmp, err := ioutil.TempFile(vaultTestFolder, "permission_")
		defer os.Remove(tmp.Name())
		if err != nil {
			t.Error(err)
		}
		tmp.WriteString("{}")

		groups, err := readPermissionFile(tmp.Name())

		assert.EqualError(t, err, ErrEmptyGroups.Error())
		assert.Empty(t, groups)
	})

	t.Run("should return error when .permissions is illegal file", func(t *testing.T) {
		groups, err := readPermissionFile(vaultTestFolder + "/.test")
		assert.Error(t, err)
		assert.Empty(t, groups)
	})
}

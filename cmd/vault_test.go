package cmd

import (
	"github.com/skatteetaten/ao/pkg/client"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var (
	vaultTestFolder = "test_files/vault_test"
)

func TestCreateVault(t *testing.T) {

}

func Test_collectSecrets(t *testing.T) {
	tmp, err := ioutil.TempFile(vaultTestFolder, "latest.properties_")
	defer os.Remove(tmp.Name())
	if err != nil {
		t.Error(err)
	}
	tmp.WriteString("FOO=BAR\nBAZ=FOOBAR")

	t.Run("should add secret latest.properties from given file to vault 'test'", func(t *testing.T) {
		vault := client.NewAuroraSecretVault("test")

		err = collectSecrets(tmp.Name(), vault)
		assert.NoError(t, err)

		secretName := strings.Split(tmp.Name(), vaultTestFolder+"/")
		secret, err := vault.Secrets.GetSecret(secretName[1])
		assert.NoError(t, err)
		assert.Equal(t, "FOO=BAR\nBAZ=FOOBAR", secret)
	})

	t.Run("should add secret and permission from given folder to vault 'test'", func(t *testing.T) {
		vault := client.NewAuroraSecretVault("test")

		err = collectSecrets(vaultTestFolder, vault)
		assert.NoError(t, err)

		secretName := strings.Split(tmp.Name(), vaultTestFolder+"/")
		secret, err := vault.Secrets.GetSecret(secretName[1])
		assert.NoError(t, err)
		assert.Equal(t, "FOO=BAR\nBAZ=FOOBAR", secret)
		assert.Equal(t, []string{"test_group"}, vault.Permissions.GetGroups())
	})

}

func Test_readPermissionFile(t *testing.T) {

	t.Run("should get groups from .permissions file", func(t *testing.T) {
		groups, err := readPermissionFile(vaultTestFolder + "/.permissions")

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

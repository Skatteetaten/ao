package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/skatteetaten/ao/pkg/client"
	"github.com/stretchr/testify/assert"
)

var (
	vaultTestFolder = "test_files/vault"
	secretFile      = "latest.properties"
	permissionsFile = ".permissions"
)

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
		assert.Equal(t, []string{"test_group"}, vault.Permissions)
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

		assert.EqualError(t, err, errEmptyGroups.Error())
		assert.Empty(t, groups)
	})

	t.Run("should return error when .permissions is illegal file", func(t *testing.T) {
		groups, err := readPermissionFile(vaultTestFolder + "/.test")
		assert.Error(t, err)
		assert.Empty(t, groups)
	})
}

func Test_aggregatePermissions(t *testing.T) {
	type args struct {
		existingGroups []string
		groups         []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Should return an error when trying to add an group that already exists",
			args: args{
				existingGroups: []string{"devops"},
				groups:         []string{"devops"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should add a new group to existingGroups",
			args: args{
				existingGroups: []string{},
				groups:         []string{"devops"},
			},
			want:    []string{"devops"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := aggregatePermissions(tt.args.existingGroups, tt.args.groups)
			if (err != nil) != tt.wantErr {
				t.Errorf("aggregatePermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("aggregatePermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

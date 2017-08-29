package auroraconfig

import (
	"testing"
	"fmt"
	"os/exec"
	"os"
)

func TestValidateRepo(t *testing.T) {

	checkoutPath := "/tmp/ao/testRepo"
	os.MkdirAll(checkoutPath, 0755)
	os.Chdir(checkoutPath)

	gitRemoteUrl := fmt.Sprintf(GIT_URL_FORMAT, "user", "aurora")
	exec.Command("git", "init").Run()
	exec.Command("git", "remote", "add", "origin", gitRemoteUrl).Run()

	if err := ValidateRepo("aurora", "user"); err != nil {
		t.Error(err)
	}
}

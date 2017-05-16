package editcmd

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/serverapi"
)

type EditcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (editcmdClass *EditcmdClass) getAffiliation() (affiliation string) {
	if editcmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = editcmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func (editcmdClass *EditcmdClass) EditFile(args []string) (output string, err error) {
	err = validateEditcmd(args)
	if err != nil {
		return
	}
	if !serverapi.ValidateLogin(editcmdClass.configuration.GetOpenshiftConfig()) {
		return "", errors.New("Not logged in, please use aoc login")
	}

	var affiliation = editcmdClass.getAffiliation()

	var filename string = args[0]
	fmt.Println("Editing " + filename + " in " + affiliation)
	return
}

func validateEditcmd(args []string) (err error) {
	if len(args) != 1 {
		err = errors.New("Usage: aoc edit [env/]file")
		return
	}

	return
}

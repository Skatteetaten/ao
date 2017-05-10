package editcmd

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/configuration"
)

type EditcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (editcmdClass *EditcmdClass) EditFile(args []string) (output string, err error) {
	err = validateEditcmd(args)
	if err != nil {
		return
	}

	var filename string = args[0]
	fmt.Println("Editing " + filename)
	return
}

func validateEditcmd(args []string) (err error) {
	if len(args) != 1 {
		err = errors.New("Usage: aoc edit [env/]file")
		return
	}

	return
}

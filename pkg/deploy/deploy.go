package deploy

import (
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"fmt"
)

func ExecuteDeploy(args []string, persistentOptions *cmdoptions.CommonCommandOptions) (
	output string, error error) {
	fmt.Println("Debug: " + persistentOptions.ListOptions())
	return
}
package deploy

import (
	"fmt"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
)

func ExecuteDeploy(args []string, persistentOptions *cmdoptions.CommonCommandOptions) (
	output string, error error) {
	fmt.Println("Debug: " + persistentOptions.ListOptions())
	return
}

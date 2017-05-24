package pingcmd

import (
	"errors"
	"github.com/skatteetaten/aoc/pkg/cmdoptions"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/serverapi_v2"
)

const usageString = "Usage: aoc ping <address> -p <port> -c <cluster>"

type PingcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (pingcmdClass *PingcmdClass) PingAddress(args []string, pingPort string, pingCluster string, persistentOptions *cmdoptions.CommonCommandOptions) (output string, err error) {
	err = validatePingcmd(args)
	if err != nil {
		return
	}

	var verbose = persistentOptions.Verbose
	var debug = persistentOptions.Debug
	address := args[0]
	argument := "host=" + address
	if pingPort == "" {
		pingPort = "80"
	}
	argument += "&port=" + pingPort
	openshiftConfig := pingcmdClass.configuration.GetOpenshiftConfig()

	_, err = serverapi_v2.CallConsole("netdebug", argument, verbose, debug, openshiftConfig)

	//resultStr := string(result)

	/*fmt.Println(jsonutil.PrettyPrintJson(resultStr))
	fmt.Println("DEBUG")
	fmt.Println(string(result))*/
	return
}

func validatePingcmd(args []string) (err error) {
	if len(args) < 1 {
		err = errors.New(usageString)
		return
	}
	return
}

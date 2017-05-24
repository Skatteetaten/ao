package pingcmd

import (
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/configuration"
)

const usageString = "Usage: aoc ping <address> -p <port> -c <cluster>"

type PingcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (pingcmdClass *PingcmdClass) PingAddress(args []string, pingPort string, pingCluster string) (output string, err error) {
	err = validatePingcmd(args)
	if err != nil {
		return
	}

	openshiftConfig := pingcmdClass.configuration.GetOpenshiftConfig()
	apiCluster := openshiftConfig.APICluster
	fmt.Println("Ping: " + apiCluster)
	return
}

func validatePingcmd(args []string) (err error) {
	if len(args) != 1 {
		err = errors.New(usageString)
		return
	}
	return
}

package pingcmd

import (
	"github.com/skatteetaten/aoc/pkg/configuration"

	"errors"
	"fmt"
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

	var clusterName string
	openshiftConfig := pingcmdClass.configuration.GetOpenshiftConfig()
	apiCluster := openshiftConfig.APICluster


	return
}

func validatePingcmd(args []string) (err error) {
	if len(args) != 1 {
		err = errors.New(usageString)
		return
	}
	return
}

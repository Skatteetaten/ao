package pingcmd

import (
	"errors"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/serverapi_v2"
	"net"
	"sort"
	"strconv"
)

const usageString = "Usage: aoc ping <address> -p <port> -c <cluster>"
const statusOpen = "OPEN"
const statusClosed = "CLOSED"
const partlyClosed = "PARTLY CLOSED"

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

	result, err := serverapi_v2.CallConsole("netdebug", argument, verbose, debug, openshiftConfig)
	if err != nil {
		return
	}

	resultStr := string(result)
	var pingResult serverapi_v2.PingResult
	pingResult, err = serverapi_v2.ParsePingResult(resultStr)
	if err != nil {
		return
	}
	var numberOfOpenHosts = 0
	var numberOfClosedHosts = 0

	var maxHostNameLength = 0
	for hostIndex := range pingResult.Items {
		if pingResult.Items[hostIndex].Result.Status == statusOpen {
			numberOfOpenHosts++
		} else {
			numberOfClosedHosts++
		}
		if persistentOptions.Verbose {
			hostNames, err := net.LookupAddr(pingResult.Items[hostIndex].HostIp)
			if err != nil {
				hostNames = make([]string, 1)
			}
			var hostName string
			if len(hostNames) == 0 {
				hostName = pingResult.Items[hostIndex].HostIp
			} else {
				hostName = hostNames[0][:len(hostNames[0])-1]
			}
			if len(hostName) > maxHostNameLength {
				maxHostNameLength = len(hostName)
			}
			pingResult.Items[hostIndex].HostName = hostName
		}
	}
	if persistentOptions.Verbose {
		var hosts []string
		for hostIndex := range pingResult.Items {
			hosts = append(hosts, pingResult.Items[hostIndex].HostName)
		}
		sort.Strings(hosts)
		for sortedHostIndex := range hosts {
			for hostIndex := range pingResult.Items {
				if pingResult.Items[hostIndex].HostName == hosts[sortedHostIndex] {
					output += "\n\tHost: " + fileutil.RightPad(pingResult.Items[hostIndex].HostName+": ", maxHostNameLength+3) +
						pingResult.Items[hostIndex].Result.Status
				}
			}
		}

	}

	var clusterStatus string
	if numberOfClosedHosts == 0 {
		clusterStatus = statusOpen
	} else {
		if numberOfOpenHosts == 0 {
			clusterStatus = statusClosed
		} else {
			clusterStatus = partlyClosed
		}

	}
	output = address + ":" + pingPort + " is " + clusterStatus + " (reachable by " + strconv.Itoa(numberOfOpenHosts) +
		" of " + strconv.Itoa(numberOfOpenHosts+numberOfClosedHosts) + " hosts)" + output

	return
}

func validatePingcmd(args []string) (err error) {
	if len(args) < 1 {
		err = errors.New(usageString)
		return
	}
	return
}

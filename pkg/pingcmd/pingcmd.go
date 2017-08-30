package pingcmd

import (
	"errors"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/serverapi"
	"net"
	"sort"
	"strconv"
)

const usageString = "Usage: ping <address> -p <port> -c <cluster>"
const statusOpen = "OPEN"
const statusClosed = "CLOSED"
const partlyClosed = "PARTLY CLOSED"

func PingAddress(args []string, pingPort string, pingCluster string, config *configuration.ConfigurationClass) (output string, err error) {
	err = validatePingcmd(args)
	if err != nil {
		return
	}

	var verbose = config.PersistentOptions.Verbose
	var debug = config.PersistentOptions.Debug
	address := args[0]
	argument := "host=" + address
	if pingPort == "" {
		pingPort = "80"
	}
	argument += "&port=" + pingPort

	result, err := serverapi.CallConsole("netdebug", argument, verbose, debug, config.OpenshiftConfig)
	if err != nil {
		return
	}

	resultStr := string(result)
	var pingResult serverapi.PingResult
	pingResult, err = serverapi.ParsePingResult(resultStr)
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
		if verbose {
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
	if verbose {
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

package cmd

import (
	"flag"
	"testing"

	"github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
)

var (
	testCommand = &cobra.Command{}
	updateFiles = flag.Bool("update.files", false, "Update files")
)

func init() {
	testing.Init()
	flag.Parse()
}

func GetDefaultAOConfig() *config.AOConfig {
	clusters := map[string]*config.Cluster{
		"utv": {
			Name:      "utv",
			Reachable: true,
			Url:       "https://utv:8443",
			BooberUrl: "http://boober.utv",
		},
		"relay": {
			Name:      "relay",
			Reachable: true,
			Url:       "https://relay:8443",
			BooberUrl: "http://boober.relay",
		},
		"test": {
			Name:      "test",
			Reachable: false,
			Url:       "https://test:8443",
			BooberUrl: "http://boober.test",
		},
	}

	return &config.AOConfig{
		APICluster:        "utv",
		AvailableClusters: []string{"utv", "relay", "test"},
		Clusters:          clusters,
	}
}

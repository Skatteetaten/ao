package cmd

import (
	"flag"
	"testing"

	"ao/pkg/config"
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
			URL:       "https://utv:8443",
			BooberURL: "http://boober.utv",
			GoboURL:   "http://gobo.utv",
		},
		"relay": {
			Name:      "relay",
			Reachable: true,
			URL:       "https://relay:8443",
			BooberURL: "http://boober.relay",
			GoboURL:   "http://gobo.relay",
		},
		"test": {
			Name:      "test",
			Reachable: false,
			URL:       "https://test:8443",
			BooberURL: "http://boober.test",
			GoboURL:   "http://gobo.test",
		},
	}

	return &config.AOConfig{
		APICluster:        "utv",
		AvailableClusters: []string{"utv", "relay", "test"},
		Clusters:          clusters,
	}
}

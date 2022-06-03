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
		"utv01": {
			Name:      "utv01",
			Reachable: true,
			URL:       "https://utv01:8443",
			BooberURL: "http://boober.utv01",
			GoboURL:   "http://gobo.utv01",
		},
		"relay01": {
			Name:      "relay01",
			Reachable: true,
			URL:       "https://relay01:8443",
			BooberURL: "http://boober.relay01",
			GoboURL:   "http://gobo.relay01",
		},
		"test01": {
			Name:      "test01",
			Reachable: false,
			URL:       "https://test01:8443",
			BooberURL: "http://boober.test01",
			GoboURL:   "http://gobo.test01",
		},
	}

	return &config.AOConfig{
		AvailableClusters: []string{"utv01", "relay01", "test01"},
		Clusters:          clusters,
	}
}

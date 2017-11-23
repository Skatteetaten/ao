package cmd

import (
	"flag"
	"github.com/skatteetaten/ao/pkg/config"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var (
	testCommand = &cobra.Command{}
	updateFiles = flag.Bool("update.files", false, "Update files")
)

func init() {
	flag.Parse()

	clusters := map[string]*config.Cluster{
		"utv": {
			Name:      "utv",
			Reachable: true,
			Url:       "https://utv:8443",
		},
		"relay": {
			Name:      "relay",
			Reachable: true,
			Url:       "https://relay:8443",
		},
		"test": {
			Name:      "test",
			Reachable: false,
			Url:       "https://test:8443",
		},
	}

	AO = &config.AOConfig{
		APICluster:        "utv",
		AvailableClusters: []string{"utv", "relay", "test"},
		Clusters:          clusters,
	}
}

func UpdateTestFile(path string, data []byte) {
	if *updateFiles {
		ioutil.WriteFile(path, data, 666)
	}
}

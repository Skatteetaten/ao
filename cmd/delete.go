package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/serverapi_v2"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources",
}

var deleteSecretCmd = &cobra.Command{
	Use:   "secret",
	Short: "delete secret",
	Example: `
Example:
	aoc delete secret <secretPath> (must be absolute path)

Example:
	aoc delete secret .secret/referanse/latest.properties`,

	Run: func(cmd *cobra.Command, args []string) {

		if len(args) <= 0 {
			fmt.Println(fmt.Errorf("error: no secrets specified"))
			fmt.Println(cmd.Example)
			os.Exit(1)
		}

		var config configuration.ConfigurationClass
		affiliation := config.GetOpenshiftConfig().Affiliation

		endpointUrl := fmt.Sprintf("/affiliation/%s/auroraconfig/secrets", affiliation)

		jsonRequestBody, _ := json.Marshal(args)
		serverapi_v2.CallApiShort(http.MethodDelete, endpointUrl, string(jsonRequestBody), &persistentOptions)
	},
}

func init() {
	deleteCmd.AddCommand(deleteSecretCmd)
	RootCmd.AddCommand(deleteCmd)
}

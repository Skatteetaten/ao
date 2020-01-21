package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func PrintAffiliationsGraphql(cmd *cobra.Command, args []string) {

	configNamesGraphqlRequest := `{auroraApiMetadata{configNames}}`
	type ConfigNamesResponse struct {
		AuroraApiMetadata struct {
			ConfigNames []string
		}
	}

	var configNamesResponse ConfigNamesResponse
	if err := DefaultApiClient.RunGraphQl(configNamesGraphqlRequest, &configNamesResponse); err != nil {
		return
	}

	var mark string
	for _, affiliation := range configNamesResponse.AuroraApiMetadata.ConfigNames {
		if affiliation == AO.Affiliation {
			mark = "*"
		} else {
			mark = " "
		}
		line := fmt.Sprintf("  %s %s", mark, affiliation)
		cmd.Println(line)
	}
}

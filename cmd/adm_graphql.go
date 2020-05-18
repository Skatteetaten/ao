package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// PrintAffiliationsGraphql is the entry point for the `adm affiliations` cli command
func PrintAffiliationsGraphql(cmd *cobra.Command, args []string) error {

	configNamesGraphqlRequest := `{auroraApiMetadata{configNames}}`
	type ConfigNamesResponse struct {
		AuroraAPIMetadata struct {
			ConfigNames []string
		}
	}

	var configNamesResponse ConfigNamesResponse
	if err := DefaultAPIClient.RunGraphQl(configNamesGraphqlRequest, &configNamesResponse); err != nil {
		return err
	}

	var mark string
	for _, affiliation := range configNamesResponse.AuroraAPIMetadata.ConfigNames {
		if affiliation == AO.Affiliation {
			mark = "*"
		} else {
			mark = " "
		}
		line := fmt.Sprintf("  %s %s", mark, affiliation)
		cmd.Println(line)
	}

	return nil
}

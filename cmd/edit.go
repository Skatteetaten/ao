package cmd

import (
	"fmt"
	"strings"

	"github.com/skatteetaten/ao/pkg/auroraconfig"

	pkgEditCmd "github.com/skatteetaten/ao/pkg/editcmd"
	"github.com/spf13/cobra"
	"github.com/renstrom/fuzzysearch/fuzzy"
	"sort"
	"gopkg.in/AlecAivazis/survey.v1"
	"encoding/json"
	"github.com/pkg/errors"
)

var editcmdObject = &pkgEditCmd.EditcmdClass{
	Configuration: config,
}

var editCmd = &cobra.Command{
	Use:   "edit [env/]file",
	Short: "Edit a single file in the AuroraConfig repository, or a secret in a vault",
	Long: `Edit a single file in the AuroraConfig repository, or a secret in a vault.
The file can be specified using unique shortened name, so given that the file superapp-test/about.json exists, then the command

	ao edit test/about

will edit this file, if there is no other file matching the same shortening.`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 1 {
			cmd.Usage()
			return
		}

		if output, err := editcmdObject.FuzzyEditFile(args); err == nil {
			if output != "" {
				fmt.Println(output)
			}
			auroraconfig.UpdateLocalRepository(config.GetAffiliation(), config.OpenshiftConfig)
		} else {
			fmt.Println(err)
		}
	},
}

var editFileCmd = &cobra.Command{
	Use:   "file [env/]<filename>",
	Short: "Edit a single configuration file",
	Long: `Edit a single configuration file or a secret in a vault.
The file can be specified using unique shortened name, so given that the file superapp-test/about.json exists, then the command

	ao edit test/about

will edit this file, if there is no other file matching the same shortening.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			return
		}

		auroraConfig, _ := auroraconfig.GetAuroraConfig(config)

		filename, err := FuzzyFindFile(args[0], auroraConfig.Files)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = pkgEditCmd.EditFile(filename, &auroraConfig, config)
		if err != nil {
			fmt.Println(err)
		}
	},
}

// TODO: Files should be a list of strings
// TODO: Test
func FuzzyFindFile(search string, files map[string]json.RawMessage) (string, error) {
	words := []string{}
	for filename, _ := range files {
		words = append(words, strings.TrimSuffix(filename, ".json"))
	}

	matches := fuzzy.RankFind(strings.TrimSuffix(search, ".json"), words)
	sort.Sort(matches)

	if len(matches) == 0 {
		return "", errors.New("No matches for " + search);
	}


	if (matches.Len() > 0 && matches[0].Distance == 0) || matches.Len() == 1 {
		return matches[0].Target+".json", nil
	}

	options := []string{}
	for _, match := range matches {
		options = append(options, match.Target+".json")
	}

	// TODO: Do we need this?
	if len(options) > 5 {
		sortByName := false
		conf := &survey.Confirm{
			Message: "Do you want to sort by name?",
		}
		survey.AskOne(conf, &sortByName, nil)

		if sortByName {
			sort.Strings(options)
		}
	}

	p := &survey.Select{
		Message: fmt.Sprintf("Matched %d files. Which file do you want to edit?", len(options)),
		PageSize: 10,
		Options: options,
	}

	var filename string
	survey.AskOne(p, &filename, nil)

	return filename, nil
}

var editVaultCmd = &cobra.Command{
	Use:   "vault <vaultname> | <vaultname>/<secretname> | <vaultname> <secretname>",
	Short: "Edit a vault or a secret in a vault",
	Long: `This command will edit the content of the given vault.
The editor will present a JSON view of the vault.
The secrets will be presented as Base64 encoded strings.
If secret-name is given, the editor will present the decoded secret string for editing.`,
	Run: func(cmd *cobra.Command, args []string) {
		var vaultname string
		var secretname string
		var output string
		var err error
		if len(args) == 1 {
			if strings.Contains(args[0], "/") {
				parts := strings.Split(args[0], "/")
				vaultname = parts[0]
				secretname = parts[1]
			} else {
				vaultname = args[0]
			}
		} else if len(args) == 2 {
			vaultname = args[0]
			secretname = args[1]
		} else {
			cmd.Usage()
			return
		}

		if secretname != "" {
			output, err = editcmdObject.EditSecret(vaultname, secretname)
		} else {
			output, err = editcmdObject.EditVault(vaultname)
		}
		if err == nil {
			fmt.Print(output)
		} else {
			fmt.Println(err)
		}

	},
}

func init() {
	RootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editFileCmd)
	editCmd.AddCommand(editVaultCmd)
}

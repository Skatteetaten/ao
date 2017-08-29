package cmd

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
	"sort"
	"strings"
	"encoding/json"
	"github.com/stromland/cobra-prompt"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Interactive shell",
	Run: func(cmd *cobra.Command, args []string) {
		shell := cobraprompt.CobraPrompt{
			RootCmd:                RootCmd,
			DynamicSuggestionsFunc: handleSuggestions,
			GoPromptOptions: []prompt.Option{
				prompt.OptionTitle("Aurora OpenShift cli"),
				prompt.OptionPrefix("ao[" + config.GetAffiliation() + "] "),
				prompt.OptionMaxSuggestion(20),
			},
		}
		shell.Run()
	},
}

func init() {
	RootCmd.AddCommand(shellCmd)
}

func handleSuggestions(annotation string, _ prompt.Document) []prompt.Suggest {
	var suggestions []prompt.Suggest

	switch annotation {
	case "GetFiles":
		return getFiles()
	case "GetDeployments":
		return getDeployments()
	default:
		return suggestions
	}
}

type AuroraConfigFile struct {
	Version string `json:"version"`
}

var acs []prompt.Suggest

func getAuroraConfig() []prompt.Suggest {
	if len(acs) > 0 {
		return acs
	}
	ac, _ := auroraconfig.GetAuroraConfig(config)

	var keys []string
	for k := range ac.Files {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	versions := make(map[string]string)

	for _, k := range keys {
		if strings.Contains(k, "about") {
			continue
		}

		var file AuroraConfigFile
		json.Unmarshal(ac.Files[k], &file)

		if file.Version != "" {
			versions[k] = file.Version
			if strings.Contains(k, "/") {
				acs = append(acs, prompt.Suggest{Text: k, Description: file.Version})
			}
		} else if strings.Contains(k, "/") {
			split := strings.Split(k, "/")
			if versions[split[1]] != "" {
				acs = append(acs, prompt.Suggest{Text: k, Description: versions[split[1]]})
			}

		}
	}

	return acs
}

var filesSuggestions []prompt.Suggest

func getFiles() []prompt.Suggest {
	if len(filesSuggestions) > 0 {
		return filesSuggestions
	}

	files, err := auroraconfig.GetFileList(config)
	sort.Strings(files)
	if err != nil {
		fmt.Println(err)
		return filesSuggestions
	}

	for _, f := range files[1:] {
		filesSuggestions = append(filesSuggestions, prompt.Suggest{Text: f})
	}

	return filesSuggestions
}

var deploymentSuggestions []prompt.Suggest

func getDeployments() []prompt.Suggest {
	if len(deploymentSuggestions) > 0 {
		return deploymentSuggestions
	}

	for _, s := range getAuroraConfig() {
		if strings.Contains(s.Text, "/") && !strings.Contains(s.Text, "about") {
			s.Text = strings.TrimSuffix(s.Text, ".json")
			deploymentSuggestions = append(deploymentSuggestions, s)
		}
	}

	return deploymentSuggestions
}

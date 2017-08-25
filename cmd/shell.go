package cmd

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/spf13/cobra"
	"github.com/stromland/coprompt"
	"sort"
	"strings"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Interactive shell",
	Run: func(cmd *cobra.Command, args []string) {
		shell := coprompt.CoPrompt{
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

	for _, s := range getFiles() {
		if strings.Contains(s.Text, "/") && !strings.Contains(s.Text, "about") {
			s.Text = strings.TrimSuffix(s.Text, ".json")
			deploymentSuggestions = append(deploymentSuggestions, s)
		}
	}

	return deploymentSuggestions
}

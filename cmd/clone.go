package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/src-d/go-git.v4"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"os/user"
	_ "go/token"
	"os"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"github.com/howeyc/gopass"
)

var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone AuroraConfig (git repository) for current affiliation",
	Run: func(cmd *cobra.Command, args []string) {

		var config configuration.ConfigurationClass
		affiliation := config.GetAffiliation()

		if affiliationFlag, _ := cmd.LocalFlags().GetString("affiliation"); len(affiliationFlag) > 0 {
			affiliation = affiliationFlag
		}

		if len(affiliation) == 0 {
			fmt.Println("No affiliation chosen, please login.")
			return
		}

		username, _ := cmd.LocalFlags().GetString("user")
		path, _ := cmd.LocalFlags().GetString("path")

		if len(path) == 0 {
			path = fmt.Sprintf("./%s", affiliation)
		}

		clone(affiliation, username, path)
	},
}

func init() {
	RootCmd.AddCommand(cloneCmd)

	defaultUsername := ""

	if currentUser, err := user.Current(); err == nil {
		defaultUsername = currentUser.Username
	}

	cloneCmd.Flags().StringP("affiliation", "a", "", "Affiliation to clone")
	cloneCmd.Flags().StringP("path", "p", "", "Clone repo to path")
	cloneCmd.Flags().StringP("user", "u", defaultUsername, "Clone repo as user")
}

func clone(affiliation string, username string, path string) {
	url := fmt.Sprintf("https://%s@git.aurora.skead.no/scm/ac/%s.git", username, affiliation)

	fmt.Printf("Cloning AuroraConfig for affiliation %s\n", affiliation)
	fmt.Printf("%s\n\n", url)

	fmt.Printf("Enter password: ")
	password, _ := gopass.GetPasswdMasked()

	fmt.Println()

	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
		Auth:     http.NewBasicAuth(username, string(password)),
	})

	if err != nil {
		fmt.Println(err.Error())
	}
}

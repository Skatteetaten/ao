package cmd

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
)

var addCmd = &cobra.Command{
	Use:         "add <file>",
	Short:       "Add a single file to AuroraConfig",
	Annotations: map[string]string{"type": "remote"},
	RunE:        Add,
}

func init() {
	RootCmd.AddCommand(addCmd)
}

func Add(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Help()
	}

	fileName := args[0]
	file, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	if file.IsDir() {
		return errors.New("Only files are supported")
	}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	if !json.Valid(data) {
		return errors.Errorf("%s contains illegal json format\n", fileName)
	}

	ac, err := DefaultApiClient.GetAuroraConfig()
	if err != nil {
		return err
	}

	path := getValidFileNameFromPath(fileName)

	if _, ok := ac.Files[path]; ok {
		return errors.Errorf("File %s already exists\n", path)
	}

	ac.Files[fileName] = data

	res, err := DefaultApiClient.SaveAuroraConfig(ac)
	if err != nil {
		return err
	}
	if res != nil {
		return errors.New(res.String())
	}

	cmd.Printf("%s has been added\n", fileName)

	return nil
}

func getValidFileNameFromPath(fileName string) string {
	split := strings.Split(fileName, "/")

	if len(split) <= 1 {
		return fileName
	}

	app := split[len(split)-1]
	env := split[len(split)-2]

	if env == "." || env == "~" {
		return app
	}

	return env + "/" + app
}

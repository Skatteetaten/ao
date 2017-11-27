package cmd

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
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

	file, err := os.Stat(args[0])
	if err != nil {
		return err
	}

	if file.IsDir() {
		return errors.New("Only files are supported")
	}

	data, err := ioutil.ReadFile(args[0])
	if err != nil {
		return err
	}

	if !json.Valid(data) {
		return errors.Errorf("%s contains illegal json format", args[0])
	}

	ac, err := DefaultApiClient.GetAuroraConfig()
	if err != nil {
		return err
	}

	ac.Files[args[0]] = data

	res, err := DefaultApiClient.SaveAuroraConfig(ac)
	if err != nil {
		return err
	}
	if res != nil {
		return errors.New(res.String())
	}

	return nil
}

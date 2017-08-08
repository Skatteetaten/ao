package executil

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const commandNotFound = "Command not found"

func RunInteractively(commandString string, foldername string, args ...string) (err error) {
	exepath, err := exec.LookPath(commandString)
	if err != nil {
		return err
	}
	if fileutil.IsLegalFileFolder(exepath) != fileutil.SpecIsFile {
		return errors.New(commandNotFound)
	}

	var command exec.Cmd
	command.Path, err = filepath.Abs(exepath)
	if err != nil {
		return err
	}

	command.Args = make([]string, len(args)+1)
	command.Args[0] = commandString
	for i, arg := range args {
		command.Args[i+1] = arg
	}

	command.Dir, err = filepath.Abs(foldername)
	if err != nil {
		return err
	}

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		return err
	}
	return nil
}

// Prompt for Yes, No or Cancel
func PromptYNC(promptString string) (retval string, err error) {
	var validResponse bool = false
	reader := bufio.NewReader(os.Stdin)
	for !validResponse {
		fmt.Print(promptString + " (Yes/No/Cancel): ")
		result, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		if strings.ContainsAny("YyNnCcYesNoCancel", result) {
			return strings.ToUpper(string(result[0])), nil
		}

	}
	return
}

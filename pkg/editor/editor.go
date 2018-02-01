package editor

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/andybalholm/crlf"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/transform"
)

const (
	invalidJson   = "Invalid JSON format"
	cancelMessage = "Edit cancelled, no changes made."

	editPattern = `## Name: %s
## Please edit the object below. Lines beginning with '##' will be ignored,
## and an empty file will abort the edit. If an error occurs while saving this file will be
## reopened with the relevant failures.
%s%s
`
)

type (
	OnSaveFunc func(modifiedContent string) ([]string, error)

	Editor struct {
		OpenEditor func(string) error
		OnSave     OnSaveFunc
	}
)

func NewEditor(saveFunc OnSaveFunc) *Editor {
	return &Editor{
		OpenEditor: openEditor,
		OnSave:     saveFunc,
	}
}

func (e Editor) Edit(content string, name string) error {

	tempFilePath, err := createTempFile()
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(tempFilePath)
		if err != nil {
			logrus.Fatal("WARNING: Unable to delete temp file: "+err.Error()+": ", tempFilePath)
		}
	}()

	var editErrors string
	originalContent := content
	currentContent := originalContent

	var done bool
	for !done {
		previousContent := currentContent
		contentToEdit := fmt.Sprintf(editPattern, name, editErrors, currentContent)
		if runtime.GOOS == "windows" {
			contentToEdit, _, err = transform.String(crlf.ToCRLF{}, contentToEdit)
		}
		err = ioutil.WriteFile(tempFilePath, []byte(contentToEdit), 0700)
		if err != nil {
			return err
		}

		err = e.OpenEditor(tempFilePath)
		if err != nil {
			return err
		}

		fileContent, err := ioutil.ReadFile(tempFilePath)
		if err != nil {
			return err
		}

		currentContent = stripComments(string(fileContent))
		if previousContent == currentContent {
			return errors.New(cancelMessage)
		}

		validationErrors, err := e.OnSave(currentContent)
		if validationErrors != nil {
			editErrors = addErrorMessage(validationErrors)
		} else if err != nil {
			return err
		} else {
			done = true
		}
	}

	return nil
}

func openEditor(filename string) error {
	var editor = os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vim"
		}
	}

	editorParts := strings.Split(editor, " ")
	editorPath := editorParts[0]

	path, err := exec.LookPath(editorPath)
	if err != nil {
		return errors.New("ERROR: Editor \"" + editorPath + "\" specified in environment variable $EDITOR is not a valid program")
	}

	editorParts[0] = path

	var cmd *exec.Cmd
	cmd = new(exec.Cmd)
	cmd.Path = path
	cmd.Args = append(editorParts, filename)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func createTempFile() (string, error) {
	const tmpFilePrefix = ".ao_edit_file_"
	var tmpDir = os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, tmpFilePrefix)
	if err != nil {
		return "", errors.New("Unable to create temporary file: " + err.Error())
	}
	err = tmpFile.Close()
	if err != nil {
		return "", errors.New("Unable to close temp file: " + err.Error())
	}

	return tmpFile.Name(), nil
}

func stripComments(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var newline = ""
	var actualContent string
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmedLine, "##") {
			actualContent += newline + line
			newline = "\n"
		}
	}

	return actualContent
}

func addErrorMessage(messages []string) string {
	comments := "##\n## ERROR:\n"
	for _, message := range messages {
		for _, line := range strings.Split(message, "\n") {
			comments += fmt.Sprintf("## %s\n", line)
		}
	}

	return comments + "##\n"
}

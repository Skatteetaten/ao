package editor

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	invalidJson   = "Invalid JSON format"
	cancelMessage = "Edit cancelled, no changes made."

	editPattern = `# Name: %s
# Please edit the object below. Lines beginning with a '#' will be ignored,
# and an empty file will abort the edit. If an error occurs while saving this file will be
# reopened with the relevant failures.
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

func (e Editor) Edit(content string, name string, isJson bool) error {

	fmt.Println("DEBUG: createing Temp File")
	tempFilePath, err := createTempFile()
	if err != nil {
		return err
	}
	fmt.Println("DEBUG: Temp File Created")
	defer func() {
		fmt.Println("DEBUG: In defer prior to remove temp file")
		err := os.Remove(tempFilePath)
		if err != nil {
			logrus.Fatal("WARNING: Unable to delete temp file ", tempFilePath)
		}
	}()

	var editErrors string
	originalContent := content
	if isJson {
		originalContent = prettyPrintJson(content)
	}
	currentContent := originalContent

	var done bool
	for !done {
		previousContent := currentContent
		contentToEdit := fmt.Sprintf(editPattern, name, editErrors, currentContent)
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

		if isJson {
			originalHasChanges := hasContentChanged(originalContent, currentContent)
			if !originalHasChanges {
				return errors.New(cancelMessage)
			}

			if !json.Valid([]byte(currentContent)) {
				editErrors = addErrorMessage([]string{invalidJson})
				continue
			}
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
	fmt.Println("DEBUG: Start openEditor")
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
		fmt.Println("DEBUG: Error looking up editor")
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
	fmt.Println("DEBUG: After Editor start")
	return cmd.Wait()
}

func createTempFile() (string, error) {
	const tmpFilePrefix = ".ao_edit_file_"
	var tmpDir = os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, tmpFilePrefix)
	if err != nil {
		return "", errors.New("Unable to create temporary file: " + err.Error())
	}
	return tmpFile.Name(), nil
}

func prettyPrintJson(jsonString string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(jsonString), "", "  ")
	if err != nil {
		return jsonString
	}
	return out.String()
}

func hasContentChanged(original, edited string) bool {

	orgBuffer := new(bytes.Buffer)
	err := json.Compact(orgBuffer, []byte(original))
	if err != nil {
		return true
	}

	editBuffer := new(bytes.Buffer)
	err = json.Compact(editBuffer, []byte(edited))
	if err != nil {
		return true
	}

	return orgBuffer.String() != editBuffer.String()
}

func stripComments(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var newline = ""
	var actualContent string
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmedLine, "#") {
			actualContent += newline + line
			newline = "\n"
		}
	}

	return actualContent
}

func addErrorMessage(messages []string) string {
	comments := "#\n# ERROR:\n"
	for _, message := range messages {
		for _, line := range strings.Split(message, "\n") {
			comments += fmt.Sprintf("# %s\n", line)
		}
	}

	return comments + "#\n"
}

package editor

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const SpecIllegal = -1
const SpecIsFile = 1
const SpecIsFolder = 2

func OpenEditor(filename string) error {
	const vi = "vim"
	var editor = os.Getenv("EDITOR")
	var editorParts []string
	if editor == "" {
		editor = vi
	}
	editorParts = strings.Split(editor, " ")
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
	err = cmd.Wait()

	return err
}

func CreateTempFile() (string, error) {
	const tmpFilePrefix = ".ao_edit_file_"
	var tmpDir = os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, tmpFilePrefix)
	if err != nil {
		return "", errors.New("Unable to create temporary file: " + err.Error())
	}
	if isLegalFileFolder(tmpFile.Name()) != SpecIsFile {
		err = errors.New("Internal error: Illegal temp file name: " + tmpFile.Name())
	}
	filename := tmpFile.Name()
	return filename, nil
}

func isLegalFileFolder(filespec string) int {
	var err error
	var absolutePath string
	var fi os.FileInfo

	absolutePath, err = filepath.Abs(filespec)
	fi, err = os.Stat(absolutePath)
	if os.IsNotExist(err) {
		return SpecIllegal
	} else {
		switch mode := fi.Mode(); {
		case mode.IsDir():
			return SpecIsFolder
		case mode.IsRegular():
			return SpecIsFile
		}
	}
	return SpecIllegal
}

func prettyPrintJson(jsonString string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(jsonString), "", "  ")
	if err != nil {
		return jsonString
	}
	return out.String()
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

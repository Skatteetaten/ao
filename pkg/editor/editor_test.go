package editor

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestEditor_Edit(t *testing.T) {

	const shouldFailError = "foo:shouldFail is not legal"

	// Should end test after two edit cycles
	cases := []struct {
		Content          string
		Edit1            string
		Edit2            string
		ExpectedModified string
		Errors           string
		Err              error
	}{
		// Should not save because no changes
		{"{}", "{}", "{}", "", "", errors.New(cancelMessage)},
		// Should not save because illegal json first, then no changes
		{"{}", `{bar}`, `{}`, "", invalidJson, errors.New(cancelMessage)},

		// Should save, has legal changes
		{"{}", `{"foo":"bar"}`, `{"foo":"bar"}`, `{"foo":"bar"}`, "", nil},
		{"{}", `{"foo":"shouldFail"}`, `{"foo":"bah"}`, `{"foo":"bah"}`, shouldFailError, nil},
	}

	fileEditor := NewEditor(nil)
	fileName := "foo.json"

	for _, tc := range cases {
		currentContent := fmt.Sprintf(editPattern, fileName, "", tc.Content)

		fileEditor.OnSave = func(modifiedContent string) ([]string, error) {

			js := make(map[string]string)
			err := json.Unmarshal([]byte(modifiedContent), &js)
			if err != nil {
				t.Error(err)
			}

			foo := js["foo"]
			if foo == "shouldFail" {
				return []string{shouldFailError}, nil
			}

			assert.Equal(t, tc.ExpectedModified, modifiedContent)

			return nil, nil
		}

		cycle := 0
		fileEditor.OpenEditor = func(tempFile string) error {

			data, err := ioutil.ReadFile(tempFile)
			if err != nil {
				t.Error(err)
			}

			assert.Equal(t, currentContent, string(data))

			messages := addErrorMessage([]string{tc.Errors})
			edit := tc.Edit1
			if cycle == 1 {
				edit = tc.Edit2
			}

			afterEditContent := fmt.Sprintf(editPattern, fileName, messages, edit)
			currentContent = afterEditContent

			ioutil.WriteFile(tempFile, []byte(afterEditContent), 0700)

			cycle++

			return nil
		}

		err := fileEditor.Edit(tc.Content, fileName, true)
		if err != nil {
			assert.EqualError(t, err, tc.Err.Error())
		}
	}
}

func TestHasContentChanged(t *testing.T) {

	testCases := []struct {
		Original string
		Edited   string
		Expected bool
	}{
		{
			Original: `{ "type": "development" }`,
			// Whitespace
			Edited:   `{     "type": "development"          }`,
			Expected: false,
		},
		{
			Original: `{ "type": "development" }`,
			// Newline
			Edited: `{     "type": "development"
			}`,
			Expected: false,
		},
		{
			Original: `{ "type": "deploy" }`,
			// Changed
			Edited:   `{     "type": "development"          }`,
			Expected: true,
		},
		{
			Original: `{ "type": "deploy" }`,
			// Illegal json
			Edited:   `{type": "deploy"}`,
			Expected: true,
		},
	}

	for _, test := range testCases {
		assert.Equal(t, test.Expected, hasContentChanged(test.Original, test.Edited))
	}
}

func TestAddComments(t *testing.T) {

	messages := []string{"FATAL ERROR"}

	expected := "#\n# ERROR:\n# FATAL ERROR\n#\n"
	errs := addErrorMessage(messages)

	assert.Equal(t, expected, errs)
}

func TestStripComments(t *testing.T) {

	content := `# Name: foo.json
{}`

	noComments := stripComments(content)
	assert.Equal(t, "{}", noComments)
}

func TestPrettyPrintJson(t *testing.T) {
	expected := `{
  "foo": "bar"
}`

	actual := prettyPrintJson(`{"foo": "bar"}`)
	assert.Equal(t, expected, actual)

	actual = prettyPrintJson("foo")
	assert.Equal(t, "foo", actual)
}

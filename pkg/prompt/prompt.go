package prompt

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/AlecAivazis/survey.v1"
)

// Password prompts user for password
func Password() string {
	p := &survey.Password{
		Message: "Password:",
	}

	var pass string
	err := survey.AskOne(p, &pass, nil)
	if err != nil {
		logrus.Error(err)
	}

	return string(pass[:])
}

// Confirm prompts user for confirmation
func Confirm(message string, defaultAnswer bool) bool {
	p := &survey.Confirm{
		Message: message,
		Default: defaultAnswer,
	}

	var update bool
	err := survey.AskOne(p, &update, nil)
	if err != nil {
		logrus.Error(err)
	}
	return update
}

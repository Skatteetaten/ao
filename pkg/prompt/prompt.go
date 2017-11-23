package prompt

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/AlecAivazis/survey.v1"
)

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

func Confirm(message string) bool {
	p := &survey.Confirm{
		Message: message,
	}

	var update bool
	err := survey.AskOne(p, &update, nil)
	if err != nil {
		logrus.Error(err)
	}
	return update
}

func MultiSelect(message string, options []string) []string {
	p := &survey.MultiSelect{
		Message:  message,
		PageSize: 10,
		Options:  options,
	}

	var applications []string
	err := survey.AskOne(p, &applications, nil)
	if err != nil {
		logrus.Error(err)
	}

	return applications
}

func Select(message string, options []string) string {

	p := &survey.Select{
		Message:  message,
		PageSize: 10,
		Options:  options,
	}

	var filename string
	err := survey.AskOne(p, &filename, nil)
	if err != nil {
		logrus.Error(err)
	}

	return filename
}

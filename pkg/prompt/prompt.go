package prompt

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/AlecAivazis/survey.v1"
)

func Affiliation(cmd string) string {
	p := &survey.Input{
		Message: cmd + " affiliation:",
	}

	var affiliation string
	err := survey.AskOne(p, &affiliation, nil)
	if err != nil {
		logrus.Error(err)
	}

	return affiliation
}

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

func ConfirmUpdate(version string) bool {
	p := &survey.Confirm{
		Message: fmt.Sprintf("Do you want update to version %s?", version),
	}

	var update bool
	err := survey.AskOne(p, &update, nil)
	if err != nil {
		logrus.Error(err)
	}
	return update
}

func ConfirmDeployAll(applications []string) bool {

	p := &survey.Confirm{
		Message: fmt.Sprintf("Do you want to add all %d application(s) to deploy?", len(applications)),
	}

	var deploy bool
	err := survey.AskOne(p, &deploy, nil)
	if err != nil {
		logrus.Error(err)
	}
	return deploy
}

func ConfirmDeploy(applications []string) bool {

	p := &survey.Confirm{
		Message: fmt.Sprintf("Do you want to deploy %d application(s)?", len(applications)),
		Default: true,
	}

	var deploy bool
	err := survey.AskOne(p, &deploy, nil)
	if err != nil {
		logrus.Error(err)
	}
	return deploy
}

func MultiSelectDeployments(options []string) []string {
	p := &survey.MultiSelect{
		Message:  "Which applications do you want to deploy?",
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

func SelectFile(options []string) string {

	p := &survey.Select{
		Message:  fmt.Sprintf("Matched %d files. Which file do you want?", len(options)),
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

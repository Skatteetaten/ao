package prompt

import (
	"gopkg.in/AlecAivazis/survey.v1"
	"fmt"
	"github.com/sirupsen/logrus"
)

func ConfirmDeploy(applications []string) bool {
	//  TODO: Print deployment table
	p := &survey.Confirm{
		Message: fmt.Sprintf("Do you want to deploy %d applications?", len(applications)),
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
		Message:  fmt.Sprintf("Matched %d files. Which applications do you want to deploy?", len(options)),
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

func SelectFileToEdit(options []string) string {

	p := &survey.Select{
		Message:  fmt.Sprintf("Matched %d files. Which file do you want to edit?", len(options)),
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

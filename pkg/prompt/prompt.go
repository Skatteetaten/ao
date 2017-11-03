package prompt

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/printutil"
	"gopkg.in/AlecAivazis/survey.v1"
	"strings"
)

func ConfirmDeployAll(applications []string) bool {

	printDeployTable(applications)

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

	printDeployTable(applications)

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

func printDeployTable(applications []string) {
	envs := []string{}
	apps := []string{}

	for _, a := range applications {
		split := strings.Split(a, "/")
		envs = append(envs, split[0])
		apps = append(apps, split[1])
	}

	fmt.Printf(printutil.FormatTable([]string{"ENVIRONMENT", "APPLICATION"}, envs, apps))
}

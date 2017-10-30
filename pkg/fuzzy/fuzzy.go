package fuzzy

import (
	"strings"
	"github.com/renstrom/fuzzysearch/fuzzy"
	"sort"
	"gopkg.in/AlecAivazis/survey.v1"
	"fmt"
	"github.com/pkg/errors"
)

func FindMatches(search string, fileNames []string, withSuffix bool) ([]string, error) {

	files := []string{}
	for _, file := range fileNames {
		files = append(files, strings.TrimSuffix(file, ".json"))
	}

	suffix := ""
	if withSuffix {
		suffix = ".json"
	}

	matches := fuzzy.RankFind(strings.TrimSuffix(search, ".json"), files)
	sort.Sort(matches)

	if len(matches) == 0 {
		return []string{}, errors.New("No matches for " + search);
	}

	firstMatch := matches[0]
	if firstMatch.Distance == 0 || len(matches) == 1 {
		return []string{firstMatch.Target + suffix}, nil
	}

	options := []string{}
	for _, match := range matches {
		options = append(options, match.Target+suffix)
	}

	return options, nil
}

func FindFileToEdit(search string, files []string, prompt bool) (string, error) {

	options, err := FindMatches(search, files, true)
	if err != nil {
		return "", err
	}

	if len(options) == 1 || !prompt {
		return options[0], nil
	}

	p := &survey.Select{
		Message:  fmt.Sprintf("Matched %d files. Which file do you want to edit?", len(options)),
		PageSize: 10,
		Options:  options,
	}

	var filename string
	err = survey.AskOne(p, &filename, nil)

	return filename, err
}

// TODO: Exact match for environment - Include all apps
// TODO: Exact match for application - Include all envs
func FindApplicationsToDeploy(search string, files []string, prompt bool) ([]string, error) {

	options, err := FindMatches(search, files, false)
	if err != nil {
		return []string{}, err
	}

	if len(options) == 1 || !prompt {
		return options, nil
	}

	p := &survey.MultiSelect{
		Message:  fmt.Sprintf("Matched %d files. Which applications do you want to deploy?", len(options)),
		PageSize: 10,
		Options:  options,
	}

	var applications []string
	err = survey.AskOne(p, &applications, nil)

	return applications, err
}

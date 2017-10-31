package fuzzy

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/renstrom/fuzzysearch/fuzzy"
	"gopkg.in/AlecAivazis/survey.v1"
	"sort"
	"strings"
)

func FilterFileNamesForDeploy(fileNames []string) []string {
	filteredFiles := []string{}
	for _, file := range fileNames {
		if strings.ContainsRune(file, '/') && !strings.Contains(file, "about") {
			filteredFiles = append(filteredFiles, strings.TrimSuffix(file, ".json"))
		}
	}
	return filteredFiles
}

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
		return []string{}, errors.New("No matches for " + search)
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

func NewDeploySet() *DeploySet {
	return &DeploySet{
		set: make(map[string]bool),
	}
}

type DeploySet struct {
	set map[string]bool
}

func (s *DeploySet) Add(key string) {
	_, found := s.set[key]
	if !found {
		s.set[key] = true
	}
}

func (s *DeploySet) Keys() []string {
	keys := []string{}
	for k, _ := range s.set {
		keys = append(keys, k)
	}

	return keys
}

type DeployFilterMode uint

const (
	APP DeployFilterMode = iota
	ENV
)

/*
	Search string must match exact
 */
func FindAllDeploysFor(mode DeployFilterMode, search string, files []string) ([]string, error) {
	deploys := make(map[string]*DeploySet)

	for _, file := range files {
		deploy := strings.Split(file, "/")
		if len(deploy) != 2 {
			continue
		}

		// Key = env, value = app
		key, value := deploy[0], deploy[1]
		if mode == APP {
			// Key = app, value = env
			key, value = deploy[1], deploy[0]
		}

		if _, found := deploys[key]; !found {
			deploys[key] = NewDeploySet()
			deploys[key].Add(value)
		} else {
			deploys[key].Add(value)
		}
	}

	deployKeys := []string{}
	for k, _ := range deploys {
		deployKeys = append(deployKeys, k)
	}

	matches := fuzzy.RankFind(search, deployKeys)
	sort.Sort(matches)

	if len(matches) < 1 || matches[0].Distance != 0 {
		return []string{}, nil
	}

	result := deploys[matches[0].Target]
	keys := result.Keys()
	sort.Strings(keys)

	return keys, nil
}

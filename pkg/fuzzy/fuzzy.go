package fuzzy

import (
	"github.com/pkg/errors"
	"github.com/renstrom/fuzzysearch/fuzzy"
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

func SearchForFile(search string, files []string) ([]string, error) {

	options, err := FindMatches(search, files, true)
	if err != nil {
		return []string{}, err
	}

	return options, nil
}

func SearchForApplications(search string, files []string) ([]string, error) {

	options := []string{}
	if !strings.Contains(search, "/") {
		options = FindAllDeploysFor(APP_FILTER, search, files)
		if len(options) == 0 {
			options = FindAllDeploysFor(ENV_FILTER, search, files)
		}
	}

	if len(options) == 0 {
		opts, err := FindMatches(search, files, false)
		if err != nil {
			return []string{}, err
		}
		options = opts
	}

	return options, nil
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
	APP_FILTER DeployFilterMode = iota
	ENV_FILTER
)

/*
	Search string must match either environment or application exact
 */
func FindAllDeploysFor(mode DeployFilterMode, search string, files []string) []string {
	deploys := make(map[string]*DeploySet)

	for _, file := range files {
		appId := strings.Split(file, "/")
		if len(appId) != 2 {
			continue
		}

		// Key = env, value = app
		key, value := appId[0], appId[1]
		if mode == APP_FILTER {
			// Key = app, value = env
			key, value = appId[1], appId[0]
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
		return []string{}
	}

	result := deploys[matches[0].Target]
	keys := result.Keys()
	sort.Strings(keys)

	allDeploys := []string{}
	for _, k := range keys {
		id := search + "/" + k
		if mode == APP_FILTER {
			id = k + "/" + search
		}

		allDeploys = append(allDeploys, id)
	}

	return allDeploys
}

package auroraconfig

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/skatteetaten/ao/pkg/collections"
)

// FilterMode is a numeric filtering parameter
type FilterMode uint

// Filters for fuzzy matching
const (
	AppFilter FilterMode = iota
	EnvFilter
	Separator = "/"
)

// FindMatches finds filenames by fuzzy matching
func FindMatches(search string, fileNames []string, withSuffix bool) []string {
	trimmedSearch := strings.TrimSuffix(search, filepath.Ext(search))
	files := FileNames(fileNames)
	matches := fuzzy.RankFind(trimmedSearch, files.WithoutExtension())
	matches = filterSplitMatches(trimmedSearch, matches)
	if len(matches) == 0 {
		return []string{}
	}

	sort.Sort(matches)

	firstMatch := matches[0]
	if firstMatch.Distance == 0 || len(matches) == 1 {
		fileName := firstMatch.Target
		if withSuffix {
			fileName, _ = files.Find(firstMatch.Target)
		}
		return []string{fileName}
	}

	options := []string{}
	for _, match := range matches {
		fileName := match.Target
		if withSuffix {
			fileName, _ = files.Find(match.Target)
		}
		options = append(options, fileName)
	}

	return options
}

// When search string does not contain "/", it should only be a match with either path or file name,
// not fuzzy matched with the complete string.
func filterSplitMatches(search string, matches fuzzy.Ranks) fuzzy.Ranks {
	if !(strings.Contains(search, Separator)) {
		var filteredMatches fuzzy.Ranks
		for _, match := range matches {
			fileName := match.Target
			if strings.Contains(fileName, Separator) {
				s := strings.SplitN(fileName, Separator, 2)
				matches := fuzzy.RankFind(search, s)
				if len(matches) == 0 {
					continue
				}
			}
			filteredMatches = append(filteredMatches, match)
		}
		return filteredMatches
	}
	return matches
}

// SearchForFile finds filenames by fuzzy matching
func SearchForFile(search string, files []string) []string {
	return FindMatches(search, files, true)
}

// SearchForApplications finds application deployments by fuzzy matching
func SearchForApplications(search string, files []string) []string {
	var options []string
	if !strings.Contains(search, "/") {
		options = FindAllDeploysFor(AppFilter, search, files)
		if len(options) == 0 {
			options = FindAllDeploysFor(EnvFilter, search, files)
		}
	}

	if len(options) == 0 {
		options = FindMatches(search, files, false)
	}

	return options
}

// FindAllDeploysFor finds deployments by FilterMode and a search string  Search string must match either environment or application exact
func FindAllDeploysFor(mode FilterMode, search string, files []string) []string {
	search = strings.TrimSuffix(search, filepath.Ext(search))
	deploys := make(map[string]*collections.StringSet)

	for _, file := range files {
		appID := strings.Split(file, "/")
		if len(appID) != 2 {
			continue
		}

		// Key = env, value = app
		key, value := appID[0], appID[1]
		if mode == AppFilter {
			// Key = app, value = env
			key, value = appID[1], appID[0]
		}

		if _, found := deploys[key]; !found {
			deploys[key] = collections.NewStringSet()
			deploys[key].Add(value)
		} else {
			deploys[key].Add(value)
		}
	}

	var deployKeys []string
	for k := range deploys {
		deployKeys = append(deployKeys, k)
	}

	matches := fuzzy.RankFind(search, deployKeys)
	sort.Sort(matches)

	if len(matches) < 1 || matches[0].Distance != 0 {
		return []string{}
	}

	result := deploys[matches[0].Target]
	keys := result.All()
	sort.Strings(keys)

	var allDeploys []string
	for _, k := range keys {
		id := search + "/" + k
		if mode == AppFilter {
			id = k + "/" + search
		}

		allDeploys = append(allDeploys, id)
	}

	return allDeploys
}

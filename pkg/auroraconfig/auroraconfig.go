package auroraconfig

import (
	"encoding/json"
	"regexp"
	"strings"
)

type (
	// Names of AuroraConfigs
	Names []string

	// AuroraConfig is a named structure of files forming an Aurora configuration
	AuroraConfig struct {
		Name  string `json:"name"`
		Files []File `json:"files"`
	}

	// File is a file in an Aurora configuration
	File struct {
		Name     string `json:"name"`
		Contents string `json:"contents"`
	}
)

// GetApplicationRefs
func GetApplicationRefs(filenames FileNames, pattern string, excludes []string) ([]string, error) {
	possibleDeploys := filenames.GetApplicationDeploymentRefs()
	applications := SearchForApplications(pattern, possibleDeploys)

	applications, err := filterExcludes(excludes, applications)
	if err != nil {
		return nil, err
	}

	return applications, nil
}

func (f *File) ToPrettyJson() string {

	var out map[string]interface{}
	err := json.Unmarshal([]byte(f.Contents), &out)
	if err != nil {
		return ""
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return ""
	}

	return string(data)
}

func (f *File) IsYaml() bool {
	return (strings.HasSuffix(strings.ToLower(f.Name), ".yaml") || (strings.HasSuffix(strings.ToLower(f.Name), ".yml")))
}

func filterExcludes(expressions, applications []string) ([]string, error) {
	apps := make([]string, len(applications))
	copy(apps, applications)
	for _, expr := range expressions {
		r, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}
		tmp := apps[:0]
		for _, app := range apps {
			match := r.MatchString(app)
			if !match {
				tmp = append(tmp, app)
			}
		}
		apps = tmp
	}

	return apps, nil
}

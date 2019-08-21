package auroraconfig

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

type (
	AuroraConfigNames []string

	AuroraConfig struct {
		Name  string             `json:"name"`
		Files []AuroraConfigFile `json:"files"`
	}

	AuroraConfigFile struct {
		Name     string `json:"name"`
		Contents string `json:"contents"`
	}

	JsonPatchOp struct {
		OP    string      `json:"op"`
		Path  string      `json:"path"`
		Value interface{} `json:"value"`
	}
)

var (
	ErrJsonPathPrefix = errors.New("json path must start with /")
)

func GetApplicationRefs(filenames FileNames, pattern string, excludes []string) ([]string, error) {
	possibleDeploys := filenames.GetApplicationDeploymentRefs()
	applications := SearchForApplications(pattern, possibleDeploys)

	applications, err := filterExcludes(excludes, applications)
	if err != nil {
		return nil, err
	}

	return applications, nil
}

func (op JsonPatchOp) Validate() error {
	if !strings.HasPrefix(op.Path, "/") {
		return ErrJsonPathPrefix
	}
	return nil
}

func (f *AuroraConfigFile) ToPrettyJson() string {

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

package auroraconfig

import (
	"regexp"
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

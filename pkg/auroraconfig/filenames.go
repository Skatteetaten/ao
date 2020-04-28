package auroraconfig

import (
	"path/filepath"
	"sort"
	"strings"

	"ao/pkg/collections"
	"github.com/pkg/errors"
)

type FileNames []string

func (f FileNames) GetApplicationDeploymentRefs() []string {
	var filteredFiles []string
	for _, file := range f.WithoutExtension() {
		if strings.ContainsRune(file, '/') && !strings.Contains(file, "about") {
			filteredFiles = append(filteredFiles, file)
		}
	}
	sort.Strings(filteredFiles)
	return filteredFiles
}

func (f FileNames) GetApplications() []string {
	unique := collections.NewStringSet()
	for _, file := range f.WithoutExtension() {
		if !strings.ContainsRune(file, '/') && !strings.Contains(file, "about") {
			unique.Add(file)
		}
	}
	filteredFiles := unique.All()
	sort.Strings(filteredFiles)
	return filteredFiles
}

func (f FileNames) GetEnvironments() []string {
	unique := collections.NewStringSet()
	for _, file := range f {
		if strings.ContainsRune(file, '/') && !strings.Contains(file, "about") {
			split := strings.Split(file, "/")
			unique.Add(split[0])
		}
	}
	filteredFiles := unique.All()
	sort.Strings(filteredFiles)
	return filteredFiles
}

func (f FileNames) WithoutExtension() []string {
	var withoutExt []string
	for _, file := range f {
		withoutExt = append(withoutExt, strings.TrimSuffix(file, filepath.Ext(file)))
	}
	return withoutExt
}

func (f FileNames) Find(name string) (string, error) {
	for _, fileName := range f {
		fileNameWithoutExtension := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		if name == fileName || name == fileNameWithoutExtension {
			return fileName, nil
		}
	}
	return "", errors.Errorf("could not find %s in AuroraConfig", name)
}

package fuzzyargs

import "github.com/skatteetaten/ao/pkg/auroraconfig"

/*
Module to create a list of apps and envs based upon user parameters from the command line.
The parameters are mathced based upon the file and folder names in the AuroraConfig

Init()							- Reads the AuroraConfig
PopulateFuzzyEnvAppList()		- Parses the args array given


*/

import (
	"errors"
	"strings"

	"github.com/skatteetaten/ao/pkg/configuration"
)

type FuzzyArgs struct {
	configuration *configuration.ConfigurationClass
	appList       []string
	envList       []string
	fileList      []string
	legalAppList  []string
	legalEnvList  []string
	legalFileList []string
}

func (fuzzyArgs *FuzzyArgs) Init(configuration *configuration.ConfigurationClass) (err error) {
	fuzzyArgs.configuration = configuration
	err = fuzzyArgs.getLegalEnvAppFileList()
	if err != nil {
		return err
	}
	return
}

func (fuzzyArgs *FuzzyArgs) addLegalApp(app string) {
	for i := range fuzzyArgs.legalAppList {
		if fuzzyArgs.legalAppList[i] == app {
			return
		}
	}
	fuzzyArgs.legalAppList = append(fuzzyArgs.legalAppList, app)
	return
}

func (fuzzyArgs *FuzzyArgs) addLegalEnv(env string) {
	for i := range fuzzyArgs.legalEnvList {
		if fuzzyArgs.legalEnvList[i] == env {
			return
		}
	}
	fuzzyArgs.legalEnvList = append(fuzzyArgs.legalEnvList, env)
	return
}

func (fuzzyArgs *FuzzyArgs) getLegalEnvAppFileList() (err error) {

	auroraConfig, err := auroraconfig.GetAuroraConfig(fuzzyArgs.configuration)
	if err != nil {
		return err
	}
	for filename := range auroraConfig.Files {
		fuzzyArgs.legalFileList = append(fuzzyArgs.legalFileList, filename)
		if strings.Contains(filename, "/") {
			// We have a full path name
			parts := strings.Split(filename, "/")
			fuzzyArgs.addLegalEnv(parts[0])
			if !strings.Contains(parts[1], "about.json") {
				if strings.HasSuffix(parts[1], ".json") {
					fuzzyArgs.addLegalApp(strings.TrimSuffix(parts[1], ".json"))
				}

			}
		}
	}

	return
}

// Try to match an argument with an app, returns "" if none found
func (fuzzyArgs *FuzzyArgs) GetFuzzyApp(arg string) (app string, err error) {
	if strings.HasSuffix(arg, ".json") {
		arg = strings.TrimSuffix(arg, ".json")
	}
	// First check for exact match
	for i := range fuzzyArgs.legalAppList {
		if fuzzyArgs.legalAppList[i] == arg {
			return arg, nil
		}
	}
	// No exact match found, look for an app name that contains the string
	for i := range fuzzyArgs.legalAppList {
		if strings.Contains(fuzzyArgs.legalAppList[i], arg) {
			if app != "" {
				err = errors.New(arg + ": Not a unique application identifier, matching " + app + " and " + fuzzyArgs.legalAppList[i])
				return "", err
			}
			app = fuzzyArgs.legalAppList[i]
		}
	}
	return app, nil
}

// Try to match an argument with an env, returns "" if none found
func (fuzzyArgs *FuzzyArgs) GetFuzzyEnv(arg string) (env string, err error) {
	// First check for exact match
	for i := range fuzzyArgs.legalEnvList {
		if fuzzyArgs.legalEnvList[i] == arg {
			return arg, nil
		}
	}
	// No exact match found, look for an env name that contains the string
	for i := range fuzzyArgs.legalEnvList {
		if strings.Contains(fuzzyArgs.legalEnvList[i], arg) {
			if env != "" {
				err = errors.New(arg + ": Not a unique environment identifier, matching both " + env + " and " + fuzzyArgs.legalEnvList[i])
				return "", err
			}
			env = fuzzyArgs.legalEnvList[i]
		}
	}
	return env, nil
}

func (fuzzyArgs *FuzzyArgs) PopulateFuzzyEnvAppList(args []string) (err error) {

	for i := range args {
		var env string
		var app string

		if strings.Contains(args[i], "/") {
			parts := strings.Split(args[i], "/")
			env, err = fuzzyArgs.GetFuzzyEnv(parts[0])
			if err != nil {
				return err
			}
			app, err = fuzzyArgs.GetFuzzyApp(parts[1])
			if err != nil {
				return err
			}
		} else {
			env, err = fuzzyArgs.GetFuzzyEnv(args[i])
			if err != nil {
				return err
			}
			app, err = fuzzyArgs.GetFuzzyApp(args[i])
			if err != nil {
				return err
			}
			if env != "" && app != "" {
				err = errors.New(args[i] + ": Not a unique identifier, matching both environment " + env + " and application " + app)
				return err
			}
		}
		if env == "" && app == "" {
			// None found, return error
			err = errors.New(args[i] + ": not found")
			return err
		}
		if env != "" {
			fuzzyArgs.envList = append(fuzzyArgs.envList, env)
		}
		if app != "" {
			fuzzyArgs.appList = append(fuzzyArgs.appList, app)
		}

	}
	return
}

func (fuzzyArgs *FuzzyArgs) GetApps() (apps []string) {
	return fuzzyArgs.appList
}

func (fuzzyArgs *FuzzyArgs) GetEnvs() (envs []string) {
	return fuzzyArgs.envList
}

func (fuzzyArgs *FuzzyArgs) GetApp() (app string, err error) {
	if len(fuzzyArgs.appList) > 1 {
		err = errors.New("No unique application identified")
		return "", err
	}
	if len(fuzzyArgs.appList) > 0 {
		return fuzzyArgs.appList[0], nil
	}
	return "", nil
}

func (fuzzyArgs *FuzzyArgs) GetEnv() (env string, err error) {
	if len(fuzzyArgs.envList) > 1 {
		err = errors.New("No unique environment identified")
		return "", err
	}
	if len(fuzzyArgs.envList) > 0 {
		return fuzzyArgs.envList[0], nil
	}
	return "", nil
}

func (fuzzyArgs *FuzzyArgs) IsLegalFile(filename string) (legal bool) {
	for i := range fuzzyArgs.legalFileList {
		if fuzzyArgs.legalFileList[i] == filename {
			return true
		}
	}
	return false
}

// Func to get a filename if we have just an appname
// Returns an error if several files exists.
func (fuzzyArgs *FuzzyArgs) App2File(app string) (filename string, err error) {
	if !strings.HasSuffix(filename, ".json") {
		filename = filename + ".json"
	}
	var found bool = false
	for i := range fuzzyArgs.legalFileList {
		if strings.Contains(fuzzyArgs.legalFileList[i], app) {
			if found {
				err = errors.New("Non-unique file identifier")
				return "", err
			}
			found = true
			filename = fuzzyArgs.legalFileList[i]
		}
	}
	if found {
		return filename, nil
	}
	return "", nil
}

// Func to get a filename if we expect the user to uniquely identify a file
func (fuzzyArgs *FuzzyArgs) GetFile() (filename string, err error) {
	if fuzzyArgs.IsLegalFile(filename) {
		return filename, nil
	}
	env, err := fuzzyArgs.GetEnv()
	if err != nil {
		return "", err
	}
	app, err := fuzzyArgs.GetApp()
	if err != nil {
		return "", err
	}
	/*if env == "" {
		// We need to find a unique env for a file
		filename, err = fuzzyArgs.App2File(app)
		if err != nil {
			return "", err
		}
	}*/

	if env == "" {
		if app == "" {
			filename = "about.json"
		} else {
			filename = app + ".json"
		}
	} else {
		if strings.Contains(filename, "about") {
			filename = env + "/" + "about.json"
		} else {
			filename = env + "/" + app + ".json"
		}
	}
	if fuzzyArgs.IsLegalFile(filename) {
		return filename, nil
	}

	err = errors.New("No such file")
	return "", err

}

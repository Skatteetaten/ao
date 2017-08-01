package newappcmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/configuration"
	"github.com/skatteetaten/aoc/pkg/fileutil"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

const UsageString = "Usage: aoc new-app <appname>"
const AppnameNeeded = "Missing appname parameter "
const MissingArgumentFormat = "Missing %v"
const InteractiveNoFlags = "No specification flags allowed for interactive run"
const DeployNeedVersion = "Need to have a version for deployment type deploy"
const appExistsError = "Error: App exists"
const notYetImplemented = "Not yet implemented"
const IllegalFolder = "Illegal folder"
const FolderNotEmpty = "Folder not empty"
const IllegalJson = "Illegal JSON in yo file"

const generatorExecutable = "yo"
const yoName = "aurora-openshift"
const generatorFileName = ".yo-rc.json"
const generatorNotInstalled = "Aurora OpenShift generator not installed"

/*{
"generator-aurora-openshift": {
"promptValues": {
"packageName": "no.skatteetaten.aurora.demo",
"maintainer": "HaakonKlausen <hakon.klausen@skatteetaten.no>"
},
"packageName": "no.skatteetaten.aurora.demo",
"description": "",
"oracle": false,
"spock": true,
"maintainer": "HaakonKlausen <hakon.klausen@skatteetaten.no>",
"baseName": "foobar"
}
}*/

type GeneratorAuroraOpenshift struct {
	GeneratorAuroraOpenshift struct {
		PackageName string `json:"packageName,omitempty"`
		Description string `json:"description,omitempty"`
		Oracle      bool   `json:"oracle,omitempty"`
		Spock       bool   `json:"spock,omitempty"`
		Maintainer  string `json:"maintainer,omitempty"`
		BaseName    string `json:"baseName,omitempty"`
	} `json:"generator-aurora-openshift,omitempty"`
}

type AuroraConfigPayload struct {
	GroupId    string `json:"groupId,omitempty"`
	ArtifactId string `json:"artifactId,omitempty"`
	Name       string `json:"name,omitempty"`
	Version    string `json:"version,omitempty"`
	Replicas   string `json:"replicas,omitempty"`
	Flags      struct {
		Rolling bool `json:"rolling,omitempty"`
		Cert    bool `json:"cert,omitempty"`
	} `json:"flags,omitempty"`
	Route struct {
		Generate bool `json:"generate,omitempty"`
	} `json:"route,omitempty"`
	Type    string `json:"type,omitempty"`
	Cluster string `json:"cluster,omitempty"`
}

type NewappcmdClass struct {
	configuration configuration.ConfigurationClass
}

func (newappcmdClass *NewappcmdClass) getAffiliation() (affiliation string) {
	if newappcmdClass.configuration.GetOpenshiftConfig() != nil {
		affiliation = newappcmdClass.configuration.GetOpenshiftConfig().Affiliation
	}
	return
}

func readGeneratorValues(foldername string) (generatorValues GeneratorAuroraOpenshift, err error) {
	absoluteFolderPath, err := filepath.Abs(foldername)
	if err != nil {
		return generatorValues, err
	}

	absoluteFilePath := filepath.Join(absoluteFolderPath, generatorFileName)

	filecontent, err := ioutil.ReadFile(absoluteFilePath)
	if err != nil {
		return generatorValues, err
	}

	if !jsonutil.IsLegalJson(string(filecontent)) {
		return generatorValues, errors.New(IllegalJson)
		return generatorValues, err
	}

	err = json.Unmarshal(filecontent, &generatorValues)
	if err != nil {
		return generatorValues, err
	}

	return generatorValues, nil
}

func startAuroraOpenshiftGenerator(foldername string, appname string) (generatorValues GeneratorAuroraOpenshift, err error) {
	exepath, err := exec.LookPath(generatorExecutable)
	if err != nil {
		return generatorValues, err
	}
	if fileutil.IsLegalFileFolder(exepath) != fileutil.SpecIsFile {
		return generatorValues, errors.New(generatorNotInstalled)
	}

	var command exec.Cmd
	command.Path, err = filepath.Abs(exepath)
	if err != nil {
		return generatorValues, err
	}

	command.Args = make([]string, 3)
	command.Args[0] = generatorExecutable
	command.Args[1] = yoName
	command.Args[2] = appname

	command.Dir, err = filepath.Abs(foldername)
	fmt.Println("DEBUG: DIR=" + command.Dir)
	fmt.Println("DEBUG: Appname:" + appname)
	if err != nil {
		return generatorValues, err
	}

	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err = command.Run()
	if err != nil {
		return generatorValues, err
	}

	// Get output values
	generatorValues, err = readGeneratorValues(foldername)
	if err != nil {
		return generatorValues, err
	}

	return generatorValues, nil
}

func (newappcmdClass *NewappcmdClass) generateAuroraConfigFiles(appname string, packagename string, env string) (payload map[string]AuroraConfigPayload, err error) {
	payload = make(map[string]AuroraConfigPayload, 3)

	var envAbout AuroraConfigPayload
	var envAboutName = env + "/about.json"

	payload[envAboutName] = envAbout

	return
}

func (newappcmdClass *NewappcmdClass) generateEnvAbout(appname string, packagename string, env string) (payload AuroraConfigPayload, err error) {
	return
}

func (newappcmdClass *NewappcmdClass) generateApp(appname string, packagename string, env string) (payload AuroraConfigPayload, err error) {
	payload.GroupId = packagename
	payload.ArtifactId = appname
	payload.Name = appname
	payload.Version = "1"
	payload.Replicas = "1"
	payload.Flags.Rolling = true
	payload.Flags.Cert = true
	payload.Route.Generate = true

	return payload, nil
}

func (newappcmdClass *NewappcmdClass) generateEnvApp(appname string, packagename string, env string) (payload AuroraConfigPayload, err error) {
	payload.Type = "development"
	payload.Cluster = "utv"

	return payload, nil
}

func (newappcmdClass *NewappcmdClass) NewappCommand(args []string, artifactid string, cluster string, env string, groupid string, interactive string, outputFolder string, deployentType string, version string) (output string, err error) {

	err = validateNewappCommand(args, artifactid, cluster, env, groupid, interactive, outputFolder, deployentType, version)
	if err != nil {
		return "", err
	}

	if interactive != "" {
		generatorValues, err := startAuroraOpenshiftGenerator(interactive, args[0])
		if err != nil {
			return "", err
		}
		fmt.Println("DEBUG: Packagename: " + generatorValues.GeneratorAuroraOpenshift.PackageName)
		//var appname = args[0]
		//var artifactId = generatorValues.GeneratorAuroraOpenshift.PackageName
		//var affiliation = newappcmdClass.getAffiliation()
		//var env = viper.GetString("USER")

		//payload, err := newappcmdClass.generateAuroraConfigFiles(appname, artifactid, env)
	}

	return
}

func validateNewappCommand(args []string, artifactid string, cluster string, env string, groupid string, interactive string, outputFolder string, deploymentType string, version string) (err error) {
	// Check for interactive, then no other parameters should be given

	if len(args) > 1 {
		err = errors.New(UsageString)
		return err
	}

	if len(args) == 0 {
		err = errors.New(AppnameNeeded)
		return err
	}

	if interactive != "" {

		if artifactid != "" || cluster != "" || env != "" || groupid != "" || outputFolder != "" {
			err = errors.New(InteractiveNoFlags)
			return err
		}
		// Check for valid folder
		if fileutil.IsLegalFileFolder(interactive) != fileutil.SpecIsFolder {
			err = errors.New(IllegalFolder)
			return err
		}

		// Check for empty folder
		isempty, err := fileutil.IsFolderEmpty(interactive)
		if err != nil {
			return err
		}
		if !isempty {
			err = errors.New(FolderNotEmpty)
			return err
		}
	} else {
		err = errors.New(notYetImplemented)
		return err

		// Check that we have a version if type is deployment
		if deploymentType == "deploy" {
			if version == "" {
				err = errors.New(DeployNeedVersion)
				return err
			}
		}

		// Check that we have cluster
		if cluster == "" {
			err = errors.New(fmt.Sprintf(MissingArgumentFormat, "cluster"))
			return err
		}

		if env == "" {
			err = errors.New(fmt.Sprintf(MissingArgumentFormat, "env"))
			return err
		}

		if groupid == "" {
			err = errors.New(fmt.Sprintf(MissingArgumentFormat, "groupid"))
			return err
		}

		if deploymentType == "" {
			err = errors.New(fmt.Sprintf(MissingArgumentFormat, "deploymentType"))
			return err
		}
	}
	return
}

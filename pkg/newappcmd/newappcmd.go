package newappcmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"strconv"

	"os"

	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/cmdoptions"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/executil"
	"github.com/skatteetaten/ao/pkg/fileutil"
	"github.com/skatteetaten/ao/pkg/jsonutil"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

const rollingDeploymentStrategy = "rolling"
const UsageString = "Usage: new-app <appname>"
const AppnameNeeded = "Missing appname parameter "
const MissingArgumentFormat = "Missing %v"
const InteractiveNoFlags = "No specification flags allowed for interactive run"
const DeployNeedVersion = "Need to have a version for deployment type deploy"
const appExistsError = "Error: App exists"
const notYetImplemented = "Not yet implemented"
const IllegalFolder = "Illegal folder"
const FolderNotEmpty = "Folder not empty"
const IllegalJson = "Illegal JSON in yo file"
const noRootFile = "Affiliation does not contain any about.json at the root level, this is currently unsupported"

const generatorExecutable = "yo"
const yoName = "aurora-openshift"
const generatorFileName = ".yo-rc.json"
const generatorNotInstalled = "Aurora OpenShift generator not installed"
const deploymentTypeDevelopment = "development"

type GeneratorAuroraOpenshift struct {
	GeneratorAuroraOpenshift struct {
		PackageName string `json:"packageName,omitempty"`
		Description string `json:"description,omitempty"`
		Oracle      bool   `json:"oracle,omitempty"`
		DbName      string `json:"dbName,omitempty"`
		Spock       bool   `json:"spock,omitempty"`
		Maintainer  string `json:"maintainer,omitempty"`
		BaseName    string `json:"baseName,omitempty"`
		Namespace   string `json:"namespace,omitempty"`
		Affiliation string `json:"affiliation,omitempty"`
	} `json:"generator-aurora-openshift,omitempty"`
}

type Route struct {
	Host string `json:"host,omitempty"`
	Path string `json:"path,omitempty"`
}

type AuroraConfig map[string]AuroraConfigPayload

type AuroraConfigPayload struct {
	GroupId        string `json:"groupId,omitempty"`
	ArtifactId     string `json:"artifactId,omitempty"`
	Name           string `json:"name,omitempty"`
	Version        string `json:"version,omitempty"`
	Replicas       string `json:"replicas,omitempty"`
	Certificate    bool   `json:"certificate,omitempty"`
	DeployStrategy struct {
		Type string `json:"type,omitempty"`
	} `json:"deployStrategy,omitempty"`
	Route    bool              `json:"route,omitempty"`
	Type     string            `json:"type,omitempty"`
	Cluster  string            `json:"cluster,omitempty"`
	Database map[string]string `json:"database,omitempty"`
}

type NewappcmdClass struct {
	Configuration *configuration.ConfigurationClass
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
	}

	err = json.Unmarshal(filecontent, &generatorValues)
	if err != nil {
		return generatorValues, err
	}

	return generatorValues, nil
}

func startAuroraOpenshiftGenerator(foldername string, appname string) (generatorValues GeneratorAuroraOpenshift, err error) {

	// Run the generator
	err = executil.RunInteractively(generatorExecutable, foldername, yoName, appname)
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

func (newappcmd *NewappcmdClass) generateEnvAbout(env string) (payload AuroraConfigPayload, filename string) {
	filename = env + "/about.json"
	return payload, filename
}

func (newappcmd *NewappcmdClass) generateApp(appname string, groupid string) (payload AuroraConfigPayload, filename string) {
	filename = appname + ".json"
	payload.GroupId = groupid
	payload.ArtifactId = appname
	payload.Name = appname
	payload.Version = "1"
	payload.Replicas = "1"
	payload.DeployStrategy.Type = rollingDeploymentStrategy
	payload.Certificate = true
	payload.Route = true

	return payload, filename
}

func (newappcmd *NewappcmdClass) generateEnvApp(appname string, env string, deploymentType string, cluster string, dbName string) (payload AuroraConfigPayload, filename string) {
	filename = env + "/" + appname + ".json"
	payload.Type = deploymentType
	payload.Cluster = cluster
	if deploymentType == deploymentTypeDevelopment {
		payload.Version = "1.0-SNAPSHOT"
	}
	if dbName != "" {
		payload.Database = make(map[string]string)
		payload.Database[dbName] = "auto"
	}
	return payload, filename
}

func (newappcmd *NewappcmdClass) mergeIntoAuroraConfig(config serverapi.AuroraConfig, env string, appname string, groupid string, deploymentType string, cluster string, dbName string) (mergedConfig serverapi.AuroraConfig, err error) {

	// Check if root about.json exists, if not exit with error
	_, rootExist := config.Files["about.json"]
	if !rootExist {
		err = errors.New(noRootFile)
		return
	}

	// Check if Env/About exists, if not create
	envAbout, envAboutFilename := newappcmd.generateEnvAbout(env)
	_, envappExists := config.Files[envAboutFilename]
	if !envappExists {
		config.Files[envAboutFilename], err = json.Marshal(envAbout)
		if err != nil {
			return
		}
		if newappcmd.Configuration.PersistentOptions.Verbose {
			fmt.Println(envAboutFilename)
			fmt.Println(jsonutil.PrettyPrintJson(string(config.Files[envAboutFilename])))
		}
	}

	// Merge app
	app, appFilename := newappcmd.generateApp(appname, groupid)
	config.Files[appFilename], err = json.Marshal(app)
	if err != nil {
		return
	}
	if newappcmd.Configuration.PersistentOptions.Verbose {
		fmt.Println(appFilename)
		fmt.Println(jsonutil.PrettyPrintJson(string(config.Files[appFilename])))
	}

	// Merge env/app
	envapp, envappFilename := newappcmd.generateEnvApp(appname, env, deploymentType, cluster, dbName)
	config.Files[envappFilename], err = json.Marshal(envapp)
	if err != nil {
		return
	}
	if newappcmd.Configuration.PersistentOptions.Verbose {
		fmt.Println(envappFilename)
		fmt.Println(jsonutil.PrettyPrintJson(string(config.Files[envappFilename])))
	}
	return config, err
}

func (newappcmd *NewappcmdClass) executeDeploy(foldername string) (err error) {
	const deployScript = "openshift-deploy.sh"
	const shellCmd = "bash"
	deployCmd, err := filepath.Abs(filepath.Join(foldername, deployScript))
	if err != nil {
		return err
	}

	err = executil.RunInteractively(shellCmd, foldername, deployCmd)
	if err != nil {
		return err
	}

	return
}

func (newappcmd *NewappcmdClass) printGeneratorValues(gv *GeneratorAuroraOpenshift) (err error) {
	fmt.Println("Generating config for application " + gv.GeneratorAuroraOpenshift.Affiliation + "-" +
		gv.GeneratorAuroraOpenshift.Namespace + "/" + gv.GeneratorAuroraOpenshift.BaseName)
	fmt.Println("\tPackage name: " + gv.GeneratorAuroraOpenshift.PackageName)
	fmt.Println("\tOracle: " + strconv.FormatBool(gv.GeneratorAuroraOpenshift.Oracle))
	fmt.Println("\tSpock: " + strconv.FormatBool(gv.GeneratorAuroraOpenshift.Spock))
	return
}

func (newappcmd *NewappcmdClass) NewappCommand(args []string, artifactid string, cluster string, env string, groupid string, folder string, outputFolder string, deploymentType string, version string, generateApp bool, persistentOptions *cmdoptions.CommonCommandOptions, deploy bool) (output string, err error) {

	err = validateNewappCommand(args, artifactid, cluster, env, groupid, folder, outputFolder, deploymentType, version, generateApp)
	if err != nil {
		return "", err
	}

	// If cluster not specified, get the API cluster from the config
	if cluster == "" {
		cluster = newappcmd.Configuration.GetApiClusterName()
	}

	var appname = args[0]
	if artifactid == "" {
		artifactid = appname
	}

	var dbName string
	if generateApp {
		folder = filepath.Join(folder, appname)
		if fileutil.IsLegalFileFolder(folder) != fileutil.SpecIllegal {
			err = errors.New("Application folder " + folder + " exists.")
			return "", err
		}

		err = os.Mkdir(folder, os.FileMode(0755))
		if err != nil {
			return "", err
		}

		var generatorValues GeneratorAuroraOpenshift
		empty, err := fileutil.IsFolderEmpty(folder)
		if err != nil {
			return "", err
		}
		if !empty {
			err = errors.New(FolderNotEmpty)
			return "", err
		}
		generatorValues, err = startAuroraOpenshiftGenerator(folder, args[0])
		if err != nil {
			return "", err
		}

		groupid = generatorValues.GeneratorAuroraOpenshift.PackageName
		database := generatorValues.GeneratorAuroraOpenshift.Oracle
		if database {
			dbName = generatorValues.GeneratorAuroraOpenshift.DbName
		}
		if env == "" {
			env = generatorValues.GeneratorAuroraOpenshift.Namespace
		}
		newappcmd.printGeneratorValues(&generatorValues)
	}

	// Get current aurora config
	fmt.Println("Getting Auroraconfig...")
	auroraConfig, err := auroraconfig.GetAuroraConfig(newappcmd.Configuration)
	if err != nil {
		return "", err
	}

	// Merge new app into aurora config
	fmt.Println("Merging new application config...")
	mergedAuroraConfig, err := newappcmd.mergeIntoAuroraConfig(auroraConfig, env, appname, groupid, deploymentType, cluster, dbName)
	if err != nil {
		return "", err
	}

	// Update aurora config in boober
	fmt.Println("Updating AuroraConfig...")
	err = auroraconfig.PutAuroraConfig(mergedAuroraConfig, newappcmd.Configuration)
	if err != nil {
		return "", err
	}

	// Execute deploy if flagged
	if deploy {
		fmt.Println("Executing deploy...")
		err = newappcmd.executeDeploy(folder)
		if err != nil {
			return "", err
		}
	}

	return
}

func validateNewappCommand(args []string, artifactid string, cluster string, env string, groupid string, folder string, outputFolder string, deploymentType string, version string, generateApp bool) (err error) {
	// Check for interactive, then no other parameters should be given

	if len(args) > 1 {
		err = errors.New(UsageString)
		return err
	}

	if len(args) == 0 {
		err = errors.New(AppnameNeeded)
		return err
	}

	if generateApp {

		if artifactid != "" || groupid != "" || outputFolder != "" {
			err = errors.New(InteractiveNoFlags)
			return err
		}
		// Check for valid folder
		/*if fileutil.IsLegalFileFolder(folder) != fileutil.SpecIsFolder {
			err = errors.New(IllegalFolder)
			return err
		}*/

		// Check for empty folder
		/*isempty, err := fileutil.IsFolderEmpty(folder)
		if err != nil {
			return err
		}
		if !isempty {
			err = errors.New(FolderNotEmpty)
			return err
		}*/
	} else {

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
	if outputFolder != "" {
		err = errors.New(notYetImplemented)
		return err
	}
	return
}

package auroraconfig

import (
	"encoding/json"

	"github.com/skatteetaten/ao/pkg/serverapi"
)

type AuroraConfigField struct {
	Path   string          `json:"path"`
	Value  json.RawMessage `json:"value"`
	Source string          `json:"source"`
}

type Mount struct {
	Path       string            `json:"path"`
	Type       string            `json:"type"`
	MountName  string            `json:"mountName"`
	VolumeName string            `json:"volumeName"`
	Exist      string            `json:"exist"`
	Content    map[string]string `json:"content"`
}

type AuroraVolume struct {
	Secrets map[string]string `json:"secrets"`
	Config  map[string]string `json:"config"`
	Mounts  []Mount           `json:"mounts"`
}

type Route struct {
	Name        string            `json:"name"`
	Host        string            `json:"host"`
	Path        string            `json:"path"`
	Annotations map[string]string `json:"annotations"`
}

type AuroraRoute struct {
	Route []Route `json:"route"`
}

type AuroraBuild struct {
	BaseName        string `json:"baseName"`
	BaseVersion     string `json:"baseVersion"`
	BuilderName     string `json:"builderName"`
	BuilderVersion  string `json:"builderVersion"`
	TestGitUrl      string `json:"testGitUrl"`
	TestTag         string `json:"testTag"`
	TestJenkinsFile string `json:"testJenkinsFile"`
	ExtraTags       string `json:"extraTags"`
	GroupId         string `json:"groupId"`
	ArtifactId      string `json:"artifactId"`
	Version         string `json:"version"`
	OutputKind      string `json:"outputKind"`
	OutputName      string `json:"outputName"`
	Triggers        string `json:"triggers"`
	BuildSuffix     string `json:"buildSuffix"`
}

type Database struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type AuroraDeploymentConfigFlags struct {
	Cert    bool `json:"cert"`
	Debug   bool `json:"debug"`
	Alarm   bool `json:"alarm"`
	Rolling bool `json:"rolling"`
}

type AuroraDeploymentConfigResource struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

type AuroraDeploymentConfigResources struct {
	Memory AuroraDeploymentConfigResource `json:"memory"`
	Cpu    AuroraDeploymentConfigResource `json:"cpu"`
}

type WebSeal struct {
	Host  string `json:"host"`
	Roles string `json:"roles"`
}

type HttpEndpoint struct {
	Path string `json:"path"`
	Port int    `json:"port"`
}

type Probe struct {
	Path    string `json:"path"`
	Port    int    `json:"port"`
	Delay   int    `json:"delay"`
	Timeout int    `json:"timeout"`
}

type AuroraDeploy struct {
	ApplicationFile string                          `json:"applicationFile"`
	OverrideFiles   map[string]json.RawMessage      `json:"overrideFiles"`
	ReleaseTo       string                          `json:"releaseTo"`
	Flags           AuroraDeploymentConfigFlags     `json:"flags"`
	Resources       AuroraDeploymentConfigResources `json:"resources"`
	Replicas        int                             `json:"replicas"`
	GroupId         string                          `json:"groupId"`
	ArtifactId      string                          `json:"artifactId"`
	Version         string                          `json:"version"`
	SplunkIndex     string                          `json:"splunkIndex"`
	Database        []Database                      `json:"database"`
	CertificateCn   string                          `json:"certificateCn"`
	WebSeal         WebSeal                         `json:"webSeal"`
	Prometheus      HttpEndpoint                    `json:"prometheus"`
	ManagementPath  string                          `json:"managementPath"`
	ServiceAccount  string                          `json:"serviceAccount"`
	Liveness        Probe                           `json:"liveness"`
	Readiness       Probe                           `json:"readiness"`
	DockerImagePath string                          `json:"dockerImagePath"`
	DockerTag       string                          `json:"dockerTag"`
}

type AuroraTemplate struct {
	Parameters map[string]string `json:"parameters"`
	Template   string            `json:"template"`
}

type AuroraLocalTemplate struct {
	Parameters   map[string]string `json:"parameters"`
	TemplateJson json.RawMessage   `json:"templateJson"`
}

type Permission struct {
	Groups       []string          `json:"groups"`
	Users        []string          `json:"users"`
	Rolebindings map[string]string `json:"rolebindings"`
}

type Permissions struct {
	Admin Permission `json:"admin"`
	View  Permission `json:"view"`
}

type AuroraApplication struct {
	SchemaVersion string                       `json:"schemaVersion"`
	Affiliation   string                       `json:"affiliation"`
	Cluster       string                       `json:"cluster"`
	Type          string                       `json:"type"`
	Name          string                       `json:"name"`
	EnvName       string                       `json:"envName"`
	Permissions   Permissions                  `json:"permissions"`
	Fields        map[string]AuroraConfigField `json:"fields"`

	Volume        AuroraVolume        `json:"volume"`
	Route         AuroraRoute         `json:"route"`
	Build         AuroraBuild         `json:"build"`
	Deploy        AuroraDeploy        `json:"deploy"`
	Template      AuroraTemplate      `json:"template"`
	LocalTemplate AuroraLocalTemplate `json:"localTemplate"`
	Namespace     string              `json:"namespace"`
}

type AuroraApplicationResult struct {
	DeployId           string                        `json:"deployId"`
	AuroraApplication  AuroraApplication             `json:"auroraApplication"`
	OpenShiftResponses []serverapi.OpenShiftResponse `json:"openShiftResponses"`
	Success            bool                          `json:"success"`
	Tag                string                        `json:"tag"`
}

func Response2ApplicationResults(response serverapi.Response) (applicationResults []AuroraApplicationResult, err error) {

	applicationResults = make([]AuroraApplicationResult, len(response.Items))
	for i := range response.Items {
		err = json.Unmarshal(response.Items[i], &applicationResults[i])
		if err != nil {
			return nil, err
		}
	}

	return applicationResults, nil
}

func ReportApplicationResuts(applicationResults []AuroraApplicationResult) (output string) {
	var newLine = ""
	for _, applicationResult := range applicationResults {
		output += newLine + "Deploy id: " + applicationResult.DeployId + " in cluster " + applicationResult.AuroraApplication.Cluster + "\n"
		output += "\tApplication: " + applicationResult.AuroraApplication.Affiliation + "-" + applicationResult.AuroraApplication.EnvName + "/" + applicationResult.AuroraApplication.Name
		newLine = "\n"
	}
	return
}

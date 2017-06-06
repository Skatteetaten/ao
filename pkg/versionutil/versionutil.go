package versionutil

import "encoding/json"

type VersionStruct struct {
	MajorVersion string `json:"majorVersion"`
	MinorVersion string `json:"minorVersion"`
	BuildNumber  string `json:"buildNumber"`
	Version      string `json:"version"`
	BuildStamp   string `json:"buildStamp"`
	Branch       string `json:"branch"`
}

func Version2Text(majorVersion string, minorVersion string, buildnumber string, githash string, branch string, buildstamp string) (output string, err error) {
	var version string
	if buildnumber != "" {
		version = version + "." + buildnumber
	}
	output = "Aurora OC version " + majorVersion + "." + minorVersion + "." + buildnumber
	if githash != "" {
		output += "\nBranch: " + branch + " (" + githash + ")"
	}
	if buildstamp != "" {
		output += "\nBuild Time: " + buildstamp
	}
	return
}

func Version2Json(majorVersion string, minorVersion string, buildnumber string, githash string, branch string, buildstamp string) (output string, err error) {

	var versionStruct VersionStruct
	versionStruct.MajorVersion = majorVersion
	versionStruct.MinorVersion = minorVersion
	versionStruct.BuildNumber = buildnumber
	versionStruct.Version = majorVersion + "." + minorVersion + "." + buildnumber
	versionStruct.BuildStamp = buildstamp
	versionStruct.Branch = branch
	outputBytes, err := json.Marshal(versionStruct)
	if err != nil {
		return
	}
	output = string(outputBytes)
	return
}

func Json2Version(jsonString []byte) (versionStruct VersionStruct, err error) {

	err = json.Unmarshal(jsonString, &versionStruct)
	if err != nil {
		return
	}

	return
}

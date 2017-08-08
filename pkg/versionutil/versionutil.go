package versionutil

import (
	"encoding/json"
	"errors"
)

type VersionStruct struct {
	MajorVersion string `json:"majorVersion"`
	MinorVersion string `json:"minorVersion"`
	BuildNumber  string `json:"buildNumber"`
	Version      string `json:"version"`
	BuildStamp   string `json:"buildStamp"`
	Branch       string `json:"branch"`
	Githash      string `json:"githash"`
}

var majorVersion = ""
var minorVersion = ""
var buildnumber = ""
var buildstamp = ""
var branch = ""
var githash = ""
var version = ""

func (versionStruct *VersionStruct) Version2Text() (output string, err error) {

	if versionStruct.BuildNumber != "" {
		version = version + "." + versionStruct.BuildNumber
	}
	output = "Aurora Oc version " + versionStruct.Version //MajorVersion + "." + versionStruct.MinorVersion + "." + versionStruct.BuildNumber
	if versionStruct.Githash != "" {
		output += "\nBranch: " + versionStruct.Branch + " (" + versionStruct.Githash + ")"
	}
	if versionStruct.BuildStamp != "" {
		output += "\nBuild Time: " + versionStruct.BuildStamp
	}

	return
}

func (versionStruct *VersionStruct) Init() { // majorVersion string, minorVersion string, buildnumber string, buildstamp string, branch string, githash string
	versionStruct.MajorVersion = majorVersion
	versionStruct.MinorVersion = minorVersion
	versionStruct.BuildNumber = buildnumber
	versionStruct.Version = version //majorVersion + "." + minorVersion + "." + buildnumber
	versionStruct.BuildStamp = buildstamp
	versionStruct.Branch = branch
	versionStruct.Githash = githash
}

func (versionStruct *VersionStruct) Version2Json() (output string, err error) {

	outputBytes, err := json.Marshal(versionStruct)
	if err != nil {
		return
	}
	output = string(outputBytes)
	return
}

func (versionStruct *VersionStruct) Version2Filename() (output string, err error) {
	output = versionStruct.Version //MajorVersion + "." + versionStruct.MinorVersion + "." + versionStruct.BuildNumber
	if output == ".." {
		err = errors.New("No version injected")
		return
	}
	output = "aoc_" + output
	return
}

func (versionStruct *VersionStruct) Version2Branch() (output string, err error) {
	output = versionStruct.Branch
	if output == "" {
		err = errors.New("No branch injected")
		return
	}
	return
}

func Json2Version(jsonString []byte) (versionStruct VersionStruct, err error) {

	err = json.Unmarshal(jsonString, &versionStruct)
	if err != nil {
		return
	}

	return
}

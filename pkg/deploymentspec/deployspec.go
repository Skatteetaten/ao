package deploymentspec

import (
	"fmt"
	"strings"
)

// DeploymentSpec represented as an empty interface.
type DeploymentSpec map[string]interface{}

// Get returns value of specified field.
func (spec DeploymentSpec) Get(jsonPointer string) interface{} {
	return spec.get(jsonPointer+"/value", "-")
}

// GetString returns string value of specified field.
func (spec DeploymentSpec) GetString(name string) string {
	return fmt.Sprintf("%v", spec.Get(name))
}

// GetBool returns true if parameter has boolean value true or string value "true" (case insensitive), otherwise false.
func (spec DeploymentSpec) GetBool(name string) bool {
	return strings.EqualFold(spec.GetString(name), "true")
}

// HasValue returns true if parameter has a value other than the default empty value, otherwise false.
func (spec DeploymentSpec) HasValue(name string) bool {
	return !strings.EqualFold(spec.GetString(name), "-")
}

// Cluster returns value of the cluster field.
func (spec DeploymentSpec) Cluster() string {
	return spec.GetString("cluster")
}

// Environment returns value of the environment field.
func (spec DeploymentSpec) Environment() string {
	return spec.GetString("envName")
}

// Name returns value of the name field.
func (spec DeploymentSpec) Name() string {
	return spec.GetString("name")
}

// Version returns value of the version field.
func (spec DeploymentSpec) Version() string {
	return spec.GetString("version")
}

// NewDeploymentSpec creates a minimal deployment spec. The purpose of this
// method is to create placeholder deployment specs for testing and error handling.
func NewDeploymentSpec(name, env, cluster, version string) DeploymentSpec {
	deploymentSpec := make(DeploymentSpec)
	deploymentSpec["name"] = map[string]interface{}{"value": name}
	deploymentSpec["envName"] = map[string]interface{}{"value": env}
	deploymentSpec["cluster"] = map[string]interface{}{"value": cluster}
	deploymentSpec["version"] = map[string]interface{}{"value": version}
	deploymentSpec["applicationDeploymentRef"] = map[string]interface{}{"value": env + "/" + name}
	return deploymentSpec
}

func (spec DeploymentSpec) get(jsonPointer, defaultValue string) interface{} {
	pointers := strings.Fields(strings.Replace(jsonPointer, "/", " ", -1))
	current := spec
	for i, pointer := range pointers {
		isLast := i == len(pointers)-1
		if next, ok := current[pointer]; ok && isLast {
			return next
		} else if ok && !isLast {
			current = next.(map[string]interface{})
		} else {
			return defaultValue
		}
	}
	return defaultValue
}

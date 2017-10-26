package configuration

import (
	"testing"
)

func TestNewTestConfiguration(t *testing.T) {
	config := NewTestConfiguration()
	if !config.Testing {
		t.Errorf("Testing: Expected true, got false")
	}
}

func TestGetApiClusterIndex(t *testing.T) {
	config := NewTestConfiguration()
	apiClusterIndex := config.GetApiClusterIndex()
	if apiClusterIndex != 0 {
		t.Errorf("Did not expect legal API cluster index, got %v", apiClusterIndex)
	}
}

func TestGetApiClusterName(t *testing.T) {
	config := NewTestConfiguration()
	apiClusterName := config.GetApiClusterName()
	if apiClusterName != "" {
		t.Errorf("Did not expect legal API cluster name, got %v", apiClusterName)
	}
}

func TestGetAffiliation(t *testing.T) {
	config := NewTestConfiguration()
	affiliation := config.GetAffiliation()
	if affiliation != "" {
		t.Errorf("Did not expect legal affiliation, got %v", affiliation)
	}
}

func TestGetPersistentOptions(t *testing.T) {
	config := NewTestConfiguration()
	persistentOptions := config.GetPersistentOptions()
	if persistentOptions.ShowConfig {
		t.Errorf("Did not expect ShowConfig")
	}
}

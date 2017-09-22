package configuration

import (
	"strings"
	"testing"
)

func TestNewTestConfiguration(t *testing.T) {
	config := NewTestConfiguration()
	if !config.Testing {
		t.Errorf("Testing: Expected true, got false")
	}
}

func TestInit(t *testing.T) {
	config := NewTestConfiguration()
	config.Init()

	const expectedConfigLocation = "ao.json"
	if !strings.Contains(config.configLocation, expectedConfigLocation) {
		t.Errorf("Expected config location %v, got %v", expectedConfigLocation, config.configLocation)
	}
}

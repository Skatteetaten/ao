package kubernetes

import "testing"

func TestGetConfig(t *testing.T) {
	var kubeConfig *KubeConfig
	kubeConfig = new(KubeConfig)
	err := kubeConfig.GetConfig()
	if err == nil {
		t.Error(("Expected Error in TestGetConfig"))
		//t.Errorf("Error in TestGetConfig: %v", err.Error())
	}
}

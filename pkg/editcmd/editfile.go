package editcmd

import (
	"github.com/skatteetaten/ao/pkg/auroraconfig"
	"github.com/skatteetaten/ao/pkg/configuration"
	"github.com/skatteetaten/ao/pkg/serverapi"
)

func EditFile(filename string, auroraConfig *serverapi.AuroraConfig, config *configuration.ConfigurationClass) (string, error) {

	content, version, err := auroraconfig.GetContent(filename, auroraConfig)
	if err != nil {
		return "", err
	}

	onSave := func(modified string) ([]string, error) {
		_, messages, err := auroraconfig.PutFile(filename, modified, version, config)
		return messages, err
	}

	_, output, err := editCycle(content, filename, config.PersistentOptions.Debug, onSave)
	if err != nil {
		return "", err
	}

	return output, nil
}

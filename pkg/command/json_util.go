package command

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

func IsLegalJson(jsonString string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(jsonString), &js) == nil
}

func parseOverride(override []string) (returnMap map[string]json.RawMessage, err error) {
	returnMap = make(map[string]json.RawMessage)

	for i := 0; i < len(override); i++ {
		indexByte := strings.IndexByte(override[i], ':')
		filename := override[i][:indexByte]

		jsonOverride := override[i][indexByte+1:]
		if !IsLegalJson(jsonOverride) {
			msg := fmt.Sprintf("%s is not a valid json", jsonOverride)
			return nil, errors.New(msg)
		}
		returnMap[filename] = json.RawMessage(jsonOverride)
	}
	return returnMap, err
}

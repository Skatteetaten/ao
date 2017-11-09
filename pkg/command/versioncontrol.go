package command

import (
	"fmt"
	"github.com/skatteetaten/ao/pkg/client"
	"strings"
)

func GetGitUrl(affiliation, user string, api *client.ApiClient) string {
	clientConfig, err := api.GetClientConfig()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	gitUrlPattern := clientConfig.GitUrlPattern

	if !strings.Contains(gitUrlPattern, "https://") {
		return fmt.Sprintf(gitUrlPattern, affiliation)
	}

	host := strings.TrimPrefix(gitUrlPattern, "https://")
	newPattern := fmt.Sprintf("https://%s@%s", user, host)
	return fmt.Sprintf(newPattern, affiliation)
}

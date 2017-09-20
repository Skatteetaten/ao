package serverapi

import (
	"net/http"
	"testing"
)

func TestCallApiInstance(t *testing.T) {
	const illegalUrl string = "https://westeros.skatteetaten.no/serverapi"
	var headers map[string]string

	_, err := callApiInstance(headers, http.MethodPut, "{\"Game\": \"Thrones\"}", false, illegalUrl, "token", false, false)
	if err == nil {
		t.Error("Did not detect illegal URL")
	}
}

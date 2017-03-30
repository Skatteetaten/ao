package serverapi

import "testing"

func TestGetApiAddress(t *testing.T) {
	booberAddress := GetApiAddress("foobar", true)
	if booberAddress == "" {
		t.Error("Boober Address for localhost blank")
	} else {
		if booberAddress != "http://localhost:8080" {
			t.Error("Illegal locahost address")
		}
	}

}

func TestCallApiInstance(t *testing.T) {
	const illegalUrl string = "https://westeros.skatteetaten.no/serverapi"

	_, err := callApiInstance("{\"Game\": \"Thrones\"}", false, false, false, illegalUrl, "", false)
	if err == nil {
		t.Error("Did not detect illegal URL")
	}
}

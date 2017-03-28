package serverapi

import "testing"

func TestGetBooberAddress(t *testing.T) {
	booberAddress := GetBooberAddress("foobar", true)
	if booberAddress == "" {
		t.Error("Boober Address for localhost blank")
	} else {
		if booberAddress != "http://localhost:8080" {
			t.Error("Illegal locahost address")
		}
	}

}

func TestCallBooberInstance(t *testing.T) {
	const illegalUrl string = "https://westeros.skatteetaten.no/serverapi"

	_, err := callBooberInstance("{\"Game\": \"Thrones\"}", false, false, false, illegalUrl, "")
	if err == nil {
		t.Error("Did not detect illegal URL")
	}
}

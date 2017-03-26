package boober

import "testing"

func TestCallBooberInstance(t *testing.T) {
	const illegalUrl string = "https://westeros.skatteetaten.no/boober"

	_, err := callBooberInstance("{\"Game\": \"Thrones\"}", false, false, false, illegalUrl, "")
	if err == nil {
		t.Error("Did not detect illegal URL")
	}
}




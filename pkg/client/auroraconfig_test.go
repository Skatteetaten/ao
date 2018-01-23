package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AuroraConfigSuccessResponseHandler(t *testing.T, responseFile string) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		assert.Contains(t, req.URL.Path, affiliation)
		data := ReadTestFile(responseFile)
		writer.Write(data)
	}
}

func AuroraConfigFailedResponseHandler(t *testing.T, responseFile string, status int) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)

		assert.Contains(t, req.URL.Path, affiliation)

		data := ReadTestFile(responseFile)
		writer.Write(data)
	}
}

func TestApi_GetAuroraConfig(t *testing.T) {

	t.Run("Successfully get AuroraConfig", func(t *testing.T) {
		fileName := "auroraconfig_paas_success_response"
		ts := httptest.NewServer(AuroraConfigSuccessResponseHandler(t, fileName))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", affiliation)
		ac, errResponse := api.GetAuroraConfig()

		assert.Empty(t, errResponse)
		assert.Len(t, ac.Files, 4)
	})
}

func TestApiClient_GetFileNames(t *testing.T) {
	t.Run("Should get all filenames in AuroraConfig for a given affiliation", func(t *testing.T) {

		fileName := "filenames_paas_success_response"
		ts := httptest.NewServer(AuroraConfigSuccessResponseHandler(t, fileName))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", affiliation)
		fileNames, err := api.GetFileNames()

		assert.NoError(t, err)
		assert.Len(t, fileNames, 4)
	})
}

func TestApiClient_PutAuroraConfig(t *testing.T) {
	t.Run("Successfully validate and save AuroraConfig", func(t *testing.T) {
		fileName := "auroraconfig_paas_success_response"
		ts := httptest.NewServer(AuroraConfigSuccessResponseHandler(t, fileName))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", affiliation)

		data := ReadTestFile("auroraconfig_paas_success_validation_request")
		var ac AuroraConfig
		err := json.Unmarshal(data, &ac)
		if err != nil {
			t.Error(err)
		}
		errResponse, err := api.ValidateAuroraConfig(&ac, false)
		assert.NoError(t, err)
		assert.Empty(t, errResponse)
	})

	t.Run("Validation and save should fail when deploy type is illegal", func(t *testing.T) {
		fileName := "auroraconfig_paas_failed_validation_response"
		ts := httptest.NewServer(AuroraConfigFailedResponseHandler(t, fileName, http.StatusBadRequest))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", affiliation)
		data := ReadTestFile("auroraconfig_paas_fail_validation_request")
		var ac AuroraConfig
		err := json.Unmarshal(data, &ac)
		if err != nil {
			t.Error(err)
		}

		errResponse, err := api.ValidateAuroraConfig(&ac, false)
		if errResponse == nil {
			fmt.Println(errResponse)
			t.Error("Expected errResponse to not be nil")
			return
		}
		assert.NoError(t, err)
		assert.NotEmpty(t, errResponse)
		assert.Len(t, errResponse.IllegalFieldErrors, 2)
	})
}

func TestApiClient_GetAuroraConfigFile(t *testing.T) {
	t.Run("Should successfully get AuroraConfigFile", func(t *testing.T) {
		fileName := "auroraconfigfile_paas_success_response"
		ts := httptest.NewServer(AuroraConfigSuccessResponseHandler(t, fileName))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", "paas")
		// TODO: Test ETag
		file, _, err := api.GetAuroraConfigFile("about.json")
		if err != nil {
			t.Error("Should not get error when fetching AuroraConfigFile")
			return
		}

		assert.Equal(t, "about.json", file.Name)
		assert.NotEmpty(t, file.Contents)
	})

	t.Run("Should return error message when AuroraConfigFile return success false", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			response := `{"success": false}`
			w.Write([]byte(response))
		}))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", "")

		// TODO: Test ETag
		file, _, err := api.GetAuroraConfigFile("about.json")
		assert.Error(t, err)
		assert.EqualError(t, err, "Failed getting file about.json")
		assert.Empty(t, file)
	})
}

func TestApiClient_PatchAuroraConfigFile(t *testing.T) {
	t.Run("", func(t *testing.T) {
		fileName := "test/foo.json"

		res := &BooberResponse{
			Success: true,
			Message: "OK",
			Items:   json.RawMessage(`[{"name":"test/foo.json","contents":"{}","version":"abbcc"}]`),
			Count:   1,
		}

		getResponse, err := json.Marshal(&res)
		if err != nil {
			t.Error(err)
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)

			if req.Method == http.MethodGet {
				w.Write(getResponse)
			} else if req.Method == http.MethodPatch {
				w.Write([]byte(`{"success": true}`))
			}
		}))
		defer ts.Close()

		api := NewApiClient(ts.URL, "", affiliation)

		op := JsonPatchOp{
			OP:    "add",
			Path:  "/version",
			Value: "develop-SNAPSHOT",
		}

		errRes, err := api.PatchAuroraConfigFile(fileName, op)
		if err != nil {
			t.Error(err)
		}

		assert.Empty(t, errRes)
	})
}

func TestJsonPatchOp_Validate(t *testing.T) {
	cases := []struct {
		JsonPath string
		Expected error
	}{
		{"/version", nil},
		{"version", ErrJsonPathPrefix},
	}

	for _, tc := range cases {
		op := JsonPatchOp{
			Path: tc.JsonPath,
		}

		err := op.Validate()
		if err != nil {
			assert.Error(t, err, tc.Expected.Error())
		}
	}
}

func TestFileNames_Filter(t *testing.T) {
	fileNames := FileNames{"about.json", "boober.json", "test/about.json", "test/boober.json"}
	deployments := fileNames.GetApplicationIds()
	environments := fileNames.GetEnvironments()
	applications := fileNames.GetApplications()

	assert.Len(t, deployments, 1)
	assert.Len(t, environments, 1)
	assert.Len(t, applications, 1)
	assert.Equal(t, "test/boober", deployments[0])
	assert.Equal(t, "test", environments[0])
	assert.Equal(t, "boober", applications[0])
}

func TestAuroraConfigFile_ToPrettyJson(t *testing.T) {
	acf := &AuroraConfigFile{
		Name:     "about.json",
		Contents: `{"type":"development"}`,
	}

	expected := "{\n  \"type\": \"development\"\n}"
	assert.Equal(t, expected, acf.ToPrettyJson())
}

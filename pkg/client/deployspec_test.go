package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApiClient_GetAuroraDeploySpec(t *testing.T) {
	t.Run("Should get aurora deploy spec", func(t *testing.T) {
		fileName := "deployspec_response"
		responseBody := ReadTestFile(fileName)

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer ts.Close()

		api := NewApiClient(ts.URL, "test", affiliation)
		spec, err := api.GetAuroraDeploySpec([]string{"aotest/redis"}, true)
		assert.NoError(t, err)

		assert.Len(t, spec, 1)
	})
}

func TestApiClient_GetAuroraDeploySpecFormatted(t *testing.T) {
	t.Run("Should get formatted aurora deploy spec", func(t *testing.T) {
		fileName := "deployspec_formatted_response"
		responseBody := ReadTestFile(fileName)
		expected, err := ioutil.ReadFile("./test_files/deployspec_formatted.txt")
		if err != nil {
			panic(err)
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseBody)
		}))
		defer ts.Close()

		api := NewApiClient(ts.URL, "test", affiliation)
		spec, err := api.GetAuroraDeploySpecFormatted("aotest", "redis", true)
		assert.NoError(t, err)

		assert.Equal(t, string(expected), spec)
	})
}

func Test_buildDeploySpecQueries(t *testing.T) {
	var applications []string
	for i := 0; i < 200; i++ {
		app := fmt.Sprintf("test/reference-application%d", i)
		applications = append(applications, app)
	}

	tests := []struct {
		name         string
		applications []string
		defaults     bool
		want         int
	}{
		{
			name:         "Should create multiple requests when url contains more than 3500 char",
			applications: applications,
			defaults:     true,
			want:         3,
		},
		{
			name:         "Should create multiple requests without defaults when url contains more than 3500 char",
			applications: applications,
			defaults:     false,
			want:         3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildDeploySpecQueries(tt.applications, tt.defaults); !reflect.DeepEqual(len(got), tt.want) {
				t.Errorf("buildDeploySpecQueries() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuroraDeploySpec_Value(t *testing.T) {
	tests := []struct {
		name        string
		a           AuroraDeploySpec
		jsonPointer string
		want        interface{}
	}{
		{
			name: "Should get version from AuroraDeploySpec",
			a: AuroraDeploySpec{
				"version": map[string]interface{}{
					"value": "3",
				},
			},
			jsonPointer: "/version",
			want:        "3",
		},
		{
			name: "Should get version from AuroraDeploySpec",
			a: AuroraDeploySpec{
				"permissions": map[string]interface{}{
					"admin": map[string]interface{}{
						"value": "TEST_group",
					},
				},
			},
			jsonPointer: "/permissions/admin",
			want:        "TEST_group",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Value(tt.jsonPointer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuroraDeploySpec.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

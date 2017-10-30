package boober

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"bytes"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"github.com/pkg/errors"
)

type ApplicationId struct {
	Environment string `json:"environment"`
	Application string `json:"application"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

type Api struct {
	Host        string
	Token       string
	Affiliation string
}

func NewApi(host, token, affiliation string) *Api {
	return &Api{
		Host:        host,
		Token:       token,
		Affiliation: affiliation,
	}
}

func (api *Api) GetFileNames() ([]string, error) {
	endpoint := fmt.Sprintf("/affiliation/%s/auroraconfig/filenames", api.Affiliation)
	body, err := api.PerformRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return []string{}, err
	}

	var res struct {
		Response
		Items []string `json:"items"`
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return []string{}, err
	}

	return res.Items, nil
}

func (api *Api) Deploy(applicationIds []ApplicationId, overrides map[string]json.RawMessage) error {

	applyPayload := struct {
		ApplicationIds []ApplicationId            `json:"applicationIds"`
		Overrides      map[string]json.RawMessage `json:"overrides"`
		Deploy         bool                       `json:"deploy"`
	}{
		ApplicationIds: applicationIds,
		Overrides:      overrides,
		Deploy:         true,
	}

	payload, err := json.Marshal(applyPayload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("/affiliation/%s/apply", api.Affiliation)
	body, err := api.PerformRequest(http.MethodPut, endpoint, payload)
	if err != nil {
		return err
	}

	var response struct {
		Response
		Items []struct {
			DeployId string `json:"deployId"`
			ADS struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"auroraDeploymentSpec"`
			Success bool `json:"success"`
		} `json:"items"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	for _, item := range response.Items {
		if !item.Success {
			fmt.Printf("Deploy failed (%s).\n", item.DeployId)
			fmt.Printf("Tried to deploy: %s/%s", item.ADS.Namespace, item.ADS.Name)
		}
		fmt.Printf("Deploy was success (%s).\n", item.DeployId)
		fmt.Printf("Deployed: %s/%s\n", item.ADS.Namespace, item.ADS.Name)
	}

	return nil
}

func (api *Api) PerformRequest(method string, endpoint string, payload []byte) ([]byte, error) {

	url := api.Host + endpoint
	reqLog := log.WithFields(log.Fields{
		"method": method,
		"url":    url,
	})

	reqLog.Info("Request")
	if len(payload) == 0 {
		reqLog.Debug("No payload")
	} else {
		log.Debug("Payload:\n", IndentJson(payload))
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return []byte{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+api.Token)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return []byte{}, errors.Wrap(err, "Error connecting to Boober")
	}

	reqLog = reqLog.WithFields(log.Fields{
		"status": res.StatusCode,
	})

	if res.StatusCode > 399 {
		reqLog.Error("Response")
		return []byte{}, errors.New("Boober request returned with error code " + res.Status)
	}

	reqLog.Info("Response")

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	log.Debug("Body:\n", IndentJson(body))

	return body, err
}

func IndentJson(data []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		log.Error("Failed to indent json ", err.Error())
		return string(data)
	}

	return string(out.Bytes())
}

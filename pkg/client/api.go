package client

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type ApiClient struct {
	Host        string
	Token       string
	Affiliation string
}

func NewApiClient(host, token, affiliation string) *ApiClient {
	return &ApiClient{
		Host:        host,
		Token:       token,
		Affiliation: affiliation,
	}
}

func (api *ApiClient) Do(method string, endpoint string, payload []byte) (*Response, error) {

	url := api.Host + endpoint
	logrus.WithFields(logrus.Fields{
		"method": method,
		"url":    url,
	}).Info("Request")

	if len(payload) == 0 {
		logrus.Debug("No payload")
	} else {
		logrus.Debug("Payload", string(payload))
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+api.Token)

	client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Error connecting to api")
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	errorCodes := []int{http.StatusNotFound, http.StatusForbidden}
	for _, c := range errorCodes {
		if res.StatusCode == c {
			return nil, errors.New(string(body))
		}
	}

	// TODO: Check content head for text/html
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"status":  res.StatusCode,
		"url":     api.Host + endpoint,
		"success": response.Success,
		"message": response.Message,
		"count":   response.Count,
	}).Info("Response")
	logrus.Debug("ResponseBody", string(body))

	return &response, nil
}

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/ao/pkg/config"
	"io/ioutil"
	"net/http"
)

const BooberApiVersion = "/v1"

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

	url := api.Host + BooberApiVersion + endpoint
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

	userAgentHeader := fmt.Sprintf("Go-http-client/1.1 ao/%s", config.Version)
	req.Header.Set("User-Agent", userAgentHeader)
	req.Header.Set("Accept", "application/json")
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

	var fields logrus.Fields
	err = json.Unmarshal(body, &fields)
	if err != nil {
		return nil, err
	}

	if res.StatusCode > 399 {
		logrus.WithFields(fields).Error("Request Error")
	}

	switch res.StatusCode {
	case http.StatusNotFound:
		return nil, errors.Errorf("Resource %s not found", endpoint)
	case http.StatusForbidden:
		return nil, errors.New("Token has expired. Please login: ao login <affiliation>")
	case http.StatusInternalServerError:
		logrus.WithFields(fields).Fatal("Unexpected error")
		return nil, errors.Errorf("Unexpected error from %s", url)
	case http.StatusServiceUnavailable:
		return nil, errors.Errorf("Service unavailable %s", api.Host)
	}

	var response Response
	if len(body) > 0 {
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
	}

	logrus.WithFields(logrus.Fields{
		"status":  res.StatusCode,
		"url":     url,
		"success": response.Success,
		"message": response.Message,
		"count":   response.Count,
	}).Info("Response")

	logrus.WithFields(fields).Debug("ResponseBody")

	return &response, nil
}
